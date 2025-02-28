package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetRemote executes "git config --get remote.origin.url" to fetch the remote URL.
func GetRemote() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not a git repository or no remote.origin found")
	}

	return strings.TrimSpace(out.String()), nil
}

// ParseRemoteURL converts a Git remote URL into a browser-friendly URL.
// It supports both SSH (git@...) and HTTPS URLs.
func ParseRemoteURL(remote string) (string, error) {
	if strings.HasPrefix(remote, "git@") {
		// Example: git@github.com:user/repo.git -> https://github.com/user/repo
		remote = strings.TrimPrefix(remote, "git@")
		parts := strings.SplitN(remote, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid remote format: %s", remote)
		}
		domain := parts[0]
		path := strings.TrimSuffix(parts[1], ".git")
		return fmt.Sprintf("https://%s/%s", domain, path), nil
	} else if strings.HasPrefix(remote, "https://") || strings.HasPrefix(remote, "http://") {
		// Remove trailing ".git", if present.
		return strings.TrimSuffix(remote, ".git"), nil
	}

	return "", fmt.Errorf("unsupported remote URL format: %s", remote)
}
