package tui

import (
	"fmt"
	"ossyne/internal/models"
	"strings"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type taskItem struct {
	models.Task
}

func (i taskItem) FilterValue() string { return i.Task.Title }
func (i taskItem) Title() string       { return i.Task.Title }
func (i taskItem) Description() string {
	bounty := ""
	if i.BountyAmount > 0 {
		bounty = fmt.Sprintf(" ($%.2f)", i.BountyAmount)
	}
	return fmt.Sprintf("Project ID: %d | Status: %s%s", i.ProjectID, strings.ToTitle(i.Status), bounty)
}

func (m model) updateTasksView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.err
		m.status = statusMessageStyle(fmt.Sprintf("Error: %v", msg.err))
		m.loading = false
		return m, nil

	case []models.Task:
		items := make([]list.Item, len(msg))
		for i, task := range msg {
			items[i] = taskItem{task}
		}
		m.tasksList.SetItems(items)
		m.status = statusMessageStyle(fmt.Sprintf("Fetched %d tasks.", len(msg)))
		m.loading = false
		if m.currentTask != nil {
			found := false
			for i, item := range items {
				if item.(taskItem).ID == m.currentTask.ID {
					m.tasksList.Select(i)
					task := item.(taskItem).Task
					m.currentTask = &task
					found = true
					break
				}
			}
			if !found && len(items) > 0 {
				m.tasksList.Select(0)
				task := items[0].(taskItem).Task
				m.currentTask = &task
			} else if len(items) == 0 {
				m.currentTask = nil
			}
		} else if len(items) > 0 {
			m.tasksList.Select(0)
			task := items[0].(taskItem).Task
			m.currentTask = &task
		} else {
			m.currentTask = nil
		}
		return m, nil

	case taskClaimedMsg:
		m.loading = true
		m.status = statusMessageStyle(fmt.Sprintf("Task %d claimed! Refreshing list...", msg.taskID))
		return m, m.apiClient.fetchTasksCmd()

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch keypress := msg.String(); keypress {
		case "enter":
			selectedItem := m.tasksList.SelectedItem()
			if selectedItem != nil {
				task := selectedItem.(taskItem).Task
				m.currentTask = &task
				m.status = statusMessageStyle(fmt.Sprintf("Selected task ID: %d", m.currentTask.ID))
			} else {
				m.status = statusMessageStyle("No task selected")
			}
			return m, nil

		case "/":
			if !m.filterInput.Focused() {
				m.filterInput.Focus()
				return m, textinput.Blink
			}

		case "esc":
			if m.filterInput.Focused() {
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.tasksList.FilterInput.SetValue("")
				return m, nil
			} else {
				m.state = viewLanding
				m.status = statusMessageStyle("Returned to landing page.")
				return m, nil
			}

		case "c":
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("You must be logged in to claim a task.")
				m.state = viewAuth
				return m, nil
			}
			selectedItem := m.tasksList.SelectedItem()
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

		case "s":
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("You must be logged in to submit contributions.")
				m.state = viewAuth
				return m, nil
			}
			selectedItem := m.tasksList.SelectedItem()
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

		case "p":
			if !m.filterInput.Focused() {
				m.state = viewProjects
				m.loading = true
				m.status = statusMessageStyle("Loading projects...")
				return m, m.apiClient.fetchUserProjectsCmd(1)
			}
			return m, nil

		case "r":
			m.loading = true
			m.status = statusMessageStyle("Refreshing tasks...")
			return m, m.apiClient.fetchTasksCmd()
		}
	}

	newListModel, listCmd := m.tasksList.Update(msg)
	m.tasksList = newListModel
	cmds = append(cmds, listCmd)

	newFilterInputModel, filterInputCmd := m.filterInput.Update(msg)
	m.filterInput = newFilterInputModel
	m.tasksList.FilterInput.SetValue(m.filterInput.Value())
	cmds = append(cmds, filterInputCmd)

	sel := m.tasksList.Index()
	items := m.tasksList.Items()
	if len(items) > 0 && sel >= 0 && sel < len(items) {
		task := items[sel].(taskItem).Task
		m.currentTask = &task
	} else {
		m.currentTask = nil
	}

	return m, tea.Batch(cmds...)
}

// viewTasksView renders the main task list view.
func (m model) viewTasksView() string {
	spinnerView := ""
	if m.loading {
		spinnerView = m.spinner.View() + " Loading..."
	} else if m.err != nil {
		spinnerView = lipgloss.NewStyle().Foreground(red).Render(m.err.Error())
	}

	header := titleStyle.Render(fmt.Sprintf(" OSM TUI - Tasks (%s)", spinnerView))

	panelWidth := (m.width-appStyle.GetHorizontalFrameSize())/2 - 1

	leftPanel := lipgloss.NewStyle().
		Width(panelWidth).
		Height(m.height - appStyle.GetVerticalFrameSize() - textInputHeight - 6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(blue).
		Padding(1).
		Render(lipgloss.JoinVertical(lipgloss.Top,
			m.filterInput.View(),
			m.tasksList.View(),
		))

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
	if m.loggedInUser == nil {
		authStatus = " (login required)"
	}
	helpText := fmt.Sprintf("↑/k up • ↓/j down • / filter • c claim%s • s submit%s • p projects • r refresh • esc back", authStatus, authStatus)
	ui := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel),
		lipgloss.NewStyle().Width(m.width-appStyle.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(m.status),
		lipgloss.NewStyle().Width(m.width-appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Foreground(gray).Render(helpText),
	)
	return ui
}

// renderTaskDetails formats the selected task's details.
func (m model) renderTaskDetails(task *models.Task) string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(blue).Render("Task Details:\n"))
	sb.WriteString(fmt.Sprintf("ID: %d\n", task.ID))
	sb.WriteString(fmt.Sprintf("Title: %s\n", lipgloss.NewStyle().Foreground(pink).Render(task.Title)))
	sb.WriteString(fmt.Sprintf("Description: %s\n", task.Description))
	sb.WriteString(fmt.Sprintf("Project ID: %d\n", task.ProjectID))
	sb.WriteString(fmt.Sprintf("Status: %s\n", lipgloss.NewStyle().Foreground(blue).Render(strings.ToTitle(task.Status))))
	if task.BountyAmount > 0 {
		sb.WriteString(fmt.Sprintf("Bounty: $%.2f\n", task.BountyAmount))
	}
	sb.WriteString(fmt.Sprintf("Difficulty: %s\n", strings.ToTitle(task.DifficultyLevel)))
	if task.EstimatedHours > 0 {
		sb.WriteString(fmt.Sprintf("Estimated Hours: %d\n", task.EstimatedHours))
	}
	if len(task.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(task.Tags, ", ")))
	}
	if len(task.SkillsRequired) > 0 {
		sb.WriteString(fmt.Sprintf("Skills Required: %s\n", strings.Join(task.SkillsRequired, ", ")))
	}
	sb.WriteString(fmt.Sprintf("Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04")))

	return sb.String()
}