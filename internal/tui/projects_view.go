package tui

import (
	"fmt"
	"ossyne/internal/models"
	"strings"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type projectItem struct {
	models.Project
}

func (i projectItem) FilterValue() string { return i.Project.Title }
func (i projectItem) Title() string       { return i.Project.Title }
func (i projectItem) Description() string {
	return fmt.Sprintf("ID: %d | Description: %s", i.Project.ID, i.Project.ShortDesc)
}

var _ list.Item = projectItem{}

func (m model) updateProjectsView(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "enter":
			selectedItem := m.projectsList.SelectedItem()
			if selectedItem != nil {
				project := selectedItem.(projectItem).Project
				m.currentProject = &project
				m.status = statusMessageStyle(fmt.Sprintf("Selected project ID: %d", m.currentProject.ID))
				m.loading = true
				m.status = statusMessageStyle(fmt.Sprintf("Loading tasks for project '%s'...", project.Title))
				return m, m.apiClient.fetchProjectTasksCmd(project.ID)
			} else {
				m.status = statusMessageStyle("No project selected")
			}
			return m, nil

		case "esc", "b":
			m.state = viewLanding
			m.projectsList.FilterInput.Blur()
			m.projectsList.FilterInput.SetValue("")
			m.status = statusMessageStyle("Returned to landing page.")
			return m, nil
		case "r":
			m.loading = true
			m.status = statusMessageStyle("Refreshing projects...")
			const maintainerID = 1
			return m, m.apiClient.fetchUserProjectsCmd(maintainerID)
		}
	}

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

func (m model) viewProjectsView() string {
	spinnerView := ""
	if m.loading {
		spinnerView = m.spinner.View() + " Loading..."
	} else if m.err != nil {
		spinnerView = lipgloss.NewStyle().Foreground(red).Render(m.err.Error())
	}

	header := titleStyle.Render(fmt.Sprintf(" Browse Public Projects (%s)", spinnerView))
	panelWidth := (m.width-appStyle.GetHorizontalFrameSize())/2 - 1
	leftPanel := lipgloss.NewStyle().
		Width(panelWidth).
		Height(m.height - appStyle.GetVerticalFrameSize() - 6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(green).
		Padding(1).
		Render(m.projectsList.View())
	projectDetails := "Select a project to view details."
	if m.currentProject != nil {
		projectDetails = m.renderProjectDetails(m.currentProject)
	} else if sel := m.projectsList.Index(); sel >= 0 {
		if items := m.projectsList.Items(); len(items) > 0 && sel < len(items) {
			project := items[sel].(projectItem).Project
			m.currentProject = &project
			projectDetails = m.renderProjectDetails(&project)
		}
	}

	rightPanel := lipgloss.NewStyle().
		Width(panelWidth).
		Height(m.height - appStyle.GetVerticalFrameSize() - 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(yellow).
		Padding(1).
		Render(projectDetails)
	helpText := "↑/k up • ↓/j down • enter view tasks • r refresh • esc back to landing"
	ui := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel),
		lipgloss.NewStyle().Width(m.width-appStyle.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(m.status),
		lipgloss.NewStyle().Width(m.width-appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Foreground(gray).Render(helpText),
	)
	return ui
}

func (m model) renderProjectDetails(project *models.Project) string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(green).Render("Project Details:\n"))
	sb.WriteString(fmt.Sprintf("ID: %d\n", project.ID))
	sb.WriteString(fmt.Sprintf("Title: %s\n", lipgloss.NewStyle().Foreground(yellow).Render(project.Title)))
	sb.WriteString(fmt.Sprintf("Description: %s\n", project.ShortDesc))
	sb.WriteString(fmt.Sprintf("Owner ID: %d\n", project.OwnerID))
	sb.WriteString(fmt.Sprintf("Repository URL: %s\n", project.RepoURL))
	sb.WriteString(fmt.Sprintf("Visibility: %s\n", lipgloss.NewStyle().Foreground(blue).Render(strings.ToTitle(project.Visibility))))
	if len(project.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(project.Tags, ", ")))
	}
	sb.WriteString(fmt.Sprintf("Created: %s\n", project.CreatedAt.Format("2006-01-02 15:04")))
	return sb.String()
}