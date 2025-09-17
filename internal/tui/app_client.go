package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ossyne/internal/models"
	"ossyne/internal/services"
	"strconv"
	"time"
	tea "github.com/charmbracelet/bubbletea"
)

type APIClient struct {
	BaseURL string
	Client  *http.Client
	Keyring services.KeyringService
}

func checkLoginStatusCmd() tea.Cmd {
	return func() tea.Msg {
		keyring := services.KeyringService{}
		token, err := keyring.GetToken()
		if err != nil {
			return notLoggedInMsg{}
		}
		// If token exists, fetch the user profile
		apiClient := NewAPIClient("http://localhost:8080")
		return apiClient.fetchMeCmd(token)()
	}
}

func (c *APIClient) fetchMeCmd(token string) tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/users/me", nil)
		if err != nil {
			return errMsg{fmt.Errorf("failed to create /me request: %w", err)}
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.Client.Do(req)
		if err != nil {
			return errMsg{fmt.Errorf("failed to connect to server for /me: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			keyringService := services.KeyringService{}
			keyringService.ClearToken()
			return notLoggedInMsg{}
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("API error for /me: %s (%s)", resp.Status, string(body))}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg{fmt.Errorf("failed to read /me response body: %w", err)}
		}

		var user models.User
		if err := json.Unmarshal(body, &user); err != nil {
			return errMsg{fmt.Errorf("failed to unmarshal user from /me: %w", err)}
		}
		return userFetchedMsg{user: &user}
	}
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
		Keyring: services.KeyringService{},
	}
}

func (c *APIClient) DoAuthenticatedRequest(method, endpoint string, payload interface{}) (*http.Response, error) {
	token, err := c.Keyring.GetToken()
	if err != nil {
		return nil, fmt.Errorf("not authenticated: %w", err)
	}

	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return c.Client.Do(req)
}

func (c *APIClient) submitContributionCmd(taskID, userID uint, prURL string) tea.Cmd {
	return func() tea.Msg {
		payloadMap := map[string]interface{}{
			"task_id": taskID,
			"user_id": userID,
			"pr_url":  prURL,
		}

		resp, err := c.DoAuthenticatedRequest(http.MethodPost, "/api/contributions", payloadMap)
		if err != nil {
			return errMsg{fmt.Errorf("failed to submit contribution: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return errMsg{fmt.Errorf("authentication required to submit contributions")}
		}
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("API error: %s (%s)", resp.Status, string(body))}
		}

		return contributionSubmittedMsg{taskID: taskID}
	}
}

func (c *APIClient) claimTaskCmd(taskID, userID uint) tea.Cmd {
	return func() tea.Msg {
		payloadMap := map[string]uint{
			"task_id": taskID,
			"user_id": userID,
		}

		resp, err := c.DoAuthenticatedRequest(http.MethodPost, "/api/claims", payloadMap)
		if err != nil {
			return errMsg{fmt.Errorf("failed to claim task: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return errMsg{fmt.Errorf("authentication required to claim tasks")}
		}
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("API error: %s (%s)", resp.Status, string(body))}
		}

		return taskClaimedMsg{taskID: taskID}
	}
}

