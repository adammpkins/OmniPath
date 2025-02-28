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

// DetectDependencies reads the composer.json file (if it exists) and returns a list of dependencies
// along with known documentation URLs.
func DetectDependencies() ([]DependencyDocs, error) {
	// For demonstration, we target Laravel.
	if _, err := os.Stat("composer.json"); os.IsNotExist(err) {
		return nil, fmt.Errorf("composer.json not found")
	}
	content, err := ioutil.ReadFile("composer.json")
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	var deps []DependencyDocs
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

	if len(deps) == 0 {
		return nil, fmt.Errorf("no known dependencies found")
	}
	return deps, nil
}
