package detect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	// Name returns the language or project type.
	Name() string
	// Detect returns true if this detector applies to the current project.
	Detect() bool
	// GetEntrypoint returns the run command for this project.
	GetEntrypoint() string
}

// -----------------------------------------------------
// Go Detector Implementation
// -----------------------------------------------------

// goDetector implements Detector for Go projects.
type goDetector struct{}

func (d goDetector) Name() string {
	return "Go"
}

func (d goDetector) Detect() bool {
	_, err := os.Stat("go.mod")
	return err == nil
}

func (d goDetector) GetEntrypoint() string {
	log.Println("Getting Go entrypoint...")

	// First try some common entrypoint patterns explicitly
	commonEntrypoints := []string{
		"./cmd/server/main.go",
		"./cmd/main/main.go",
	}

	for _, entry := range commonEntrypoints {
		if _, err := os.Stat(entry); err == nil {
			log.Printf("Found common entrypoint: %s", entry)
			return "go run " + entry
		}
	}

	// If common patterns don't work, use the recursive search
	entry, err := findGoEntrypoint(".")
	if err == nil {
		// Make sure the path starts with ./ for Go run
		if !strings.HasPrefix(entry, "./") && !strings.HasPrefix(entry, "/") {
			entry = "./" + entry
		}
		log.Printf("Using found entrypoint: %s", entry)
		return "go run " + entry
	}

	// If all else fails, try one last approach for common Go project structures
	log.Println("Trying module-based execution...")
	if modName, err := getGoModuleName(); err == nil && modName != "" {
		for _, dir := range []string{"cmd/server", "cmd/main", "cmd"} {
			if _, err := os.Stat(dir); err == nil {
				log.Printf("Trying module-based command: go run %s/%s", modName, dir)
				return fmt.Sprintf("go run %s/%s", modName, dir)
			}
		}
	}

	log.Println("No Go entrypoint found")
	return "echo 'No Go entrypoint found. Try running the application manually.'"
}

// getGoModuleName extracts the module name from go.mod
func getGoModuleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	// Simple regex-free parsing for "module name"
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	return "", fmt.Errorf("no module declaration found in go.mod")
}

// -----------------------------------------------------
// JavaScript Detector Implementation
// -----------------------------------------------------

// jsDetector implements Detector for JavaScript (Node.js) projects.
type jsDetector struct{}

func (d jsDetector) Name() string {
	return "JavaScript"
}

func (d jsDetector) Detect() bool {
	// Check for package.json first (Node.js projects)
	if _, err := os.Stat("package.json"); err == nil {
		return true
	}

	// Also detect projects with just .js files in the root
	jsFiles, _ := filepath.Glob("*.js")
	return len(jsFiles) > 0
}

func (d jsDetector) GetEntrypoint() string {
	log.Println("Finding JavaScript entrypoint...")

	// Check package.json for "main" field
	if _, err := os.Stat("package.json"); err == nil {
		data, err := ioutil.ReadFile("package.json")
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				// Check for main field
				if mainEntry, ok := pkg["main"].(string); ok && mainEntry != "" {
					log.Printf("Found main entry in package.json: %s", mainEntry)
					return "node " + mainEntry
				}

				// Check for start script
				if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
					if start, ok := scripts["start"].(string); ok && start != "" {
						log.Printf("Found start script in package.json: %s", start)
						return "npm start"
					}
				}

				// Check for dev script as fallback
				if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
					if dev, ok := scripts["dev"].(string); ok && dev != "" {
						log.Printf("Found dev script in package.json: %s", dev)
						return "npm run dev"
					}
				}
			}
		}
	}

	// Check for common entry files
	commonEntries := []string{
		"index.js",
		"app.js",
		"server.js",
		"main.js",
		"src/index.js",
		"src/app.js",
		"src/main.js",
	}

	for _, entry := range commonEntries {
		if _, err := os.Stat(entry); err == nil {
			log.Printf("Found JavaScript entry file: %s", entry)
			return "node " + entry
		}
	}

	// If we have a Next.js, React, Vue, or similar project
	if hasFile("next.config.js") || hasDir("pages") || hasDir("src/pages") {
		log.Println("Detected Next.js project")
		return "npm run dev"
	}

	if hasFile("vue.config.js") || hasDir("src/components") {
		log.Println("Detected Vue.js project")
		return "npm run serve"
	}

	if hasDir("src") && hasFile("src/App.js") {
		log.Println("Detected React project")
		return "npm start"
	}

	// Default fallback
	log.Println("Using default Node.js entrypoint")
	return "node ."
}

// -----------------------------------------------------
// Python Detector Implementation
// -----------------------------------------------------

// pythonDetector implements Detector for Python projects.
type pythonDetector struct{}

func (d pythonDetector) Name() string {
	return "Python"
}

func (d pythonDetector) Detect() bool {
	// Check for requirements.txt first (common in Python projects)
	if _, err := os.Stat("requirements.txt"); err == nil {
		return true
	}

	// Check for pyproject.toml (newer Python projects)
	if _, err := os.Stat("pyproject.toml"); err == nil {
		return true
	}

	// Check for setup.py (setuptools projects)
	if _, err := os.Stat("setup.py"); err == nil {
		return true
	}

	// Check for Pipfile (pipenv projects)
	if _, err := os.Stat("Pipfile"); err == nil {
		return true
	}

	// Also detect projects with just .py files in the root
	pyFiles, _ := filepath.Glob("*.py")
	return len(pyFiles) > 0
}

