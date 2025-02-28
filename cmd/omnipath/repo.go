package omnipath

import (
	"fmt"
	"log"

	"github.com/adammpkins/OmniPath/internal/browser"
	"github.com/adammpkins/OmniPath/internal/git"

	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Open the GitHub repository in a browser",
	Run: func(cmd *cobra.Command, args []string) {
		remote, err := git.GetRemote()
		if err != nil {
			log.Fatalf("Error retrieving git remote: %v", err)
		}

		url, err := git.ParseRemoteURL(remote)
		if err != nil {
			log.Fatalf("Error parsing remote URL: %v", err)
		}

		fmt.Printf("Opening %s in your browser...\n", url)
		if err := browser.OpenURL(url); err != nil {
			log.Fatalf("Failed to open browser: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
