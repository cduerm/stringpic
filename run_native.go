//go:build ignore

package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	log.Println("Building native desktop application...")
	cmdBuild := exec.Command("go", "build", "-o", "stringpic_app", "main.go")
	cmdBuild.Stdout = os.Stdout
	cmdBuild.Stderr = os.Stderr
	err := cmdBuild.Run()
	if err != nil {
		log.Fatalf("Failed to build native binary: %v\nNote: If you are on Linux, make sure you have the required graphics libraries installed (e.g. libxkbcommon-dev, libx11-xcb-dev, libegl1-mesa-dev, libwayland-dev, libx11-dev).\n", err)
	}

	log.Println("Running native desktop application...")
	cmdRun := exec.Command("./stringpic_app")
	cmdRun.Stdout = os.Stdout
	cmdRun.Stderr = os.Stderr
	err = cmdRun.Run()
	if err != nil {
		log.Fatalf("Failed to run native binary: %v\n", err)
	}
}
