package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ossyne/internal/models"
	"strconv"
	"strings"
	"github.com/spf13/cobra"
)

func NewTaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
		Long:  `The task command lets you create, list, claim, and submit tasks.`,
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		Long:  `Create a new task for an existing project.`,
		Run: func(cmd *cobra.Command, args []string) {
			projectIDStr, _ := cmd.Flags().GetString("project-id")
			title, _ := cmd.Flags().GetString("title")
			description, _ := cmd.Flags().GetString("description")
			difficulty, _ := cmd.Flags().GetString("difficulty")
			estimatedHoursStr, _ := cmd.Flags().GetString("estimated-hours")
			tagsStr, _ := cmd.Flags().GetString("tags")
			skillsStr, _ := cmd.Flags().GetString("skills-required")
			bountyAmountStr, _ := cmd.Flags().GetString("bounty-amount")

			if projectIDStr == "" || title == "" {
				fmt.Println("Error: --project-id and --title flags are required.")
				return
			}

			projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid project-id: %v\n", err)
				return
			}

			estimatedHours, err := strconv.Atoi(estimatedHoursStr)
			if estimatedHoursStr != "" && err != nil {
				fmt.Printf("Error: Invalid estimated-hours: %v\n", err)
				return
			}

			bountyAmount, err := strconv.ParseFloat(bountyAmountStr, 64)
			if bountyAmountStr != "" && err != nil {
				fmt.Printf("Error: Invalid bounty-amount: %v\n", err)
				return
			}

			var tags []string
			if tagsStr != "" {
				err = json.Unmarshal([]byte(tagsStr), &tags)
				if err != nil {
					fmt.Printf("Error: Invalid tags JSON: %v. Please use format `[\"tag1\",\"tag2\"]`\n", err)
					return
				}
			}

			var skillsRequired []string
			if skillsStr != "" {
				err = json.Unmarshal([]byte(skillsStr), &skillsRequired)
				if err != nil {
					fmt.Printf("Error: Invalid skills-required JSON: %v. Please use format `[\"skill1\",\"skill2\"]`\n", err)
					return
				}
			}

			apiClient := NewAPIClient()
			payloadMap := map[string]interface{}{
				"project_id":       uint(projectID),
				"title":            title,
				"description":      description,
				"difficulty_level": difficulty,
				"estimated_hours":  estimatedHours,
				"bounty_amount":    bountyAmount,
			}
			if len(tags) > 0 {
				payloadMap["tags"] = tags
			}
			if len(skillsRequired) > 0 {
				payloadMap["skills_required"] = skillsRequired
			}

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPost, "/tasks", payloadMap)
			if err != nil {
				fmt.Printf("Error creating task: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error creating task: %s\n", string(body))
				return
			}

			fmt.Println("Task created successfully!")
		},
	}
	createCmd.Flags().StringP("project-id", "p", "", "ID of the project this task belongs to")
	createCmd.Flags().StringP("title", "t", "", "Title of the task")
	createCmd.Flags().StringP("description", "d", "", "Detailed description of the task (optional)")
	createCmd.Flags().String("difficulty", "easy", "Difficulty level (easy, medium, hard)")
	createCmd.Flags().String("estimated-hours", "0", "Estimated hours to complete (optional)")
	createCmd.Flags().String("tags", "", "JSON array of tags, e.g., '[\"bug\",\"feature\"]' (optional)")
	createCmd.Flags().String("skills-required", "", "JSON array of required skills, e.g., '[\"go\",\"testing\"]' (optional)")
	createCmd.Flags().String("bounty-amount", "0.00", "Monetary bounty for completing this task (optional)")
	createCmd.MarkFlagRequired("project-id")
	createCmd.MarkFlagRequired("title")
	taskCmd.AddCommand(createCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long:  `List tasks, optionally filtered by project or status.`,
		Run: func(cmd *cobra.Command, args []string) {
			projectIDStr, _ := cmd.Flags().GetString("project-id")
			status, _ := cmd.Flags().GetString("status")
			listTasks(projectIDStr, status)
		},
	}
	listCmd.Flags().StringP("project-id", "p", "", "Filter tasks by Project ID (optional)")
	listCmd.Flags().StringP("status", "s", "", "Filter tasks by status (open, claimed, in_progress, submitted, completed, archived) (optional)")
	taskCmd.AddCommand(listCmd)

	claimCmd := &cobra.Command{
		Use:   "claim [task-id]",
		Short: "Claim an open task",
		Long:  `Claim an open task to work on.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskIDStr := args[0]
			devUserID, _ := cmd.Flags().GetString("dev-user-id")

			taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid task ID: %v\n", err)
				return
			}

			payload := map[string]uint{"task_id": uint(taskID)}

			// Development mode: bypass authentication if dev-user-id is provided
			if devUserID != "" {
				userID, err := strconv.ParseUint(devUserID, 10, 64)
				if err != nil {
					fmt.Printf("Error: Invalid dev-user-id: %v\n", err)
					return
				}
				payload["user_id"] = uint(userID)

				// Use direct HTTP request for development
				jsonData, _ := json.Marshal(payload)
				resp, err := http.Post("http://localhost:8080/dev/claims", "application/json", strings.NewReader(string(jsonData)))
				if err != nil {
					fmt.Printf("Error claiming task: %v\n", err)
					return
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading response: %v\n", err)
					return
				}

				if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
					fmt.Printf("Error claiming task: %s\n", string(body))
					return
				}

				fmt.Printf("Successfully claimed task %d as user %d (dev mode)\n", taskID, userID)
				return
			}

			apiClient := NewAPIClient()
			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPost, "/claims", payload)
			if err != nil {
				fmt.Printf("Error claiming task: %v\n", err)
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}
			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error claiming task: %s\n", string(body))
				return
			}
			fmt.Printf("Successfully claimed task %d\n", taskID)
		},
	}
	claimCmd.Flags().StringP("dev-user-id", "", "", "Development only: specify user ID to claim task as (bypasses auth)")
	taskCmd.AddCommand(claimCmd)

	submitCmd := &cobra.Command{
		Use:   "submit [task-id] --pr-url [PR-URL]",
		Short: "Submit a PR for a claimed task",
		Long:  `Submit a Pull Request URL for a task you have claimed.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskIDStr := args[0]
			prURL, _ := cmd.Flags().GetString("pr-url")
			devUserID, _ := cmd.Flags().GetString("dev-user-id")

			if prURL == "" {
				fmt.Println("Error: --pr-url flag is required.")
				return
			}

			taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid task ID: %v\n", err)
				return
			}

			payload := map[string]interface{}{
				"task_id": uint(taskID),
				"pr_url":  prURL,
			}

			// Development mode: bypass authentication if dev-user-id is provided
			if devUserID != "" {
				userID, err := strconv.ParseUint(devUserID, 10, 64)
				if err != nil {
					fmt.Printf("Error: Invalid dev-user-id: %v\n", err)
					return
				}
				payload["user_id"] = uint(userID)

				// Use direct HTTP request for development
				jsonData, _ := json.Marshal(payload)
				resp, err := http.Post("http://localhost:8080/dev/contributions", "application/json", strings.NewReader(string(jsonData)))
				if err != nil {
					fmt.Printf("Error submitting contribution: %v\n", err)
					return
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading response: %v\n", err)
					return
				}
				if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
					fmt.Printf("Error submitting contribution: %s\n", string(body))
					return
				}
				fmt.Printf("Successfully submitted contribution for task %d as user %d (dev mode)\n", taskID, userID)
				return
			}

			apiClient := NewAPIClient()
			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPost, "/contributions", payload)
			if err != nil {
				fmt.Printf("Error submitting contribution: %v\n", err)
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}
			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error submitting contribution: %s\n", string(body))
				return
			}

			fmt.Printf("Successfully submitted contribution for task %d\n", taskID)
		},
	}
	submitCmd.Flags().String("pr-url", "", "Full URL of the Pull Request (e.g., https://github.com/org/repo/pull/123)")
	submitCmd.Flags().StringP("dev-user-id", "", "", "Development only: specify user ID to submit as (bypasses auth)")
	submitCmd.MarkFlagRequired("pr-url")
	taskCmd.AddCommand(submitCmd)

	return taskCmd
}

func listTasks(projectIDStr, status string) {
	baseURL := "http://localhost:8080/tasks"
	params := []string{}
	if projectIDStr != "" {
		params = append(params, "project_id="+projectIDStr)
	}
	if status != "" {
		params = append(params, "status="+status)
	}

	url := baseURL
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", baseURL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Failed to list tasks (Status: %s)\n", resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}

	var tasks []models.Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		fmt.Printf("Error parsing server response: %v\n", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Println("--- Tasks ---")
	for _, t := range tasks {
		fmt.Printf("ID: %d, Title: %s, Project ID: %d, Status: %s, Bounty: %.2f\n",
			t.ID, t.Title, t.ProjectID, t.Status, t.BountyAmount)
	}
}