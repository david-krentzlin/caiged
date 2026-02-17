package exec

import (
	"bytes"
	"fmt"
	"strings"
)

// MockExecutor is a test implementation of CmdExecutor
type MockExecutor struct {
	// Commands stores all commands that were executed
	Commands []MockCommand

	// Responses maps command patterns to responses
	// Key format: "command arg1 arg2"
	Responses map[string]MockResponse

	// DefaultResponse is used when no specific response is found
	DefaultResponse MockResponse
}

// MockCommand represents a captured command execution
type MockCommand struct {
	Name string
	Args []string
	Opts RunOptions
}

// MockResponse defines the response for a mocked command
type MockResponse struct {
	Output []byte
	Error  error
}

// NewMockExecutor creates a new mock executor
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		Commands:  []MockCommand{},
		Responses: make(map[string]MockResponse),
		DefaultResponse: MockResponse{
			Output: []byte{},
			Error:  nil,
		},
	}
}

// AddResponse adds a response for a specific command
func (m *MockExecutor) AddResponse(name string, args []string, output string, err error) {
	key := m.makeKey(name, args)
	m.Responses[key] = MockResponse{
		Output: []byte(output),
		Error:  err,
	}
}

// AddResponseForPrefix adds a response for any command starting with the given pattern
func (m *MockExecutor) AddResponseForPrefix(name string, output string, err error) {
	m.Responses[name] = MockResponse{
		Output: []byte(output),
		Error:  err,
	}
}

// Run executes a mock command
func (m *MockExecutor) Run(name string, args []string, opts RunOptions) error {
	m.Commands = append(m.Commands, MockCommand{
		Name: name,
		Args: args,
		Opts: opts,
	})

	response := m.findResponse(name, args)

	// Write output to stdout if provided
	if opts.Stdout != nil && len(response.Output) > 0 {
		opts.Stdout.Write(response.Output)
	}

	return response.Error
}

// Output executes a mock command and returns output
func (m *MockExecutor) Output(name string, args []string) ([]byte, error) {
	m.Commands = append(m.Commands, MockCommand{
		Name: name,
		Args: args,
	})

	response := m.findResponse(name, args)
	return response.Output, response.Error
}

// GetCommand returns the nth command that was executed (0-indexed)
func (m *MockExecutor) GetCommand(index int) (MockCommand, bool) {
	if index < 0 || index >= len(m.Commands) {
		return MockCommand{}, false
	}
	return m.Commands[index], true
}

// GetLastCommand returns the most recent command
func (m *MockExecutor) GetLastCommand() (MockCommand, bool) {
	if len(m.Commands) == 0 {
		return MockCommand{}, false
	}
	return m.Commands[len(m.Commands)-1], true
}

// CommandCount returns the number of commands executed
func (m *MockExecutor) CommandCount() int {
	return len(m.Commands)
}

// Reset clears all recorded commands
func (m *MockExecutor) Reset() {
	m.Commands = []MockCommand{}
}

// findResponse finds the appropriate response for a command
func (m *MockExecutor) findResponse(name string, args []string) MockResponse {
	// Try exact match first
	key := m.makeKey(name, args)
	if response, ok := m.Responses[key]; ok {
		return response
	}

	// Try prefix match (just command name)
	if response, ok := m.Responses[name]; ok {
		return response
	}

	// Try partial matches (command + first few args)
	for i := len(args); i > 0; i-- {
		partialKey := m.makeKey(name, args[:i])
		if response, ok := m.Responses[partialKey]; ok {
			return response
		}
	}

	return m.DefaultResponse
}

// makeKey creates a lookup key from command and args
func (m *MockExecutor) makeKey(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return name + " " + strings.Join(args, " ")
}

// String returns a string representation of the mock executor state
func (m *MockExecutor) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("MockExecutor: %d commands executed\n", len(m.Commands)))
	for i, cmd := range m.Commands {
		buf.WriteString(fmt.Sprintf("  [%d] %s %v\n", i, cmd.Name, cmd.Args))
	}
	return buf.String()
}

// AssertCommandExecuted checks if a specific command was executed
func (m *MockExecutor) AssertCommandExecuted(t interface{ Errorf(string, ...interface{}) }, name string, args ...string) bool {
	for _, cmd := range m.Commands {
		if cmd.Name == name {
			if len(args) == 0 {
				return true
			}
			if m.argsMatch(cmd.Args, args) {
				return true
			}
		}
	}
	t.Errorf("Command not executed: %s %v\nExecuted commands:\n%s", name, args, m.String())
	return false
}

// argsMatch checks if two argument slices match
func (m *MockExecutor) argsMatch(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range actual {
		if actual[i] != expected[i] {
			return false
		}
	}
	return true
}

// MockError creates a simple error for testing
func MockError(msg string) error {
	return fmt.Errorf("%s", msg)
}