func (d pythonDetector) GetEntrypoint() string {
	log.Println("Finding Python entrypoint...")

	// Check for Django projects
	if hasFile("manage.py") {
		log.Println("Detected Django project")
		return "python manage.py runserver"
	}

	// Check for Flask projects
	if hasFile("app.py") && fileContains("app.py", "Flask") {
		log.Println("Detected Flask project")
		return "python app.py"
	}

	// Check for FastAPI projects
	if hasAnyFile([]string{"main.py", "app.py"}, "FastAPI") {
		log.Println("Detected FastAPI project")
		if hasFile("main.py") {
			return "uvicorn main:app --reload"
		}
		return "uvicorn app:app --reload"
	}

	// Check common entry points by name
	commonEntries := []string{
		"main.py",
		"app.py",
		"run.py",
		"server.py",
		"application.py",
	}

	for _, entry := range commonEntries {
		if _, err := os.Stat(entry); err == nil {
			log.Printf("Found Python entry file: %s", entry)
			return "python " + entry
		}
	}

	// Search for files with if __name__ == "__main__"
	if entry, err := findPythonEntrypoint("."); err == nil {
		log.Printf("Found Python script with __main__: %s", entry)
		return "python " + entry
	}

	// Default fallback - try to use a module approach
	pkgName := filepath.Base(currentDir())
	log.Printf("Using Python module approach for: %s", pkgName)
	return fmt.Sprintf("python -m %s", pkgName)
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
		// You can add more detectors here in the future.
	}

	// First log what we're looking for
	log.Println("Detecting project type...")

	// Try to detect each language
	for _, d := range detectors {
		if d.Detect() {
			lang := d.Name()
			cmd := d.GetEntrypoint()
			log.Printf("Detected %s project, command: %s", lang, cmd)
			return lang, cmd
		}
	}

	log.Println("No recognized project type detected")
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

// findGoEntrypoint recursively searches for Go entrypoints in the project
// It prioritizes main.go files with "package main" declarations
func findGoEntrypoint(root string) (string, error) {
	log.Println("Searching for Go entrypoint in project...")

	// First, create a list to collect all potential entrypoints
	var entrypoints []string

	// A counter for logging purposes
	foundCount := 0

	// Walk the directory tree to find all potential main.go files
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden and vendor directories
		if d.IsDir() && (strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor") {
			return filepath.SkipDir
		}

		// Check for main.go files
		if !d.IsDir() && d.Name() == "main.go" {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil // Continue searching even if we can't read this file
			}

			if bytes.Contains(data, []byte("package main")) {
				log.Printf("Found potential entrypoint: %s", path)
				entrypoints = append(entrypoints, path)
				foundCount++
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking directory tree: %v", err)
	}

	if len(entrypoints) == 0 {
		return "", fmt.Errorf("no Go entrypoint found")
	}

	log.Printf("Found %d potential Go entrypoints", foundCount)

	// Prioritize entrypoints based on common patterns
	// First, check for entrypoints in cmd/*/main.go pattern (common in Go projects)
	for _, entry := range entrypoints {
		if strings.Contains(entry, "/cmd/") && strings.HasSuffix(entry, "/main.go") {
			log.Printf("Selected entrypoint (cmd pattern): %s", entry)
			return entry, nil
		}
	}

	// If no cmd/ entrypoint found, just use the first one
	log.Printf("Selected entrypoint (first found): %s", entrypoints[0])
	return entrypoints[0], nil
}

// findPythonEntrypoint searches recursively from root for a .py file containing a __main__ block.
func findPythonEntrypoint(root string) (string, error) {
	var entrypoint string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories and common virtual environment directories.
		if d.IsDir() && (strings.HasPrefix(d.Name(), ".") || d.Name() == "venv" || d.Name() == "__pycache__") {
			return filepath.SkipDir
		}
		// Only check .py files.
		if !d.IsDir() && filepath.Ext(d.Name()) == ".py" {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}
			if bytes.Contains(data, []byte("if __name__ == \"__main__\"")) ||
				bytes.Contains(data, []byte("if __name__ == '__main__'")) {
				entrypoint = path
				return errFound // Break early.
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

// -----------------------------------------------------
// Additional Helper Functions
// -----------------------------------------------------

// hasFile checks if a file exists
func hasFile(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// hasDir checks if a directory exists
func hasDir(dirname string) bool {
	info, err := os.Stat(dirname)
	return err == nil && info.IsDir()
}

// fileContains checks if a file contains a specific string
func fileContains(filename, searchString string) bool {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return false
	}
	return bytes.Contains(data, []byte(searchString))
}

// hasAnyFile checks if any of the files contain the search string
func hasAnyFile(filenames []string, searchString string) bool {
	for _, filename := range filenames {
		if fileContains(filename, searchString) {
			return true
		}
	}
	return false
}

// currentDir gets the current directory name
func currentDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(pwd)
}
