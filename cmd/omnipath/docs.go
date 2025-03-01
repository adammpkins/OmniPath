package omnipath

import (
	"fmt"
	"log"

	"github.com/adammpkins/OmniPath/internal/browser"
	"github.com/adammpkins/OmniPath/internal/docs"
	"github.com/adammpkins/OmniPath/internal/tui"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Open dependency documentation for the current project",
	Run: func(cmd *cobra.Command, args []string) {
		deps, err := docs.DetectDependencies()
		if err != nil {
			log.Fatalf("Error detecting dependencies: %v", err)
		}

		var selected docs.DependencyDocs
		if len(deps) == 1 {
			selected = deps[0]
		} else {
			// Use our interactive Bubbletea selector.
			selected, err = tui.SelectDependency(deps)
			if err != nil {
				log.Fatalf("Error selecting dependency: %v", err)
			}
		}

		fmt.Printf("Opening documentation for %s: %s\n", selected.Name, selected.DocURL)
		if err := browser.OpenURL(selected.DocURL); err != nil {
			log.Fatalf("Failed to open browser: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
