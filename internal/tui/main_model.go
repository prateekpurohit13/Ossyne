package tui

import (
	"fmt"
	"ossyne/internal/models"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	state              viewState
	previousView       viewState
	loggedInUser       *models.User
	tasksList          list.Model
	projectsList       list.Model
	projectTasksList   list.Model
	currentTask        *models.Task
	currentProject     *models.Project
	spinner            spinner.Model
	status             string
	err                error
	width              int
	height             int
	apiClient          *APIClient
	loading            bool
	filterInput        textinput.Model
	submitInput        textinput.Model
	taskFormInputs     []textinput.Model
	taskFormFocused    int
	bountyInput        textinput.Model
	createProjectForm  createProjectFormModel
	forceCompactBanner bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		checkLoginStatusCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listWidth := max(10, m.width/3-4)
		listHeight := max(5, m.height-appStyle.GetVerticalFrameSize()-6)
		m.tasksList.SetSize(listWidth, max(5, listHeight-textInputHeight))
		m.projectsList.SetSize(listWidth, listHeight)
		m.projectTasksList.SetSize(listWidth, listHeight)
		m.filterInput.Width = listWidth
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case notLoggedInMsg:
		m.state = viewLanding
		m.loading = false
		m.status = statusMessageStyle("Welcome to OSSYNE! Browse projects or login for more features.")
		return m, nil

	case userFetchedMsg:
		m.loggedInUser = msg.user
		m.state = viewLanding
		m.loading = false
		m.status = statusMessageStyle(fmt.Sprintf("Welcome back, %s!", m.loggedInUser.Username))
		return m, nil

	case startLoginFlowMsg:
		fmt.Println("Exiting TUI to open browser for login.")
		fmt.Println("Please run 'go run ./cmd/ossyne-cli/main.go auth login' and then restart the TUI with 'go run ./cmd/ossyne-cli/main.go'.")
		return m, tea.Quit

	case contributionSubmittedMsg:
		if m.currentProject != nil {
			m.state = viewProjectTasks
		} else {
			m.state = viewTasks
		}
		m.submitInput.SetValue("")
		m.submitInput.Blur()
		m.filterInput.Blur()
		if m.state == viewTasks {
			sel := m.tasksList.Index()
			m.tasksList.Select(sel)
			if items := m.tasksList.Items(); len(items) > 0 && sel >= 0 && sel < len(items) {
				task := items[sel].(taskItem).Task
				m.currentTask = &task
			}
			m.loading = true
			m.status = statusMessageStyle(fmt.Sprintf("Contribution for task %d submitted! Refreshing...", msg.taskID))
			return m, m.apiClient.fetchTasksCmd()
		} else {
			m.loading = true
			m.status = statusMessageStyle(fmt.Sprintf("Contribution for task %d submitted! Refreshing...", msg.taskID))
			return m, m.apiClient.fetchProjectTasksCmd(m.currentProject.ID)
		}

	case taskCreatedMsg:
		m.state = viewManageProjects
		for i := 0; i < len(m.taskFormInputs); i++ {
			m.taskFormInputs[i].SetValue("")
			m.taskFormInputs[i].Blur()
		}
		m.taskFormFocused = 0
		m.loading = false
		m.status = statusMessageStyle(fmt.Sprintf("Task %d created successfully!", msg.taskID))
		return m, nil

	case projectCreatedMsg:
		m.state = viewManageProjects
		m.loading = true
		m.status = statusMessageStyle(fmt.Sprintf("Project %d created successfully! Refreshing...", msg.projectID))
		if m.loggedInUser != nil {
			return m, m.apiClient.fetchUserProjectsCmd(m.loggedInUser.ID)
		}
		return m, nil

	case bountyFundedMsg:
		m.state = viewManageProjects
		m.bountyInput.SetValue("")
		m.bountyInput.Blur()
		m.loading = false
		m.status = statusMessageStyle(fmt.Sprintf("Bounty $%.2f funded for task %d!", msg.amount, msg.taskID))
		return m, nil

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
			m.previousView = m.state
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
		if m.state == viewManageProjects && m.currentProject != nil {
			m.previousView = m.state
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
	case viewLanding:
		return m.updateLandingView(msg)
	case viewAuth:
		return m.updateAuthView(msg)
	case viewTasks:
		return m.updateTasksView(msg)
	case viewSubmit:
		return m.updateSubmitView(msg)
	case viewProjects:
		return m.updateProjectsView(msg)
	case viewCreateProject:
		return m.updateCreateProjectView(msg)
	case viewManageProjects:
		return m.updateManageProjectsView(msg)
	case viewCreateProjectForm:
		return m.updateCreateProjectFormView(msg)
	case viewCreateTask:
		return m.updateCreateTaskView(msg)
	case viewProjectTasks:
		return m.updateProjectTasksView(msg)
	case viewFundBountyForm:
		return m.updateFundBountyFormView(msg)
	case viewMyContributions:
		return m.updateMyContributionsView(msg)
	case viewReviewContributions:
		return m.updateReviewContributionsView(msg)
	case viewMyWallet:
		return m.updateMyWalletView(msg)
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}
	switch m.state {
	case viewLanding:
		return appStyle.Render(m.viewLandingView())
	case viewAuth:
		return appStyle.Render(m.viewAuthView())
	case viewSubmit:
		return appStyle.Render(m.viewSubmitView())
	case viewProjects:
		return appStyle.Render(m.viewProjectsView())
	case viewCreateProject:
		return appStyle.Render(m.viewCreateProjectView())
	case viewManageProjects:
		return m.viewManageProjectsView()
	case viewCreateProjectForm:
		content := m.createProjectForm.viewCreateProjectForm()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	case viewCreateTask:
		return appStyle.Render(m.viewCreateTaskView())
	case viewProjectTasks:
		return appStyle.Render(m.viewProjectTasksView())
	case viewFundBountyForm:
		return appStyle.Render(m.viewFundBountyFormView())
	case viewMyContributions:
		return appStyle.Render(m.viewMyContributionsView())
	case viewReviewContributions:
		return appStyle.Render(m.viewReviewContributionsView())
	case viewMyWallet:
		return appStyle.Render(m.viewMyWalletView())
	default:
		return appStyle.Render(m.viewTasksView())
	}
}

func (m model) updateCreateProjectFormView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = viewManageProjects
			return m, m.apiClient.fetchUserProjectsCmd(m.loggedInUser.ID)
		}
	case projectCreateSubmitMsg:
		visibility := "private"
		if msg.isPublic {
			visibility = "public"
		}
		return m, m.apiClient.createProjectCmd(
			msg.title,
			msg.description,
			msg.repositoryURL,
			visibility,
			msg.tags,
		)
	}

	m.createProjectForm, cmd = m.createProjectForm.updateCreateProjectForm(msg)
	return m, cmd
}