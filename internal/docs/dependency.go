package docs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// DependencyDocs holds information about a dependency and its documentation URL.
type DependencyDocs struct {
	Name   string
	DocURL string
}

// DetectDependencies reads the composer.json and go.mod files (if they exist)
// and returns a list of dependencies along with known documentation URLs.
func DetectDependencies() ([]DependencyDocs, error) {
	var deps []DependencyDocs

	// Check for composer.json
	if _, err := os.Stat("composer.json"); err == nil {
		content, err := ioutil.ReadFile("composer.json")
		if err != nil {
			return nil, err
		}

		var data map[string]interface{}
		if err := json.Unmarshal(content, &data); err != nil {
			return nil, err
		}

		// composer.json typically has a "require" field with dependencies.
		if req, ok := data["require"].(map[string]interface{}); ok {
			for key := range req {
				if strings.EqualFold(key, "laravel/framework") {
					deps = append(deps, DependencyDocs{
						Name:   "Laravel",
						DocURL: "https://laravel.com/docs",
					})
				}
				// Extend this block for additional known dependencies as needed.
			}
		}
	}

	// Check for go.mod
	if _, err := os.Stat("go.mod"); err == nil {
		// For demonstration, we add a dependency for Go documentation.
		deps = append(deps, DependencyDocs{
			Name:   "Go",
			DocURL: "https://golang.org/doc/",
		})
		// Check the go.mod for Fiber dependency
		content, err := ioutil.ReadFile("go.mod")
		if err != nil {
			return nil, err
		}
		if strings.Contains(string(content), "github.com/gofiber/fiber") {
			deps = append(deps, DependencyDocs{
				Name:   "Fiber",
				DocURL: "https://docs.gofiber.io/",
			})
		}
	}

	if len(deps) == 0 {
		return nil, fmt.Errorf("no known dependencies found")
	}
	return deps, nil
}
