package omnipath

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "omnipath",
	Short: "OmniPath - A smart directory-based automation tool",
	Long:  "OmniPath helps navigate projects, open repositories, serve local docs, open dependency documentation, and auto-run projects based on context.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Omnipath CLI - Use a subcommand. Try 'omnipath repo', 'omnipath docs', 'omnipath depdocs' or 'omnipath run'")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
