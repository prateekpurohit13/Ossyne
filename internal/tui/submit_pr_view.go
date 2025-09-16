package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) updateSubmitView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			prURL := m.submitInput.Value()
			if prURL != "" && m.currentTask != nil {
				m.loading = true
				m.status = statusMessageStyle("Submitting contribution...")
				const userID = 2
				return m, m.apiClient.submitContributionCmd(m.currentTask.ID, userID, prURL)
			}
		case tea.KeyEsc:
			m.state = viewTasks
			m.submitInput.Blur()
			m.filterInput.Focus()
			m.status = statusMessageStyle("Submission cancelled.")
			return m, nil
		}
	}

	m.submitInput, cmd = m.submitInput.Update(msg)
	return m, cmd
}

func (m model) viewSubmitView() string {
	return lipgloss.Place(m.width, m.height-appStyle.GetVerticalFrameSize(), lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left,
			"Enter the full Pull Request URL for your contribution.",
			m.submitInput.View(),
			lipgloss.NewStyle().Foreground(gray).Render("(Press Enter to submit, Esc to cancel)"),
		),
	)
}