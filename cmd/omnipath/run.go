package omnipath

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adammpkins/OmniPath/internal/detect"
	"github.com/spf13/cobra"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Detector defines the interface for project entrypoint detection.
type Detector interface {
	Name() string
	Detect() bool
	GetEntrypoint() string
}

// -----------------------------------------------------
// Go Detector Implementation
// -----------------------------------------------------

type goDetector struct{}

func (d goDetector) Name() string {
	return "Go"
}

func (d goDetector) Detect() bool {
	_, err := os.Stat("go.mod")
	return err == nil
}

func (d goDetector) GetEntrypoint() string {
	// First, check for a common entrypoint location: "cmd/main/main.go"
	if _, err := os.Stat("cmd/main/main.go"); err == nil {
		return "go run ./cmd/main/main.go"
	}
	// Otherwise, do a recursive search from the project root.
	if entry, err := findGoEntrypoint("."); err == nil {
		return "go run " + entry
	}
	// Fallback if nothing is found.
	return "go run ."
}

// -----------------------------------------------------
// JavaScript Detector Implementation
// -----------------------------------------------------

type jsDetector struct{}

func (d jsDetector) Name() string {
	return "JavaScript"
}

func (d jsDetector) Detect() bool {
	_, err := os.Stat("package.json")
	return err == nil
}

func (d jsDetector) GetEntrypoint() string {
	data, err := ioutil.ReadFile("package.json")
	if err == nil {
		var pkg map[string]interface{}
		if err := json.Unmarshal(data, &pkg); err == nil {
			if mainEntry, ok := pkg["main"].(string); ok && mainEntry != "" {
				return "node " + mainEntry
			}
		}
	}
	if _, err := os.Stat("index.js"); err == nil {
		return "node index.js"
	}
	return "node ."
}

// -----------------------------------------------------
// Python Detector Implementation
// -----------------------------------------------------

type pythonDetector struct{}

func (d pythonDetector) Name() string {
	return "Python"
}

func (d pythonDetector) Detect() bool {
	_, err := os.Stat("requirements.txt")
	return err == nil
}

func (d pythonDetector) GetEntrypoint() string {
	// Prefer main.py in the project root.
	if _, err := os.Stat("main.py"); err == nil {
		return "python main.py"
	}
	if entry, err := findPythonEntrypoint("."); err == nil {
		return "python " + entry
	}
	return "python ."
}

// -----------------------------------------------------
// Unified Entrypoint Detection
// -----------------------------------------------------

// GetEntrypoint iterates over all registered detectors and returns the first one
// that detects the project. It returns a tuple of (language, runCommand).
func GetEntrypoint() (string, string) {
	detectors := []Detector{
		goDetector{},
		jsDetector{},
		pythonDetector{},
		// Add more detectors here as needed.
	}
	for _, d := range detectors {
		if d.Detect() {
			return d.Name(), d.GetEntrypoint()
		}
	}
	return "", ""
}

// RunProject executes the given command using the shell.
func RunProject(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// -----------------------------------------------------
// Helper Functions for Go and Python Detection
// -----------------------------------------------------

var errFound = errors.New("found entrypoint")

// findGoEntrypoint recursively searches for a "main.go" that contains "package main" starting from root.
func findGoEntrypoint(root string) (string, error) {
	var entrypoint string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden and vendor directories.
		if d.IsDir() && (strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor") {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "main.go" {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}
			if bytes.Contains(data, []byte("package main")) {
				entrypoint = path
				return errFound // break early
			}
		}
		return nil
	})
	if err != nil && err != errFound {
		return "", err
	}
	if entrypoint == "" {
		return "", errors.New("no Go entrypoint found")
	}
	return entrypoint, nil
}

// findPythonEntrypoint recursively searches for a .py file containing a __main__ block.
func findPythonEntrypoint(root string) (string, error) {
	var entrypoint string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (strings.HasPrefix(d.Name(), ".") || d.Name() == "venv" || d.Name() == "__pycache__") {
			return filepath.SkipDir
		}
		if !d.IsDir() && filepath.Ext(d.Name()) == ".py" {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}
			if bytes.Contains(data, []byte("if __name__ == \"__main__\"")) ||
				bytes.Contains(data, []byte("if __name__ == '__main__'")) {
				entrypoint = path
				return errFound
			}
		}
		return nil
	})
	if err != nil && err != errFound {
		return "", err
	}
	if entrypoint == "" {
		return "", errors.New("no Python entrypoint found")
	}
	return entrypoint, nil
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Detects and runs the project based on its type",
	Run: func(cmd *cobra.Command, args []string) {
		lang, command := detect.GetEntrypoint()
		if lang == "" {
			log.Println("Unknown project type. No run command available.")
			return
		}
		fmt.Printf("Detected project type: %s\nRunning command: %s\n", lang, command)
		if err := detect.RunProject(command); err != nil {
			log.Fatalf("Error running project: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
