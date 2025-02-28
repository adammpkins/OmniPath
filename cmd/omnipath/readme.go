package omnipath

import (
	"log"

	"github.com/adammpkins/OmniPath/internal/browser"
	"github.com/adammpkins/OmniPath/internal/docs"

	"github.com/spf13/cobra"
)

var readmeCmd = &cobra.Command{
	Use:   "readme",
	Short: "Serve README.md as HTML with dark styling",
	Run: func(cmd *cobra.Command, args []string) {
		port := "8080"
		readmePath := "README.md"

		go func() {
			url := "http://localhost:" + port
			if err := browser.OpenURL(url); err != nil {
				log.Fatalf("Failed to open browser: %v", err)
			}
		}()

		docs.ServeReadmeAsHTML(readmePath, port)
	},
}

func init() {
	rootCmd.AddCommand(readmeCmd)
}
