package tui

import (
	"fmt"
	"ossyne/internal/models"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	state            viewState
	tasksList        list.Model
	projectsList     list.Model
	projectTasksList list.Model
	currentTask      *models.Task
	currentProject   *models.Project
	spinner          spinner.Model
	status           string
	err              error
	width            int
	height           int
	apiClient        *APIClient
	loading          bool
	filterInput      textinput.Model
	submitInput      textinput.Model
	taskFormInputs   []textinput.Model
	taskFormFocused  int
	bountyInput      textinput.Model
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.apiClient.fetchTasksCmd(),
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listWidth := max(10, m.width/2-4)
		listHeight := max(5, m.height-appStyle.GetVerticalFrameSize()-6)
		m.tasksList.SetSize(listWidth, max(5, listHeight-textInputHeight))
		m.projectsList.SetSize(listWidth, listHeight)
		m.projectTasksList.SetSize(listWidth, listHeight)
		m.filterInput.Width = listWidth
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case contributionSubmittedMsg:
		m.state = viewTasks
		m.submitInput.SetValue("")
		m.submitInput.Blur()
		m.filterInput.Blur()
		sel := m.tasksList.Index()
		m.tasksList.Select(sel)
		if items := m.tasksList.Items(); len(items) > 0 && sel >= 0 && sel < len(items) {
			task := items[sel].(taskItem).Task
			m.currentTask = &task
		}
		m.loading = true
		m.status = statusMessageStyle(fmt.Sprintf("Contribution for task %d submitted! Refreshing...", msg.taskID))
		return m, m.apiClient.fetchTasksCmd()

	case taskCreatedMsg:
		m.state = viewProjects
		for i := 0; i < len(m.taskFormInputs); i++ {
			m.taskFormInputs[i].SetValue("")
			m.taskFormInputs[i].Blur()
		}
		m.taskFormFocused = 0
		m.loading = true
		m.status = statusMessageStyle(fmt.Sprintf("Task %d created successfully! Refreshing projects...", msg.taskID))
		const maintainerID = 1
		return m, m.apiClient.fetchUserProjectsCmd(maintainerID)

	case bountyFundedMsg:
		m.state = viewProjectTasks
		m.bountyInput.SetValue("")
		m.bountyInput.Blur()
		m.loading = true
		m.status = statusMessageStyle(fmt.Sprintf("Bounty $%.2f funded for task %d! Refreshing tasks...", msg.amount, msg.taskID))
		if m.currentProject != nil {
			return m, m.apiClient.fetchProjectTasksCmd(m.currentProject.ID)
		}
		return m, m.apiClient.fetchTasksCmd()

	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
		}
		return m, cmd

	case []models.Project:
		items := make([]list.Item, len(msg))
		for i, project := range msg {
			items[i] = projectItem{project}
		}
		m.projectsList.SetItems(items)
		m.status = statusMessageStyle(fmt.Sprintf("Fetched %d projects.", len(msg)))
		m.loading = false
		if len(items) > 0 {
			project := items[0].(projectItem).Project
			m.projectsList.Select(0)
			m.currentProject = &project
		} else {
			m.currentProject = nil
		}
		return m, nil

	case []models.Task:
		if m.state == viewProjects && m.currentProject != nil {
			m.state = viewProjectTasks
			items := make([]list.Item, len(msg))
			for i, task := range msg {
				items[i] = taskItem{task}
			}
			m.projectTasksList.SetItems(items)
			m.status = statusMessageStyle(fmt.Sprintf("Fetched %d tasks for project %d.", len(msg), m.currentProject.ID))
			m.loading = false
			if len(items) > 0 {
				m.projectTasksList.Select(0)
				task := items[0].(taskItem).Task
				m.currentTask = &task
			} else {
				m.currentTask = nil
			}
			return m, nil
		}
		if m.state == viewProjectTasks {
			items := make([]list.Item, len(msg))
			for i, task := range msg {
				items[i] = taskItem{task}
			}
			m.projectTasksList.SetItems(items)
			m.status = statusMessageStyle(fmt.Sprintf("Fetched %d tasks for project %d.", len(msg), m.currentProject.ID))
			m.loading = false
			if len(items) > 0 {
				m.projectTasksList.Select(0)
				task := items[0].(taskItem).Task
				m.currentTask = &task
			} else {
				m.currentTask = nil
			}
			return m, nil
		}

		items := make([]list.Item, len(msg))
		for i, task := range msg {
			items[i] = taskItem{task}
		}
		m.tasksList.SetItems(items)
		m.status = statusMessageStyle(fmt.Sprintf("Fetched %d tasks.", len(msg)))
		m.loading = false
		if len(items) > 0 {
			task := items[0].(taskItem).Task
			m.tasksList.Select(0)
			m.currentTask = &task
		} else {
			m.currentTask = nil
		}
		return m, nil
	}

	switch m.state {
	case viewTasks:
		return m.updateTasksView(msg)
	case viewSubmit:
		return m.updateSubmitView(msg)
	case viewProjects:
		return m.updateProjectsView(msg)
	case viewCreateTask:
		return m.updateCreateTaskView(msg)
	case viewProjectTasks:
		return m.updateProjectTasksView(msg)
	case viewFundBountyForm:
		return m.updateFundBountyFormView(msg)
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}
	switch m.state {
	case viewSubmit:
		return appStyle.Render(m.viewSubmitView())
	case viewProjects:
		return appStyle.Render(m.viewProjectsView())
	case viewCreateTask:
		return appStyle.Render(m.viewCreateTaskView())
	case viewProjectTasks:
		return appStyle.Render(m.viewProjectTasksView())
	case viewFundBountyForm:
		return appStyle.Render(m.viewFundBountyFormView())
	default:
		return appStyle.Render(m.viewTasksView())
	}
}