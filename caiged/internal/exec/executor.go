package exec

import (
	"io"
	"os"
	osexec "os/exec"
)

// CmdExecutor defines the interface for executing commands
// This allows for dependency injection and testing
type CmdExecutor interface {
	// Run executes a command and returns an error if it fails
	Run(name string, args []string, opts RunOptions) error

	// Output executes a command and returns its combined stdout/stderr output
	Output(name string, args []string) ([]byte, error)
}

// RunOptions configures command execution
type RunOptions struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// RealExecutor executes commands using os/exec
type RealExecutor struct{}

// NewRealExecutor creates a new real command executor
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Run executes a command with the given options
func (e *RealExecutor) Run(name string, args []string, opts RunOptions) error {
	cmd := osexec.Command(name, args...)

	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// Output executes a command and returns its combined output
func (e *RealExecutor) Output(name string, args []string) ([]byte, error) {
	cmd := osexec.Command(name, args...)
	return cmd.CombinedOutput()
}
