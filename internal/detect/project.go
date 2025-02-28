package detect

import (
	"os"
	"os/exec"
)

// DetectProjectType checks for common project files and returns the project type and run command.
func DetectProjectType() (string, string) {
	if _, err := os.Stat("go.mod"); err == nil {
		return "Go", "go run ."
	}
	if _, err := os.Stat("package.json"); err == nil {
		return "JavaScript", "npm start"
	}
	if _, err := os.Stat("requirements.txt"); err == nil {
		return "Python", "python main.py"
	}
	return "", ""
}

// RunProject executes the detected run command.
func RunProject(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
