package tui

import (
	"fmt"
	"ossyne/internal/models"
	"strings"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) updateProjectTasksView(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			selectedItem := m.projectTasksList.SelectedItem()
			if selectedItem != nil {
				task := selectedItem.(taskItem).Task
				m.currentTask = &task
				m.status = statusMessageStyle(fmt.Sprintf("Selected task ID: %d", m.currentTask.ID))
			} else {
				m.status = statusMessageStyle("No task selected")
			}
			return m, nil
		case "esc", "p":
			m.state = viewProjects
			m.projectTasksList.FilterInput.Blur()
			m.projectTasksList.FilterInput.SetValue("")
			m.projectsList.Select(m.projectsList.Index())
			m.status = statusMessageStyle(fmt.Sprintf("Returned to projects view for '%s'.", m.currentProject.Title))
			return m, nil
		case "r":
			m.loading = true
			if m.currentProject != nil {
				m.status = statusMessageStyle("Refreshing tasks for project...")
				return m, m.apiClient.fetchProjectTasksCmd(m.currentProject.ID)
			} else {
				m.loading = false
				m.status = statusMessageStyle("No project selected to refresh tasks.")
				return m, nil
			}
		case "f":
			selectedItem := m.projectTasksList.SelectedItem()
			if selectedItem != nil {
				task := selectedItem.(taskItem).Task
				if task.Status == models.TaskStatusOpen || task.Status == models.TaskStatusClaimed {
					m.state = viewFundBountyForm
					m.err = nil
					m.bountyInput.SetValue("")
					m.bountyInput.Focus()
					m.status = statusMessageStyle(fmt.Sprintf("Fund bounty for task '%s'", task.Title))
					return m, textinput.Blink
				} else {
					m.status = statusMessageStyle(fmt.Sprintf("Cannot fund bounty for task %d (Status: %s)", task.ID, strings.ToTitle(task.Status)))
				}
			} else {
				m.status = statusMessageStyle("Select a task first to fund its bounty.")
			}
			return m, nil
		}
	}

	newProjectTasksListModel, listCmd := m.projectTasksList.Update(msg)
	m.projectTasksList = newProjectTasksListModel
	cmds = append(cmds, listCmd)

	sel := m.projectTasksList.Index()
	items := m.projectTasksList.Items()
	if len(items) > 0 && sel >= 0 && sel < len(items) {
		task := items[sel].(taskItem).Task
		m.currentTask = &task
	} else {
		m.currentTask = nil
	}

	return m, tea.Batch(cmds...)
}

func (m model) viewProjectTasksView() string {
	spinnerView := ""
	if m.loading {
		spinnerView = m.spinner.View() + " Loading..."
	} else if m.err != nil {
		spinnerView = lipgloss.NewStyle().Foreground(red).Render(m.err.Error())
	}

	projectTitle := "No Project Selected"
	projectID := "N/A"
	if m.currentProject != nil {
		projectTitle = m.currentProject.Title
		projectID = fmt.Sprintf("%d", m.currentProject.ID)
	}

	header := titleStyle.Render(fmt.Sprintf(" OSM TUI - Tasks for Project: %s (ID: %s) (%s)", projectTitle, projectID, spinnerView))
	panelWidth := (m.width-appStyle.GetHorizontalFrameSize())/2 - 1
	leftPanel := lipgloss.NewStyle().
		Width(panelWidth).
		Height(m.height - appStyle.GetVerticalFrameSize() - 6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(blue).
		Padding(1).
		Render(m.projectTasksList.View())
	taskDetails := "Select a task to view details."
	if m.currentTask != nil {
		taskDetails = m.renderTaskDetails(m.currentTask)
	}
	rightPanel := lipgloss.NewStyle().
		Width(panelWidth).
		Height(m.height - appStyle.GetVerticalFrameSize() - 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(pink).
		Padding(1).
		Render(taskDetails)

	helpText := "↑/k up • ↓/j down • f fund bounty • r refresh • esc/p back to projects • q quit"
	ui := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel),
		lipgloss.NewStyle().Width(m.width-appStyle.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(m.status),
		lipgloss.NewStyle().Width(m.width-appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Foreground(gray).Render(helpText),
	)

	return ui
}