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

// Service represents a runnable service with a name and a command.
type Service struct {
	Name    string
	Command string
}

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
// PHP Detector Implementation
// -----------------------------------------------------

type phpDetector struct{}

func (d phpDetector) Name() string {
	// We override the name later if Laravel Sail is detected.
	return "PHP"
}

func (d phpDetector) Detect() bool {
	_, err := os.Stat("composer.json")
	return err == nil
}

func (d phpDetector) GetEntrypoint() string {
	log.Println("Getting PHP entrypoint...")
	contents, err := os.ReadFile("composer.json")
	if err == nil {
		var data map[string]interface{}
		if err := json.Unmarshal(contents, &data); err == nil {
			if checkDependency(data, "require-dev", "laravel/sail") || checkDependency(data, "require", "laravel/sail") {
				log.Println("Detected Laravel Sail project")
				return "./vendor/bin/sail up"
			}
		}
	}

	commonEntrypoints := []string{
		"./public/index.php",
		"./index.php",
	}

	for _, entry := range commonEntrypoints {
		if _, err := os.Stat(entry); err == nil {
			log.Printf("Found common entrypoint: %s", entry)
			docRoot := filepath.Dir(entry)
			return fmt.Sprintf("php -S localhost:8000 -t %s", docRoot)
		}
	}

	phpFiles, err := findPhpFiles(".")
	if err == nil && len(phpFiles) > 0 {
		entry := phpFiles[0]
		log.Printf("Using found PHP file as entrypoint: %s", entry)
		docRoot := filepath.Dir(entry)
		return fmt.Sprintf("php -S localhost:8000 -t %s", docRoot)
	}

	log.Println("No PHP entrypoint found")
	return "echo 'No PHP entrypoint found. Try running the application manually.'"
}

func findPhpFiles(root string) ([]string, error) {
	var phpFiles []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".php") {
			phpFiles = append(phpFiles, path)
		}
		return nil
	})
	return phpFiles, err
}

func checkDependency(data map[string]interface{}, field, dependency string) bool {
	if deps, ok := data[field].(map[string]interface{}); ok {
		if _, exists := deps[dependency]; exists {
			return true
		}
	}
	return false
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
	log.Println("Getting Go entrypoint...")
	if _, err := os.Stat("air.toml"); err == nil {
		log.Println("Detected Air project")
		return "air"
	}

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

	entry, err := findGoEntrypoint(".")
	if err == nil {
		if !strings.HasPrefix(entry, "./") && !strings.HasPrefix(entry, "/") {
			entry = "./" + entry
		}
		log.Printf("Using found entrypoint: %s", entry)
		return "go run " + entry
	}

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

func getGoModuleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("no module declaration found in go.mod")
}

func findGoEntrypoint(root string) (string, error) {
	log.Println("Searching for Go entrypoint in project...")
	var entrypoints []string
	foundCount := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor") {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "main.go" {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
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
	for _, entry := range entrypoints {
		if strings.Contains(entry, "/cmd/") && strings.HasSuffix(entry, "/main.go") {
			log.Printf("Selected entrypoint (cmd pattern): %s", entry)
			return entry, nil
		}
	}
	log.Printf("Selected entrypoint (first found): %s", entrypoints[0])
	return entrypoints[0], nil
}

// -----------------------------------------------------
// JavaScript Detector Implementation
// -----------------------------------------------------

type jsDetector struct{}

func (d jsDetector) Name() string {
	return "JavaScript"
}

func (d jsDetector) Detect() bool {
	if _, err := os.Stat("package.json"); err == nil {
		return true
	}
	jsFiles, _ := filepath.Glob("*.js")
	return len(jsFiles) > 0
}

