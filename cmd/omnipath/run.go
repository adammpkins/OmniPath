package omnipath

import (
	"fmt"
	"log"

	"github.com/adammpkins/OmniPath/internal/detect"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Detects and runs the project based on its type",
	Run: func(cmd *cobra.Command, args []string) {
		projectType, command := detect.DetectProjectType()
		if projectType == "" {
			log.Println("Unknown project type. No run command available.")
			return
		}

		fmt.Printf("Detected project type: %s\nRunning command: %s\n", projectType, command)
		err := detect.RunProject(command)
		if err != nil {
			log.Fatalf("Error running project: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
