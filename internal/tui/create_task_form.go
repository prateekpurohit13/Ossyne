package tui

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) updateCreateTaskView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.taskFormFocused = (m.taskFormFocused + 1) % len(m.taskFormInputs)
			m.updateTaskFormFocus()
			return m, textinput.Blink

		case "shift+tab", "up":
			m.taskFormFocused = (m.taskFormFocused - 1 + len(m.taskFormInputs)) % len(m.taskFormInputs)
			m.updateTaskFormFocus()
			return m, textinput.Blink

		case "enter":
			if m.taskFormFocused < len(m.taskFormInputs)-1 {
				m.taskFormFocused++
				m.updateTaskFormFocus()
				return m, textinput.Blink
			} else {
				return m.submitTaskForm()
			}

		case "esc":
			m.state = viewManageProjects
			m.status = statusMessageStyle("Task creation cancelled.")
			m.resetTaskForm()
			return m, nil

		case "ctrl+s":
			return m.submitTaskForm()
		}
	}

	var inputCmd tea.Cmd
	m.taskFormInputs[m.taskFormFocused], inputCmd = m.taskFormInputs[m.taskFormFocused].Update(msg)
	if inputCmd != nil {
		cmds = append(cmds, inputCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateTaskFormFocus() {
	for i := 0; i < len(m.taskFormInputs); i++ {
		if i == m.taskFormFocused {
			m.taskFormInputs[i].Focus()
		} else {
			m.taskFormInputs[i].Blur()
		}
	}
}

func (m *model) resetTaskForm() {
	for i := 0; i < len(m.taskFormInputs); i++ {
		m.taskFormInputs[i].SetValue("")
		m.taskFormInputs[i].Blur()
	}
	m.taskFormFocused = 0
	m.err = nil
}

func (m model) submitTaskForm() (tea.Model, tea.Cmd) {
	titleValue := strings.TrimSpace(m.taskFormInputs[0].Value())
	if titleValue == "" {
		m.err = fmt.Errorf("task title is required")
		m.taskFormFocused = 0
		m.taskFormInputs[0].Focus()
		return m, textinput.Blink
	}

	descriptionValue := strings.TrimSpace(m.taskFormInputs[1].Value())
	if descriptionValue == "" {
		m.err = fmt.Errorf("task description is required")
		m.taskFormFocused = 1
		m.taskFormInputs[1].Focus()
		return m, textinput.Blink
	}

	for i := 0; i < len(m.taskFormInputs); i++ {
		input := &m.taskFormInputs[i]
		if input.Validate != nil {
			if err := input.Validate(input.Value()); err != nil {
				m.err = fmt.Errorf("validation error in %s: %v",
					[]string{"Title", "Description", "Difficulty", "Estimated Hours", "Tags", "Skills Required", "Bounty Amount"}[i],
					err)
				m.taskFormFocused = i
				input.Focus()
				return m, textinput.Blink
			}
		}
	}

	if m.currentProject == nil {
		m.err = fmt.Errorf("internal error: no project selected for task creation")
		return m, nil
	}

	m.err = nil
	m.loading = true
	m.status = statusMessageStyle("Creating task...")

	return m, m.apiClient.createTaskFormCmd(
		m.currentProject.ID,
		titleValue,                  // Title
		descriptionValue,            // Description
		m.taskFormInputs[2].Value(), // Difficulty
		m.taskFormInputs[3].Value(), // Estimated Hours
		m.taskFormInputs[4].Value(), // Tags (JSON string)
		m.taskFormInputs[5].Value(), // Skills Required (JSON string)
		m.taskFormInputs[6].Value(), // Bounty Amount
	)
}

func (m model) viewCreateTaskView() string {
	var b strings.Builder
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("39")).
		Padding(0, 1).
		Align(lipgloss.Center)

	headerContent := fmt.Sprintf("Create New Task for Project: %s", m.currentProject.Title)
	b.WriteString(headerStyle.Render(headerContent))
	b.WriteString("\n\n")

	fieldStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	labels := []string{
		"Title:",
		"Description:",
		"Difficulty:",
		"Estimated Hours:",
		"Tags:",
		"Skills Required:",
		"Bounty Amount:",
	}

	placeholderHints := []string{
		"Enter a clear, concise task title",
		"Provide detailed task requirements and expectations",
		"easy, medium, or hard",
		"How many hours do you estimate? (e.g., 8)",
		"JSON format: [\"go\", \"cli\", \"backend\"]",
		"JSON format: [\"sql\", \"testing\", \"api-design\"]",
		"Amount in USD (e.g., 50.00)",
	}

	for i, input := range m.taskFormInputs {
		b.WriteString(fieldStyle.Render(labels[i]))
		b.WriteString("\n")
		inputStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0, 1)

		if i == m.taskFormFocused {
			inputStyle = inputStyle.BorderForeground(lipgloss.Color("39"))
		} else {
			inputStyle = inputStyle.BorderForeground(lipgloss.Color("240"))
		}
		inputView := input.View()
		if input.Value() == "" && i != m.taskFormFocused {
			hintStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")).
				Italic(true)
			inputView = hintStyle.Render(placeholderHints[i])
		}

		b.WriteString(inputStyle.Render(inputView))

		if input.Err != nil {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true).
				MarginTop(1)
			b.WriteString("\n")
			b.WriteString(errorStyle.Render("Error: " + input.Err.Error()))
		}
		b.WriteString("\n\n")
	}

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Italic(true)

	b.WriteString(instructionStyle.Render("• Use Tab/Shift+Tab or Up/Down to navigate"))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("• Press Enter to move to next field"))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("• Press Ctrl+S or Enter on last field to submit"))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("• Press Esc to cancel"))
	b.WriteString("\n")

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Submission Error: " + m.err.Error()))
	}

	formContent := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, formContent)
}