func (d jsDetector) GetEntrypoint() string {
	log.Println("Finding JavaScript entrypoint...")
	if _, err := os.Stat("package.json"); err == nil {
		data, err := ioutil.ReadFile("package.json")
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if mainEntry, ok := pkg["main"].(string); ok && mainEntry != "" {
					log.Printf("Found main entry in package.json: %s", mainEntry)
					return "node " + mainEntry
				}
				if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
					if start, ok := scripts["start"].(string); ok && start != "" {
						log.Printf("Found start script in package.json: %s", start)
						return "npm start"
					}
					if dev, ok := scripts["dev"].(string); ok && dev != "" {
						log.Printf("Found dev script in package.json: %s", dev)
						return "npm run dev"
					}
				}
			}
		}
	}

	commonEntries := []string{
		"index.js", "app.js", "server.js", "main.js",
		"src/index.js", "src/app.js", "src/main.js",
	}
	for _, entry := range commonEntries {
		if _, err := os.Stat(entry); err == nil {
			log.Printf("Found JavaScript entry file: %s", entry)
			return "node " + entry
		}
	}
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
	log.Println("Using default Node.js entrypoint")
	return "node ."
}

func hasFile(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func hasDir(dirname string) bool {
	info, err := os.Stat(dirname)
	return err == nil && info.IsDir()
}

// -----------------------------------------------------
// Python Detector Implementation
// -----------------------------------------------------

type pythonDetector struct{}

func (d pythonDetector) Name() string {
	return "Python"
}

func (d pythonDetector) Detect() bool {
	if _, err := os.Stat("requirements.txt"); err == nil {
		return true
	}
	if _, err := os.Stat("pyproject.toml"); err == nil {
		return true
	}
	if _, err := os.Stat("setup.py"); err == nil {
		return true
	}
	if _, err := os.Stat("Pipfile"); err == nil {
		return true
	}
	pyFiles, _ := filepath.Glob("*.py")
	return len(pyFiles) > 0
}

func (d pythonDetector) GetEntrypoint() string {
	log.Println("Finding Python entrypoint...")
	if hasFile("manage.py") {
		log.Println("Detected Django project")
		return "python manage.py runserver"
	}
	commonEntries := []string{
		"main.py", "app.py", "run.py", "server.py", "application.py",
	}
	for _, entry := range commonEntries {
		if _, err := os.Stat(entry); err == nil {
			log.Printf("Found Python entry file: %s", entry)
			return "python " + entry
		}
	}
	if entry, err := findPythonEntrypoint("."); err == nil {
		log.Printf("Found Python script with __main__: %s", entry)
		return "python " + entry
	}
	pkgName := filepath.Base(currentDir())
	log.Printf("Using Python module approach for: %s", pkgName)
	return fmt.Sprintf("python -m %s", pkgName)
}

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
				return errors.New("found")
			}
		}
		return nil
	})
	if err != nil && err.Error() != "found" {
		return "", err
	}
	if entrypoint == "" {
		return "", errors.New("no Python entrypoint found")
	}
	return entrypoint, nil
}

func currentDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(pwd)
}

// -----------------------------------------------------
// Unified Entrypoint Detection (Original)
// -----------------------------------------------------

func GetEntrypoint() (string, string) {
	detectors := []Detector{
		goDetector{},
		jsDetector{},
		pythonDetector{},
		phpDetector{},
	}
	log.Println("Detecting project type...")
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

func RunProject(command string) error {
	c := exec.Command("sh", "-c", command)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// -----------------------------------------------------
// New Function: GetServices
// -----------------------------------------------------

func GetServices() []Service {
	detectors := []Detector{
		goDetector{},
		jsDetector{},
		pythonDetector{},
		phpDetector{},
	}
	var services []Service
	log.Println("Detecting all available services...")
	for _, d := range detectors {
		if d.Detect() {
			cmd := d.GetEntrypoint()
			name := d.Name()
			if d.Name() == "PHP" && strings.Contains(cmd, "sail") {
				name = "Laravel Sail"
			}
			if d.Name() == "JavaScript" && strings.Contains(cmd, "run dev") {
				name = "NPM Dev Script"
			}
			service := Service{
				Name:    name,
				Command: cmd,
			}
			log.Printf("Detected service: %+v", service)
			services = append(services, service)
		}
	}
	if len(services) == 0 {
		lang, cmd := GetEntrypoint()
		if lang != "" && cmd != "" {
			service := Service{
				Name:    fmt.Sprintf("%s (default)", lang),
				Command: cmd,
			}
			services = append(services, service)
		}
	}
	return services
}
