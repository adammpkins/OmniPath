package omnipath

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/adammpkins/OmniPath/internal/browser"
	"github.com/adammpkins/OmniPath/internal/docs"

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
			fmt.Println("Multiple dependencies detected. Please select one:")
			for i, dep := range deps {
				fmt.Printf("%d: %s\n", i+1, dep.Name)
			}
			fmt.Print("Enter number: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Error reading input: %v", err)
			}
			input = strings.TrimSpace(input)
			idx, err := strconv.Atoi(input)
			if err != nil || idx < 1 || idx > len(deps) {
				log.Fatalf("Invalid selection")
			}
			selected = deps[idx-1]
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
