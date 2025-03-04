package detect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Service represents a runnable service with a name, command, and an interactive flag.
type Service struct {
	Name        string
	Command     string
	Interactive bool
}

// Detector defines the interface for project entrypoint detection.
// Now each detector returns a slice of Service values.
type Detector interface {
	Name() string
	Detect() bool
	GetServices() []Service
}

// --- Go Detector Implementation ---

type goDetector struct{}

func (d goDetector) Name() string {
	return "Go"
}

func (d goDetector) Detect() bool {
	_, err := os.Stat("go.mod")
	return err == nil
}

// getGoServices returns Go service options based on which files exist.
func getGoServices() []Service {
	var services []Service
	// If .air.toml exists, only present "Air" (interactive).
	if _, err := os.Stat(".air.toml"); err == nil {
		services = append(services, Service{
			Name:        "Air",
			Command:     "air",
			Interactive: true,
		})
		return services
	}
	// If ./cmd/server/main.go exists, offer that as interactive.
	if _, err := os.Stat("./cmd/server/main.go"); err == nil {
		services = append(services, Service{
			Name:        "Go Server",
			Command:     "go run ./cmd/server/main.go",
			Interactive: true,
		})
		return services
	}
	//if ./main.go exists, it's just a command.
	if _, err := os.Stat("./main.go"); err == nil {
		services = append(services, Service{
			Name:        "Go App",
			Command:     "go run ./main.go",
			Interactive: false,
		})
		return services
	}
	// Otherwise, if it's just a command (cmd/main/main.go or cmd/main.go), run it non-interactively.
	if _, err := os.Stat("./cmd/main/main.go"); err == nil {
		services = append(services, Service{
			Name:        "Go App",
			Command:     "go run ./cmd/main/main.go",
			Interactive: false,
		})
		return services
	}
	if _, err := os.Stat("./cmd/main.go"); err == nil {
		services = append(services, Service{
			Name:        "Go App",
			Command:     "go run ./cmd/main.go",
			Interactive: false,
		})
		return services
	}
	return services
}

func (d goDetector) GetServices() []Service {
	return getGoServices()
}

// --- PHP Detector Implementation ---

type phpDetector struct{}

func (d phpDetector) Name() string {
	return "PHP"
}

func (d phpDetector) Detect() bool {
	_, err := os.Stat("composer.json")
	return err == nil
}

func (d phpDetector) GetServices() []Service {
	log.Println("Getting PHP entrypoint...")
	contents, err := os.ReadFile("composer.json")
	if err == nil {
		var data map[string]interface{}
		if err := json.Unmarshal(contents, &data); err == nil {
			if checkDependency(data, "require-dev", "laravel/sail") || checkDependency(data, "require", "laravel/sail") {
				log.Println("Detected Laravel Sail project")
				return []Service{{
					Name:        "Laravel Sail",
					Command:     "DOCKER_STREAMS=1 DOCKER_PLAIN_OUTPUT=1 script -q /dev/null ./vendor/bin/sail up",
					Interactive: true,
				}}
			}
		}
	}
	// Fallback: a standard PHP server.
	commonEntrypoints := []string{
		"./public/index.php",
		"./index.php",
	}
	for _, entry := range commonEntrypoints {
		if _, err := os.Stat(entry); err == nil {
			docRoot := filepath.Dir(entry)
			return []Service{{
				Name:        "PHP",
				Command:     fmt.Sprintf("php -S localhost:8000 -t %s", docRoot),
				Interactive: true,
			}}
		}
	}
	return []Service{{
		Name:        "PHP (default)",
		Command:     "echo 'No PHP entrypoint found. Try running the application manually.'",
		Interactive: true,
	}}
}

// --- JavaScript Detector Implementation ---

type jsDetector struct{}

func (d jsDetector) Name() string {
	return "JavaScript"
}

func (d jsDetector) Detect() bool {
	if _, err := os.Stat("package.json"); err == nil {
		return true
	}
	// Fallback: check for any .js files.
	jsFiles, _ := filepath.Glob("*.js")
	return len(jsFiles) > 0
}

func (d jsDetector) GetServices() []Service {
	var services []Service
	data, err := ioutil.ReadFile("package.json")
	if err != nil {
		return services
	}
	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return services
	}
	if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
		if dev, ok := scripts["dev"].(string); ok && dev != "" {
			services = append(services, Service{
				Name:        "NPM Dev Script",
				Command:     "npm run dev",
				Interactive: true,
			})
		} else if start, ok := scripts["start"].(string); ok && start != "" {
			services = append(services, Service{
				Name:        "NPM Start",
				Command:     "npm start",
				Interactive: true,
			})
		}
	}
	return services
}

// --- Helper Function ---

func checkDependency(data map[string]interface{}, field, dependency string) bool {
	if deps, ok := data[field].(map[string]interface{}); ok {
		_, exists := deps[dependency]
		return exists
	}
	return false
}

// --- Unified Entrypoint Detection ---

func GetServices() []Service {
	var services []Service
	// Include all detectors.
	detectors := []Detector{
		goDetector{},
		phpDetector{},
		jsDetector{},
		// Add other detectors as needed.
	}
	for _, d := range detectors {
		if d.Detect() {
			services = append(services, d.GetServices()...)
		}
	}
	return services
}
