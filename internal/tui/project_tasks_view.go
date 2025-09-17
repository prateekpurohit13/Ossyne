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
			m.state = m.previousView
			if m.previousView == viewManageProjects {
				m.status = statusMessageStyle(fmt.Sprintf("Returned to project management for '%s'.", m.currentProject.Title))
			} else {
				m.status = statusMessageStyle(fmt.Sprintf("Returned to browse projects for '%s'.", m.currentProject.Title))
			}
			m.projectTasksList.FilterInput.Blur()
			m.projectTasksList.FilterInput.SetValue("")
			m.projectsList.Select(m.projectsList.Index())
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
		case "c": // Claim task (only for authenticated users)
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("Please login first to claim tasks.")
				m.state = viewAuth
				return m, nil
			}
			selectedItem := m.projectTasksList.SelectedItem()
			if selectedItem != nil {
				task := selectedItem.(taskItem).Task
				if task.Status == models.TaskStatusOpen {
					m.loading = true
					m.status = statusMessageStyle(fmt.Sprintf("Claiming task %d...", task.ID))
					return m, m.apiClient.claimTaskCmd(task.ID, m.loggedInUser.ID)
				} else {
					m.status = statusMessageStyle(fmt.Sprintf("Cannot claim task %d (Status: %s)", task.ID, strings.ToTitle(task.Status)))
				}
			} else {
				m.status = statusMessageStyle("No task selected to claim")
			}
			return m, nil

		case "s": // Submit contribution (only for authenticated users)
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("Please login first to submit contributions.")
				m.state = viewAuth
				return m, nil
			}
			selectedItem := m.projectTasksList.SelectedItem()
			if selectedItem != nil {
				task := selectedItem.(taskItem).Task
				if task.Status == models.TaskStatusClaimed || task.Status == models.TaskStatusInProgress {
					m.state = viewSubmit
					m.submitInput.Focus()
					m.submitInput.SetValue("")
					m.status = statusMessageStyle(fmt.Sprintf("Enter PR URL for task '%s'", task.Title))
					return m, textinput.Blink
				} else {
					m.status = statusMessageStyle(fmt.Sprintf("Cannot submit for task %d (Status: %s)", task.ID, strings.ToTitle(task.Status)))
				}
			} else {
				m.status = statusMessageStyle("No task selected to submit")
			}
			return m, nil

		case "f": // Fund bounty (only available for owned projects)
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("Please login first to fund bounties.")
				m.state = viewAuth
				return m, nil
			}
			if m.currentProject == nil || m.currentProject.OwnerID != m.loggedInUser.ID {
				m.status = statusMessageStyle("You can only fund bounties for your own projects.")
				return m, nil
			}
			selectedItem := m.projectTasksList.SelectedItem()
			if selectedItem != nil {
				task := selectedItem.(taskItem).Task
				m.currentTask = &task
				m.state = viewFundBountyForm
				m.bountyInput.Focus()
				m.bountyInput.SetValue("")
				m.status = statusMessageStyle(fmt.Sprintf("Enter bounty amount for task '%s'", task.Title))
				return m, textinput.Blink
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

	header := titleStyle.Render(fmt.Sprintf(" Project Tasks: %s (ID: %s) (%s)", projectTitle, projectID, spinnerView))
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

	authStatus := ""
	fundStatus := ""
	if m.loggedInUser == nil {
		authStatus = " (login required for claim/submit/fund)"
	} else if m.currentProject != nil && m.currentProject.OwnerID == m.loggedInUser.ID {
		fundStatus = " • f fund bounty"
	}
	helpText := fmt.Sprintf("↑/k up • ↓/j down • c claim • s submit%s%s • r refresh • esc back to projects", fundStatus, authStatus)
	ui := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel),
		lipgloss.NewStyle().Width(m.width-appStyle.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(m.status),
		lipgloss.NewStyle().Width(m.width-appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Foreground(gray).Render(helpText),
	)

	return ui
}