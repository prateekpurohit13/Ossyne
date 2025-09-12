package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"ossyne/internal/models"
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

			createTask(uint(projectID), title, description, difficulty, estimatedHours, tags, skillsRequired, bountyAmount)
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
			userIDStr, _ := cmd.Flags().GetString("user-id")

			if userIDStr == "" {
				fmt.Println("Error: --user-id flag is required to claim a task.")
				return
			}

			taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid task ID: %v\n", err)
				return
			}
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid user ID: %v\n", err)
				return
			}

			claimTask(uint(taskID), uint(userID))
		},
	}
	claimCmd.Flags().StringP("user-id", "u", "", "ID of the user claiming the task")
	claimCmd.MarkFlagRequired("user-id")
	taskCmd.AddCommand(claimCmd)

	submitCmd := &cobra.Command{
		Use:   "submit [task-id]",
		Short: "Submit a PR for a claimed task",
		Long:  `Submit a Pull Request URL for a task you have claimed.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskIDStr := args[0]
			userIDStr, _ := cmd.Flags().GetString("user-id")
			prURL, _ := cmd.Flags().GetString("pr-url")

			if userIDStr == "" || prURL == "" {
				fmt.Println("Error: --user-id and --pr-url flags are required.")
				return
			}

			taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid task ID: %v\n", err)
				return
			}
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid user ID: %v\n", err)
				return
			}

			submitContribution(uint(taskID), uint(userID), prURL)
		},
	}
	submitCmd.Flags().StringP("user-id", "u", "", "ID of the user submitting the PR")
	submitCmd.Flags().String("pr-url", "", "Full URL of the Pull Request (e.g., https://github.com/org/repo/pull/123)")
	submitCmd.MarkFlagRequired("user-id")
	submitCmd.MarkFlagRequired("pr-url")
	taskCmd.AddCommand(submitCmd)

	return taskCmd
}

func createTask(projectID uint, title, description, difficulty string, estimatedHours int, tags, skills []string, bounty float64) {
	const serverURL = "http://localhost:8080/tasks"

	payloadMap := map[string]interface{}{
		"project_id":       projectID,
		"title":            title,
		"description":      description,
		"difficulty_level": difficulty,
		"estimated_hours":  estimatedHours,
		"bounty_amount":    bounty,
	}
	if len(tags) > 0 {
		payloadMap["tags"] = tags
	}
	if len(skills) > 0 {
		payloadMap["skills_required"] = skills
	}

	payload, err := json.Marshal(payloadMap)
	if err != nil {
		fmt.Printf("Error creating request payload: %v\n", err)
		return
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", serverURL)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("Task created successfully!")
	} else {
		fmt.Printf("Error: Failed to create task (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
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

func claimTask(taskID, userID uint) {
	const serverURL = "http://localhost:8080/claims"
	payload, err := json.Marshal(map[string]uint{
		"task_id": taskID,
		"user_id": userID,
	})
	if err != nil {
		fmt.Printf("Error creating request payload: %v\n", err)
		return
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", serverURL)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("Task claimed successfully!")
	} else {
		fmt.Printf("Error: Failed to claim task (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func submitContribution(taskID, userID uint, prURL string) {
	const serverURL = "http://localhost:8080/contributions"

	payload, err := json.Marshal(map[string]interface{}{
		"task_id": taskID,
		"user_id": userID,
		"pr_url":  prURL,
	})
	if err != nil {
		fmt.Printf("Error creating request payload: %v\n", err)
		return
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", serverURL)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("Contribution submitted successfully!")
	} else {
		fmt.Printf("Error: Failed to submit contribution (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}