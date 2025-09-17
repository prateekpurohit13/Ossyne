package tui

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) updateManageProjectsView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.err
		m.status = statusMessageStyle(fmt.Sprintf("Error: %v", msg.err))
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch keypress := msg.String(); keypress {
		case "esc", "b":
			m.state = viewLanding
			if m.projectsList.FilterInput.Focused() {
				m.projectsList.FilterInput.Blur()
				m.projectsList.FilterInput.SetValue("")
			}
			m.status = statusMessageStyle("Returned to landing page.")
			return m, nil

		case "r":
			if m.loggedInUser != nil {
				m.loading = true
				m.status = statusMessageStyle("Refreshing your projects...")
				return m, m.apiClient.fetchUserProjectsCmd(m.loggedInUser.ID)
			}
			return m, nil

		case "enter":
			selectedItem := m.projectsList.SelectedItem()
			if selectedItem != nil {
				project := selectedItem.(projectItem).Project
				m.currentProject = &project
				m.loading = true
				m.status = statusMessageStyle(fmt.Sprintf("Loading tasks for project '%s'...", project.Title))
				return m, m.apiClient.fetchProjectTasksCmd(project.ID)
			} else {
				m.status = statusMessageStyle("No project selected")
			}
			return m, nil

		case "p":
			m.state = viewCreateProjectForm
			m.createProjectForm = newCreateProjectFormModel()
			return m, nil

		case "t":
			selectedItem := m.projectsList.SelectedItem()
			if selectedItem == nil {
				m.status = statusMessageStyle("Select a project first to create a task.")
				return m, nil
			}
			project := selectedItem.(projectItem).Project
			m.currentProject = &project
			m.state = viewCreateTask
			m.err = nil
			for i := 0; i < len(m.taskFormInputs); i++ {
				m.taskFormInputs[i].SetValue("")
				m.taskFormInputs[i].Blur()
			}
			m.taskFormFocused = 0
			m.taskFormInputs[m.taskFormFocused].Focus()
			m.status = statusMessageStyle(fmt.Sprintf("Creating task for project '%s'", project.Title))
			return m, textinput.Blink

		case "f":
			selectedItem := m.projectsList.SelectedItem()
			if selectedItem == nil {
				m.status = statusMessageStyle("Select a project first to fund bounties.")
				return m, nil
			}
			project := selectedItem.(projectItem).Project
			m.currentProject = &project
			m.loading = true
			m.status = statusMessageStyle(fmt.Sprintf("Loading tasks for funding in project '%s'...", project.Title))
			return m, m.apiClient.fetchProjectTasksCmd(project.ID)
		}
	}

	// Update project list
	newProjectListModel, listCmd := m.projectsList.Update(msg)
	m.projectsList = newProjectListModel
	sel := m.projectsList.Index()
	items := m.projectsList.Items()
	if len(items) > 0 && sel >= 0 && sel < len(items) {
		project := items[sel].(projectItem).Project
		m.currentProject = &project
	} else {
		m.currentProject = nil
	}
	cmds = append(cmds, listCmd)
	return m, tea.Batch(cmds...)
}

func (m model) viewManageProjectsView() string {
	if m.loggedInUser == nil {
		content := "Authentication required to access project management."
		content += "\n\nPlease login first to create and manage your projects."
		content += "\n\n[esc] Back to Landing • [l] Login"
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	spinnerView := ""
	if m.loading {
		spinnerView = " " + m.spinner.View() + " Loading..."
	} else if m.err != nil {
		spinnerView = " " + lipgloss.NewStyle().Foreground(red).Render(m.err.Error())
	}
	headerContent := fmt.Sprintf("Project Management - %s%s", m.loggedInUser.Username, spinnerView)
	header := titleStyle.Width(m.width).
		Align(lipgloss.Center).
		Render(headerContent)

	panelWidth := m.width/3 - 2
	panelHeight := m.height - 4

	leftPanel := m.renderProjectListPanel(panelWidth, panelHeight)
	middlePanel := m.renderProjectDetailsPanel(panelWidth, panelHeight)
	rightPanel := m.renderActionsPanel(panelWidth, panelHeight)

	statusBar := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Left).
		Render(m.status)

	helpText := "↑/k up • ↓/j down • enter view tasks • p create project • t create task • f fund bounty • r refresh • esc back"
	help := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(gray).
		Render(helpText)

	ui := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, middlePanel, rightPanel),
		statusBar,
		help,
	)

	return ui
}

