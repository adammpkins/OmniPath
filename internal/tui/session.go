package tui

import (
	"io"
	"os/exec"
)

// Service represents a runnable service with a name, command, and a flag indicating if it should run interactively.
type Service struct {
	Name        string
	Command     string
	Interactive bool
}

// Session represents a running service with its stdin pipe, accumulated output, and command reference.
type Session struct {
	Name   string         // The name of the service.
	Stdin  io.WriteCloser // The pipe to send input to the process.
	Output string         // Accumulated output from the process.
	Cmd    *exec.Cmd      // Reference to the running command.
}
