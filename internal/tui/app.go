package tui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func InitModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	filterInput := textinput.New()
	filterInput.Placeholder = "Filter tasks..."
	filterInput.Blur()
	filterInput.CharLimit = 150
	filterInput.Prompt = "Filter: "
	filterInput.ShowSuggestions = true
	filterInput.Cursor.SetMode(cursor.CursorStatic)

	submitInput := textinput.New()
	submitInput.Placeholder = "https://github.com/org/repo/pull/123"
	submitInput.CharLimit = 250
	submitInput.Prompt = "PR URL: "

	bountyInput := textinput.New()
	bountyInput.Placeholder = "Amount (e.g., 100.00)"
	bountyInput.CharLimit = 10
	bountyInput.Prompt = "Bounty Amount: "
	bountyInput.Validate = func(s string) error {
		if s == "" {
			return nil
		}
		_, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		return nil
	}
	bountyInput.Cursor.Style = lipgloss.NewStyle().Foreground(blue)
	bountyInput.TextStyle = formValueStyle
	bountyInput.PromptStyle = lipgloss.NewStyle().Foreground(blue)
	bountyInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	taskFormInputs := make([]textinput.Model, 7)
	for i := range taskFormInputs {
		input := textinput.New()
		input.CharLimit = 250
		input.Prompt = ""
		input.Cursor.Style = lipgloss.NewStyle().Foreground(blue)
		input.TextStyle = formValueStyle
		input.PromptStyle = lipgloss.NewStyle().Foreground(blue)
		input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		switch i {
		case 0:
			input.Placeholder = "Title"
			input.CharLimit = 100
		case 1:
			input.Placeholder = "Description"
			input.CharLimit = 500
		case 2:
			input.Placeholder = "Difficulty (easy, medium, hard)"
			input.CharLimit = 10
			input.SetValue("easy")
		case 3:
			input.Placeholder = "Estimated Hours (e.g., 8)"
			input.CharLimit = 5
			input.Validate = func(s string) error {
				if s == "" {
					return nil
				}
				_, err := strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("must be a number")
				}
				return nil
			}
		case 4:
			input.Placeholder = "Tags (JSON: [\"go\",\"cli\"])"
			input.CharLimit = 200
			input.Validate = func(s string) error {
				if s == "" {
					return nil
				}
				var dummy []string
				return json.Unmarshal([]byte(s), &dummy)
			}
		case 5:
			input.Placeholder = "Skills Required (JSON: [\"sql\",\"testing\"])"
			input.CharLimit = 200
			input.Validate = func(s string) error {
				if s == "" {
					return nil
				}
				var dummy []string
				return json.Unmarshal([]byte(s), &dummy)
			}
		case 6:
			input.Placeholder = "Bounty Amount (e.g., 50.00)"
			input.CharLimit = 10
			input.Validate = func(s string) error {
				if s == "" {
					return nil
				}
				_, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return fmt.Errorf("must be a number")
				}
				return nil
			}
		}
		taskFormInputs[i] = input
	}

	tasksList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	tasksList.Title = "OSM Tasks"
	tasksList.SetShowStatusBar(false)
	tasksList.SetFilteringEnabled(true)
	tasksList.Styles.Title = titleStyle
	tasksList.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	tasksList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(pink)
	tasksList.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(4)
	tasksList.Styles.HelpStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("241"))

	projectsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	projectsList.Title = "My Projects"
	projectsList.SetShowStatusBar(false)
	projectsList.SetFilteringEnabled(true)
	projectsList.Styles.Title = titleStyle
	projectsList.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	projectsList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(green)
	projectsList.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(4)
	projectsList.Styles.HelpStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("241"))

	projectTasksList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	projectTasksList.Title = "Project Tasks"
	projectTasksList.SetShowStatusBar(false)
	projectTasksList.SetFilteringEnabled(true)
	projectTasksList.Styles.Title = titleStyle
	projectTasksList.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	projectTasksList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(blue)
	projectTasksList.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(4)
	projectTasksList.Styles.HelpStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("241"))

	return model{
		state:            viewTasks,
		tasksList:        tasksList,
		projectsList:     projectsList,
		projectTasksList: projectTasksList,
		spinner:          s,
		status:           "Initializing...",
		apiClient:        NewAPIClient("http://localhost:8080"),
		loading:          true,
		filterInput:      filterInput,
		submitInput:      submitInput,
		taskFormInputs:   taskFormInputs,
		taskFormFocused:  0,
		bountyInput:      bountyInput,
	}
}