package omnipath

import (
	"fmt"
	"log"

	"github.com/adammpkins/OmniPath/internal/browser"
	"github.com/adammpkins/OmniPath/internal/docs"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Serve local documentation",
	Run: func(cmd *cobra.Command, args []string) {
		port := "8080"
		dir := "docs"

		url := fmt.Sprintf("http://localhost:%s", port)
		go func() {
			if err := browser.OpenURL(url); err != nil {
				log.Fatalf("Failed to open browser: %v", err)
			}
		}()

		docs.ServeLocalDocs(dir, port)
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