func (m model) renderProjectListPanel(width, height int) string {
	content := m.projectsList.View()

	if len(m.projectsList.Items()) == 0 {
		content = lipgloss.NewStyle().
			Foreground(gray).
			Italic(true).
			Render("No projects found.\nPress 'p' to create your first project!")
	}

	title := lipgloss.NewStyle().
		Foreground(green).
		Bold(true).
		Render(" 📂 Your Projects ")

	panel := lipgloss.NewStyle().
		Width(width).
		Height(height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(green).
		Padding(1).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Top, title, panel)
}

func (m model) renderProjectDetailsPanel(width, height int) string {
	var content string
	if m.currentProject != nil {
		var sb strings.Builder
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(yellow).Render(m.currentProject.Title))
		sb.WriteString("\n\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(gray).Render("ID: "))
		sb.WriteString(fmt.Sprintf("%d\n", m.currentProject.ID))
		sb.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Description: "))
		sb.WriteString(fmt.Sprintf("%s\n\n", m.currentProject.ShortDesc))
		sb.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Repository: "))
		sb.WriteString(fmt.Sprintf("%s\n", m.currentProject.RepoURL))
		sb.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Visibility: "))
		sb.WriteString(lipgloss.NewStyle().Foreground(blue).Render(strings.ToTitle(m.currentProject.Visibility)))
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Created: "))
		sb.WriteString(fmt.Sprintf("%s\n", m.currentProject.CreatedAt.Format("2006-01-02 15:04")))

		if len(m.currentProject.Tags) > 0 {
			sb.WriteString("\n")
			sb.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Tags: "))
			for i, tag := range m.currentProject.Tags {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(lipgloss.NewStyle().
					Background(lipgloss.Color("240")).
					Foreground(lipgloss.Color("15")).
					Padding(0, 1).
					Render(tag))
			}
		}

		content = sb.String()
	} else {
		content = lipgloss.NewStyle().
			Foreground(gray).
			Italic(true).
			Render("Select a project from the list to view details")
	}

	title := lipgloss.NewStyle().
		Foreground(yellow).
		Bold(true).
		Render("Project Details ")
	panel := lipgloss.NewStyle().
		Width(width).
		Height(height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(yellow).
		Padding(1).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Top, title, panel)
}

func (m model) renderActionsPanel(width, height int) string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(blue).Render("Quick Actions"))
	sb.WriteString("\n\n")

	actions := []struct {
		key    string
		desc   string
		color  lipgloss.Color
		active bool
	}{
		{"p", "Create New Project", green, true},
		{"t", "Create Task", yellow, m.currentProject != nil},
		{"f", "Fund Bounties", pink, m.currentProject != nil},
		{"enter", "View Tasks", blue, m.currentProject != nil},
	}

	for _, action := range actions {
		style := lipgloss.NewStyle().Foreground(gray)
		if action.active {
			style = lipgloss.NewStyle().Foreground(action.color)
		}

		keyStyle := lipgloss.NewStyle().
			Background(action.color).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)

		if !action.active {
			keyStyle = keyStyle.Background(gray)
		}

		sb.WriteString(keyStyle.Render(action.key))
		sb.WriteString(" ")
		sb.WriteString(style.Render(action.desc))
		sb.WriteString("\n")
	}

	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(pink).Render("Quick Stats"))
	sb.WriteString("\n\n")

	projectCount := len(m.projectsList.Items())
	sb.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Total Projects: "))
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(green).Render(fmt.Sprintf("%d", projectCount)))
	sb.WriteString("\n")

	if m.currentProject != nil {
		sb.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Selected: "))
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(yellow).Render(m.currentProject.Title))
		sb.WriteString("\n")
	}

	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(purple).Render("💡 Tips"))
	sb.WriteString("\n\n")

	tips := []string{
		"• Select a project to enable task actions",
		"• Created projects appear in Browse Projects",
		"• Fund bounties to attract contributors",
		"• Use CLI for advanced project creation",
	}

	for _, tip := range tips {
		sb.WriteString(lipgloss.NewStyle().Foreground(gray).Italic(true).Render(tip))
		sb.WriteString("\n")
	}

	title := lipgloss.NewStyle().
		Foreground(purple).
		Bold(true).
		Render(" Actions & Info ")

	panel := lipgloss.NewStyle().
		Width(width).
		Height(height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(1).
		Render(sb.String())

	return lipgloss.JoinVertical(lipgloss.Top, title, panel)
}