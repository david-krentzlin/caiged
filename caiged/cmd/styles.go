package cmd

import "github.com/charmbracelet/lipgloss"

var (
	// Section headers
	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

	// Emphasis and labels
	LabelStyle = lipgloss.NewStyle().Bold(true)
	ValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))

	// Project/Container names
	ProjectStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	ContainerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Status indicators
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	RunningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StoppedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	ErrorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	// Dividers
	DividerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	SectionDivider = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)

	// Commands
	CommandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

	// Info
	InfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)
