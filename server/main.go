package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"image"
	"log"
	"net/http"
	"sync"

	_ "image/jpeg"
	"image/png"

	"github.com/cduerm/stringpic/stringer"
)

// Job represents the state of an image processing task
type Job struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"` // "pending", "processing", "completed", "failed"
	TextData    string      `json:"text_data,omitempty"`
	Image       []byte      `json:"-"` // Unexported from JSON so we don't dump raw bytes
	TargetImage image.Image `json:"-"` // Unexported from JSON so we don't dump raw bytes
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

func main() {
	log.Default().SetFlags(log.Ldate | log.Lmicroseconds)

	mux := http.NewServeMux()

	// It tells Go to serve the index.html file at the root route.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend.html")
	})

	// Note: The "METHOD /path" syntax requires Go 1.22 or higher.
	mux.HandleFunc("POST /api/jobs", handleCreateJob)
	mux.HandleFunc("GET /api/jobs/{id}", handleGetJobStatus)
	mux.HandleFunc("GET /api/jobs/{id}/image", handleGetJobImage)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleCreateJob(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Job")
	// 1. Parse your multipart form (the uploaded image and parameters) here...
	// For brevity, we are skipping the parsing logic.
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

	firstBytes := make([]byte, 512)
	_, err = file.Read(firstBytes)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
	}
	log.Printf("File type: %s\n", http.DetectContentType(firstBytes))

	//create an image.Image from the file
	file.Seek(0, 0)
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	// 2. Create a new job
	jobID := generateID()
	job := &Job{
		ID:          jobID,
		Status:      "processing",
		TargetImage: img,
	}

	// 3. Save to in-memory store
	store.Lock()
	store.jobs[jobID] = job
	store.Unlock()
	log.Printf("Job %s created", jobID)

	// 4. Kick off the heavy processing in a background goroutine!
	go processImage(jobID)

	// 5. Immediately return the job ID to the user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted
	json.NewEncoder(w).Encode(job)
}

func handleGetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id") // Go 1.22 feature

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

func handleGetJobImage(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	store.RLock()
	job, exists := store.jobs[jobID]
	store.RUnlock()

	if !exists || job.Status != "completed" {
		http.Error(w, "Image not found or not ready", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png") // Change depending on your output type
	w.Write(job.Image)
}

// processImage simulates your long-running application logic
func processImage(jobID string) {
	job, exists := store.jobs[jobID]
	if !exists {
		log.Printf("Job %s not found", jobID)
		return
	}

	result, err := stringer.Generate(job.TargetImage)
	if err != nil {
		job.Status = "failed"
		job.TextData = err.Error()
		return
	}
	log.Printf("Job %s completed", jobID)

	// Update the job with the results
	store.Lock()
	defer store.Unlock()
	instructions := stringer.InstructionsText(result.Instructions[:10], result.StringLength)
	imageBytes := new(bytes.Buffer)
	png.Encode(imageBytes, result.EndImage)

	if job, exists := store.jobs[jobID]; exists {
		job.Status = "completed"
		job.TextData = instructions
		job.Image = imageBytes.Bytes()
	}
}
