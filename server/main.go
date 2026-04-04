package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	_ "image/jpeg"
	"image/png"

	_ "golang.org/x/image/webp"

	"github.com/cduerm/stringpic/stringer"
)

const (
	keepJobs time.Duration = 10 * time.Second

	minLineCount  = 100
	maxLineCount  = 6000
	minPinCount   = 2
	maxPinCount   = 1000
	minEraseValue = 0.0
	maxEraseValue = 50.0
)

// Job represents the state of an image processing task
type Job struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"` // "pending", "processing", "completed", "failed"
	TextData    string      `json:"text_data,omitempty"`
	Image       []byte      `json:"-"`
	TargetImage image.Image `json:"-"`
	LineCount   int         `json:"-"`
	PinCount    int         `json:"-"`
	EraseValue  float64     `json:"-"`
}

// JobStore handles our in-memory state safely across multiple goroutines
type JobStore struct {
	sync.RWMutex
	jobs map[string]*Job
}

var store = &JobStore{
	jobs: make(map[string]*Job),
}

// generateID creates a simple random hex string to use as a Job ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// main starts the HTTP server and registers the API routes.
func main() {
	log.Default().SetFlags(log.Ldate | log.Lmicroseconds)

	mux := http.NewServeMux()

	// Serve static files from the embedded filesystem
	subFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	mux.HandleFunc("POST /api/jobs", handleCreateJob)
	mux.HandleFunc("GET /api/jobs/{id}", handleGetJobStatus)
	mux.HandleFunc("GET /api/jobs/{id}/image", handleGetJobImage)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// handleCreateJob parses the uploaded image, creates a new background job, and returns the job ID.
func handleCreateJob(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Job")
	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	lineCount, _ := strconv.Atoi(r.FormValue("lineCount"))
	pinCount, _ := strconv.Atoi(r.FormValue("pinCount"))
	eraseValue, _ := strconv.ParseFloat(r.FormValue("eraseValue"), 64)

	lineCount = max(minLineCount, min(maxLineCount, lineCount))
	pinCount = max(minPinCount, min(maxPinCount, pinCount))
	eraseValue = max(minEraseValue, min(maxEraseValue, eraseValue))

	jobID := generateID()
	job := &Job{
		ID:          jobID,
		Status:      "processing",
		TargetImage: img,
		LineCount:   lineCount,
		PinCount:    pinCount,
		EraseValue:  eraseValue,
	}

	store.Lock()
	store.jobs[jobID] = job
	store.Unlock()
	log.Printf("Job %s created", jobID)

	go processImage(jobID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

// handleGetJobStatus returns the current status and metadata of a specific job.
func handleGetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	store.RLock()
	job, exists := store.jobs[jobID]
	store.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// handleGetJobImage serves the resulting image of a completed job.
func handleGetJobImage(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	store.RLock()
	job, exists := store.jobs[jobID]
	store.RUnlock()

	if !exists || job.Status != "completed" {
		http.Error(w, "Image not found or not ready", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(job.Image)
}

// processImage performs the long-running image generation process in the background.
func processImage(jobID string) {
	store.RLock()
	job, exists := store.jobs[jobID]
	store.RUnlock()

	if !exists {
		log.Printf("Job %s not found", jobID)
		return
	}
	startTime := time.Now()

	var opts []stringer.Option
	if job.LineCount > 0 {
		opts = append(opts, stringer.WithLinesCount(job.LineCount))
	}
	if job.PinCount > 0 {
		opts = append(opts, stringer.WithPinCount(job.PinCount))
	}
	if job.EraseValue >= 0 {
		opts = append(opts, stringer.WithEraseValue(job.EraseValue))
	}

	result, err := stringer.Generate(job.TargetImage, opts...)
	if err != nil {
		job.Status = "failed"
		job.TextData = err.Error()
		return
	}
	log.Printf("Job %s completed. Took %v. Deleting in %v...", jobID, time.Since(startTime), keepJobs)

	imageBytes := new(bytes.Buffer)
	png.Encode(imageBytes, result.Image)
	instructions := fmt.Sprintf("Image requires %.0f meters of string (22 cm diameter frame)", result.StringLength)

	store.Lock()
	if job, exists := store.jobs[jobID]; exists {
		job.Status = "completed"
		job.TextData = instructions
		job.Image = imageBytes.Bytes()
	}
	store.Unlock()

	time.Sleep(keepJobs)

	store.Lock()
	delete(store.jobs, jobID)
	store.Unlock()
	log.Printf("Job %s deleted", jobID)
}
