package tui

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) updateFundBountyFormView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.handleFundBountySubmit()

		case "esc":
			m.state = viewProjectTasks
			m.status = statusMessageStyle("Bounty funding cancelled.")
			m.bountyInput.SetValue("")
			m.bountyInput.Blur()
			m.err = nil
			return m, m.apiClient.fetchProjectTasksCmd(m.currentProject.ID)

		case "ctrl+s":
			return m.handleFundBountySubmit()
		}
	}

	m.bountyInput, cmd = m.bountyInput.Update(msg)
	return m, cmd
}

func (m model) handleFundBountySubmit() (tea.Model, tea.Cmd) {
	// Check authentication first
	if m.loggedInUser == nil {
		m.err = fmt.Errorf("you must be logged in to fund bounties")
		m.state = viewAuth
		return m, nil
	}

	amountStr := m.bountyInput.Value()
	if amountStr == "" {
		m.err = fmt.Errorf("bounty amount cannot be empty")
		m.bountyInput.Focus()
		return m, textinput.Blink
	}
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		m.err = fmt.Errorf("invalid amount: must be a number")
		m.bountyInput.Focus()
		return m, textinput.Blink
	}
	if amount <= 0 {
		m.err = fmt.Errorf("bounty amount must be positive")
		m.bountyInput.Focus()
		return m, textinput.Blink
	}

	m.err = nil

	selectedItem := m.projectTasksList.SelectedItem()
	if selectedItem == nil {
		m.err = fmt.Errorf("internal error: no task selected for bounty funding")
		return m, nil
	}
	task := selectedItem.(taskItem).Task

	if m.currentProject == nil {
		m.err = fmt.Errorf("internal error: no project context for bounty funding")
		return m, nil
	}
	if task.ProjectID != m.currentProject.ID {
		m.err = fmt.Errorf("selected task %d does not belong to current project %d", task.ID, m.currentProject.ID)
		return m, nil
	}

	if task.BountyAmount > 0 && task.BountyEscrowID != nil && *task.BountyEscrowID != "" {
		m.err = fmt.Errorf("task %d already has an active bounty of $%.2f", task.ID, task.BountyAmount)
		return m, nil
	}
	m.loading = true
	m.status = statusMessageStyle(fmt.Sprintf("Funding $%.2f bounty for task %d...", amount, task.ID))
	const funderUserID = 1
	return m, m.apiClient.fundBountyCmd(task.ID, funderUserID, amount, "USD")
}

func (m model) viewFundBountyFormView() string {
	var b strings.Builder
	taskTitle := "No Task Selected"
	if sel := m.projectTasksList.SelectedItem(); sel != nil {
		taskTitle = sel.(taskItem).Task.Title
	}

	titleContent := fmt.Sprintf(" Fund Bounty for Task: %s ", taskTitle)
	b.WriteString(titleStyle.Width(m.width - appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Render(titleContent))
	b.WriteString("\n\n")

	b.WriteString(formLabelStyle.Render(m.bountyInput.Placeholder + ":"))
	b.WriteString(m.bountyInput.View())
	if m.bountyInput.Focused() {
		b.WriteString(formValueStyle.Render(" <"))
	}
	if m.bountyInput.Err != nil {
		b.WriteString(formErrorStyle.Render(" " + m.bountyInput.Err.Error()))
	}
	b.WriteString("\n")

	b.WriteString(formHelpStyle.Render("\nEnter: submit • Ctrl+S: submit • Esc: cancel"))

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(formErrorStyle.Render(fmt.Sprintf("Submission Error: %v", m.err)))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}