func (c *APIClient) fetchTasksCmd() tea.Cmd {
	return func() tea.Msg {
		resp, err := c.Client.Get(c.BaseURL + "/tasks")
		if err != nil {
			return errMsg{fmt.Errorf("failed to connect to server: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("API error: %s (%s)", resp.Status, string(body))}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg{fmt.Errorf("failed to read response body: %w", err)}
		}

		var tasks []models.Task
		if err := json.Unmarshal(body, &tasks); err != nil {
			return errMsg{fmt.Errorf("failed to unmarshal tasks: %w", err)}
		}
		return tasks
	}
}

func (c *APIClient) fetchProjectTasksCmd(projectID uint) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("%s/tasks?project_id=%d", c.BaseURL, projectID)
		resp, err := c.Client.Get(url)
		if err != nil {
			return errMsg{fmt.Errorf("failed to connect to server: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("API error: %s (%s)", resp.Status, string(body))}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg{fmt.Errorf("failed to read response body: %w", err)}
		}

		var tasks []models.Task
		if err := json.Unmarshal(body, &tasks); err != nil {
			return errMsg{fmt.Errorf("failed to unmarshal project tasks: %w", err)}
		}
		return tasks
	}
}

func (c *APIClient) fundBountyCmd(taskID, funderUserID uint, amount float64, currency string) tea.Cmd {
	return func() tea.Msg {
		payloadMap := map[string]interface{}{
			"task_id":        taskID,
			"funder_user_id": funderUserID,
			"amount":         amount,
			"currency":       currency,
		}

		resp, err := c.DoAuthenticatedRequest(http.MethodPost, "/api/bounties/fund", payloadMap)
		if err != nil {
			return errMsg{fmt.Errorf("failed to fund bounty: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return errMsg{fmt.Errorf("authentication required to fund bounties")}
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("API error: %s (%s)", resp.Status, string(body))}
		}

		return bountyFundedMsg{taskID: taskID, amount: amount}
	}
}

func (c *APIClient) fetchUserProjectsCmd(userID uint) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("%s/users/%d/projects", c.BaseURL, userID)
		resp, err := c.Client.Get(url)
		if err != nil {
			return errMsg{fmt.Errorf("failed to connect to server: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("API error: %s (%s)", resp.Status, string(body))}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg{fmt.Errorf("failed to read response body: %w", err)}
		}

		var projects []models.Project
		if err := json.Unmarshal(body, &projects); err != nil {
			return errMsg{fmt.Errorf("failed to unmarshal projects: %w", err)}
		}
		return projects
	}
}

func (c *APIClient) createTaskFormCmd(
	projectID uint,
	title, description, difficulty string,
	estimatedHoursStr, tagsStr, skillsStr, bountyAmountStr string,
) tea.Cmd {
	return func() tea.Msg {
		if title == "" {
			return errMsg{fmt.Errorf("title cannot be empty")}
		}
		if projectID == 0 {
			return errMsg{fmt.Errorf("internal error: project ID not set for task creation")}
		}

		estimatedHours := 0
		if estimatedHoursStr != "" {
			val, err := strconv.Atoi(estimatedHoursStr)
			if err != nil {
				return errMsg{fmt.Errorf("invalid estimated hours: must be a number")}
			}
			estimatedHours = val
		}

		bountyAmount := 0.00
		if bountyAmountStr != "" {
			val, err := strconv.ParseFloat(bountyAmountStr, 64)
			if err != nil {
				return errMsg{fmt.Errorf("invalid bounty amount: must be a number")}
			}
			bountyAmount = val
		}

		var tags models.JSONStringSlice
		if tagsStr != "" {
			if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
				return errMsg{fmt.Errorf("invalid tags JSON format")}
			}
		}

		var skillsRequired models.JSONStringSlice
		if skillsStr != "" {
			if err := json.Unmarshal([]byte(skillsStr), &skillsRequired); err != nil {
				return errMsg{fmt.Errorf("invalid skills JSON format: '%s' (%v)", skillsStr, err)}
			}
		}

		payloadMap := map[string]interface{}{
			"project_id":       projectID,
			"title":            title,
			"description":      description,
			"difficulty_level": difficulty,
			"estimated_hours":  estimatedHours,
			"tags":             tags,
			"skills_required":  skillsRequired,
			"bounty_amount":    bountyAmount,
		}

		resp, err := c.DoAuthenticatedRequest(http.MethodPost, "/api/tasks", payloadMap)
		if err != nil {
			return errMsg{fmt.Errorf("failed to create task: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return errMsg{fmt.Errorf("authentication required to create tasks")}
		}
		if resp.StatusCode != http.StatusCreated {
			bodyContent, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("API error: %s (%s)", resp.Status, string(bodyContent))}
		}

		bodyContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg{fmt.Errorf("failed to read response body: %w", err)}
		}

		var createdTask models.Task
		if err := json.Unmarshal(bodyContent, &createdTask); err != nil {
			return errMsg{fmt.Errorf("failed to unmarshal created task response: %w", err)}
		}

		return taskCreatedMsg{taskID: createdTask.ID}
	}
}

func (c *APIClient) createProjectCmd(title, shortDesc, repoURL, visibility string, tags []string) tea.Cmd {
	return func() tea.Msg {
		keyring := services.KeyringService{}
		token, err := keyring.GetToken()
		if err != nil {
			return errMsg{fmt.Errorf("authentication required to create project: %w", err)}
		}

		project := models.Project{
			Title:      title,
			ShortDesc:  shortDesc,
			RepoURL:    repoURL,
			Visibility: visibility,
			Tags:       tags,
		}

		projectJSON, err := json.Marshal(project)
		if err != nil {
			return errMsg{fmt.Errorf("failed to marshal project: %w", err)}
		}

		req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/projects", bytes.NewBuffer(projectJSON))
		if err != nil {
			return errMsg{fmt.Errorf("failed to create project request: %w", err)}
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.Client.Do(req)
		if err != nil {
			return errMsg{fmt.Errorf("failed to create project: %w", err)}
		}
		defer resp.Body.Close()

		bodyContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg{fmt.Errorf("failed to read create project response: %w", err)}
		}

		if resp.StatusCode != http.StatusCreated {
			return errMsg{fmt.Errorf("failed to create project: %s", string(bodyContent))}
		}

		var createdProject models.Project
		if err := json.Unmarshal(bodyContent, &createdProject); err != nil {
			return errMsg{fmt.Errorf("failed to unmarshal created project response: %w", err)}
		}

		return projectCreatedMsg{projectID: createdProject.ID}
	}
}