package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ossyne/internal/models"
	"strconv"
	"time"
	tea "github.com/charmbracelet/bubbletea"
)

type APIClient struct {
	BaseURL string
	Client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *APIClient) submitContributionCmd(taskID, userID uint, prURL string) tea.Cmd {
	return func() tea.Msg {
		payloadMap := map[string]interface{}{
			"task_id": taskID,
			"user_id": userID,
			"pr_url":  prURL,
		}
		payload, err := json.Marshal(payloadMap)
		if err != nil {
			return errMsg{fmt.Errorf("failed to marshal submission payload: %w", err)}
		}

		resp, err := c.Client.Post(c.BaseURL+"/contributions", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return errMsg{fmt.Errorf("failed to connect to server: %w", err)}
		}
		defer resp.Body.Close()

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
		payload, err := json.Marshal(payloadMap)
		if err != nil {
			return errMsg{fmt.Errorf("failed to marshal claim payload: %w", err)}
		}

		resp, err := c.Client.Post(c.BaseURL+"/claims", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return errMsg{fmt.Errorf("failed to connect to server: %w", err)}
		}
		defer resp.Body.Close()

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
		payload, err := json.Marshal(payloadMap)
		if err != nil {
			return errMsg{fmt.Errorf("failed to marshal fund bounty payload: %w", err)}
		}

		resp, err := c.Client.Post(c.BaseURL+"/bounties/fund", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return errMsg{fmt.Errorf("failed to connect to server: %w", err)}
		}
		defer resp.Body.Close()

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
		payload, err := json.Marshal(payloadMap)
		if err != nil {
			return errMsg{fmt.Errorf("failed to marshal task creation payload: %w", err)}
		}

		resp, err := c.Client.Post(c.BaseURL+"/tasks", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return errMsg{fmt.Errorf("failed to connect to server: %w", err)}
		}
		defer resp.Body.Close()

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