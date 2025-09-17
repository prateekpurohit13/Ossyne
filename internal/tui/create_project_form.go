package tui

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type createProjectFormModel struct {
	titleInput       textinput.Model
	descriptionInput textinput.Model
	repoURLInput     textinput.Model
	tagsInput        textinput.Model
	focusedInput     int
	isPublic         bool
	err              error
}

func newCreateProjectFormModel() createProjectFormModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter project title"
	titleInput.Focus()
	titleInput.CharLimit = 100
	titleInput.Width = 50

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "Enter project description"
	descriptionInput.CharLimit = 500
	descriptionInput.Width = 50

	repoURLInput := textinput.New()
	repoURLInput.Placeholder = "https://github.com/owner/repo (optional)"
	repoURLInput.CharLimit = 200
	repoURLInput.Width = 50

	tagsInput := textinput.New()
	tagsInput.Placeholder = "Enter tags separated by commas (e.g., go, web, api)"
	tagsInput.CharLimit = 200
	tagsInput.Width = 50

	return createProjectFormModel{
		titleInput:       titleInput,
		descriptionInput: descriptionInput,
		repoURLInput:     repoURLInput,
		tagsInput:        tagsInput,
		focusedInput:     0,
		isPublic:         true,
	}
}

func (m createProjectFormModel) updateCreateProjectForm(msg tea.Msg) (createProjectFormModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusedInput == 4 {
				return m.submitForm()
			}

			if s == "up" || s == "shift+tab" {
				m.focusedInput--
			} else {
				m.focusedInput++
			}

			if m.focusedInput > 4 {
				m.focusedInput = 0
			} else if m.focusedInput < 0 {
				m.focusedInput = 4
			}

			cmds = append(cmds, (&m).updateFocus())
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		case " ":
			if m.focusedInput == 4 {
				m.isPublic = !m.isPublic
			}
		case "ctrl+s":
			return m.submitForm()
		}
	}

	switch m.focusedInput {
	case 0:
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		m.descriptionInput, cmd = m.descriptionInput.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		m.repoURLInput, cmd = m.repoURLInput.Update(msg)
		cmds = append(cmds, cmd)
	case 3:
		m.tagsInput, cmd = m.tagsInput.Update(msg)
		cmds = append(cmds, cmd)
	case 4:
	}

	return m, tea.Batch(cmds...)
}

func (m createProjectFormModel) inputs() []*textinput.Model {
	return []*textinput.Model{
		&m.titleInput,
		&m.descriptionInput,
		&m.repoURLInput,
		&m.tagsInput,
	}
}

func (m *createProjectFormModel) updateFocus() tea.Cmd {
	switch m.focusedInput {
	case 0:
		m.titleInput.Focus()
		m.descriptionInput.Blur()
		m.repoURLInput.Blur()
		m.tagsInput.Blur()
	case 1:
		m.titleInput.Blur()
		m.descriptionInput.Focus()
		m.repoURLInput.Blur()
		m.tagsInput.Blur()
	case 2:
		m.titleInput.Blur()
		m.descriptionInput.Blur()
		m.repoURLInput.Focus()
		m.tagsInput.Blur()
	case 3:
		m.titleInput.Blur()
		m.descriptionInput.Blur()
		m.repoURLInput.Blur()
		m.tagsInput.Focus()
	case 4:
		m.titleInput.Blur()
		m.descriptionInput.Blur()
		m.repoURLInput.Blur()
		m.tagsInput.Blur()
	}
	return nil
}

func (m createProjectFormModel) submitForm() (createProjectFormModel, tea.Cmd) {
	if strings.TrimSpace(m.titleInput.Value()) == "" {
		m.err = fmt.Errorf("project title is required")
		return m, nil
	}

	if strings.TrimSpace(m.descriptionInput.Value()) == "" {
		m.err = fmt.Errorf("project description is required")
		return m, nil
	}

	var tags []string
	if tagsStr := strings.TrimSpace(m.tagsInput.Value()); tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	// Create project data
	projectData := map[string]interface{}{
		"title":       strings.TrimSpace(m.titleInput.Value()),
		"description": strings.TrimSpace(m.descriptionInput.Value()),
		"is_public":   m.isPublic,
		"tags":        tags,
	}

	if repoURL := strings.TrimSpace(m.repoURLInput.Value()); repoURL != "" {
		projectData["repository_url"] = repoURL
	}

	m.err = nil
	return m, func() tea.Msg {
		return projectCreateSubmitMsg{
			title:       strings.TrimSpace(m.titleInput.Value()),
			description: strings.TrimSpace(m.descriptionInput.Value()),
			repositoryURL: func() string {
				if url := strings.TrimSpace(m.repoURLInput.Value()); url != "" {
					return url
				}
				return ""
			}(),
			isPublic: m.isPublic,
			tags:     tags,
		}
	}
}

func (m createProjectFormModel) viewCreateProjectForm() string {
	var b strings.Builder
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("13")).
		Padding(0, 1).
		Align(lipgloss.Center).
		Width(54)

	b.WriteString(headerStyle.Render("Create New Project"))
	b.WriteString("\n\n")

	// Form fields
	fieldStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	inputs := m.inputs()
	labels := []string{"Title:", "Description:", "Repository URL:", "Tags:"}

	for i, input := range inputs {
		b.WriteString(fieldStyle.Render(labels[i]))
		b.WriteString("\n")
		inputStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Width(52)

		if m.focusedInput == i {
			inputStyle = inputStyle.BorderForeground(lipgloss.Color("39"))
		} else {
			inputStyle = inputStyle.BorderForeground(lipgloss.Color("240"))
		}

		b.WriteString(inputStyle.Render(input.View()))
		b.WriteString("\n\n")
	}

	b.WriteString(fieldStyle.Render("Visibility:"))
	b.WriteString("\n")

	visibilityStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(52)

	if m.focusedInput == 4 {
		visibilityStyle = visibilityStyle.BorderForeground(lipgloss.Color("39"))
	} else {
		visibilityStyle = visibilityStyle.BorderForeground(lipgloss.Color("240"))
	}

	visibilityText := "🔓 Public"
	if !m.isPublic {
		visibilityText = "🔒 Private"
	}
	b.WriteString(visibilityStyle.Render(visibilityText))
	b.WriteString("\n\n")
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Italic(true)

	b.WriteString(instructionStyle.Render("• Use Tab/Shift+Tab or Up/Down to navigate"))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("• Press Space to toggle visibility"))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("• Press Ctrl+S or Enter to submit"))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("• Press Esc to cancel"))
	b.WriteString("\n")

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.err.Error()))
	}

	formContent := b.String()
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("13")).
		Padding(2, 3).
		MarginTop(1).
		MarginBottom(1)

	return panelStyle.Render(formContent)
}