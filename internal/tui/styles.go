package tui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle           = lipgloss.NewStyle().Margin(1, 2)
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	statusMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render
	pink           = lipgloss.Color("#FF75B5")
	blue           = lipgloss.Color("#00ADD8")
	green          = lipgloss.Color("#00FF00")
	red            = lipgloss.Color("#FF0000")
	yellow         = lipgloss.Color("#FFFF00")
	gray           = lipgloss.Color("240")
	formLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Width(15).Align(lipgloss.Right).PaddingRight(2)
	formValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	formHelpStyle  = lipgloss.NewStyle().Foreground(gray)
	formErrorStyle = lipgloss.NewStyle().Foreground(red).Bold(true)
)
