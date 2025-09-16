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
		case "tab":
			m.taskFormFocused = (m.taskFormFocused + 1) % len(m.taskFormInputs)
			for i := 0; i < len(m.taskFormInputs); i++ {
				m.taskFormInputs[i].Blur()
			}
			m.taskFormInputs[m.taskFormFocused].Focus()
			return m, textinput.Blink

		case "shift+tab":
			m.taskFormFocused = (m.taskFormFocused - 1 + len(m.taskFormInputs)) % len(m.taskFormInputs)
			for i := 0; i < len(m.taskFormInputs); i++ {
				m.taskFormInputs[i].Blur()
			}
			m.taskFormInputs[m.taskFormFocused].Focus()
			return m, textinput.Blink

		case "enter":
			if m.taskFormFocused < len(m.taskFormInputs)-1 {
				m.taskFormFocused++
				for i := 0; i < len(m.taskFormInputs); i++ {
					m.taskFormInputs[i].Blur()
				}
				m.taskFormInputs[m.taskFormFocused].Focus()
				return m, textinput.Blink
			}
		case "esc":
			m.state = viewProjects
			m.status = statusMessageStyle("Task creation cancelled.")
			for i := 0; i < len(m.taskFormInputs); i++ {
				m.taskFormInputs[i].SetValue("")
				m.taskFormInputs[i].Blur()
			}
			m.taskFormFocused = 0
			m.err = nil
			return m, nil

		case "ctrl+s":
			for i := 0; i < len(m.taskFormInputs); i++ {
				input := &m.taskFormInputs[i]
				if i == 0 && input.Value() == "" {
					m.err = fmt.Errorf("title is required")
					m.taskFormFocused = i
					input.Focus()
					return m, textinput.Blink
				}
				if input.Validate != nil {
					if err := input.Validate(input.Value()); err != nil {
						m.err = fmt.Errorf("%s: %v", input.Placeholder, err)
						m.taskFormFocused = i
						input.Focus()
						return m, textinput.Blink
					}
				}
			}
			m.err = nil

			if m.currentProject == nil {
				m.err = fmt.Errorf("internal error: no project selected for task creation")
				return m, nil
			}

			m.loading = true
			m.status = statusMessageStyle("Creating task...")
			return m, m.apiClient.createTaskFormCmd(
				m.currentProject.ID,
				m.taskFormInputs[0].Value(), // Title
				m.taskFormInputs[1].Value(), // Description
				m.taskFormInputs[2].Value(), // Difficulty
				m.taskFormInputs[3].Value(), // Estimated Hours
				m.taskFormInputs[4].Value(), // Tags (JSON string)
				m.taskFormInputs[5].Value(), // Skills Required (JSON string)
				m.taskFormInputs[6].Value(), // Bounty Amount
			)
		}
	}

	var inputCmd tea.Cmd
	m.taskFormInputs[m.taskFormFocused], inputCmd = m.taskFormInputs[m.taskFormFocused].Update(msg)
	if inputCmd != nil {
		cmds = append(cmds, inputCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) viewCreateTaskView() string {
	var b strings.Builder
	titleContent := fmt.Sprintf(" Create New Task for Project: %s (ID: %d) ", m.currentProject.Title, m.currentProject.ID)
	b.WriteString(titleStyle.Width(m.width - appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Render(titleContent))
	b.WriteString("\n\n")

	for i, input := range m.taskFormInputs {
		label := formLabelStyle.Render(input.Placeholder + ":")
		inputView := input.View()
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, label, inputView))
		if i == m.taskFormFocused {
			b.WriteString(formValueStyle.Render(" <"))
		}
		if input.Err != nil {
			b.WriteString(formErrorStyle.Render(" " + input.Err.Error()))
		}
		b.WriteString("\n")
	}

	b.WriteString(formHelpStyle.Render("\nTab: next field • Shift+Tab: prev field • Enter: next field • Ctrl+S: submit • Esc: cancel"))

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(formErrorStyle.Render(fmt.Sprintf("Submission Error: %v", m.err)))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}