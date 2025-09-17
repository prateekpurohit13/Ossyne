package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ossyne/internal/models"
	"github.com/spf13/cobra"
)

func NewProjectCmd() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
		Long:  `The project command lets you create, view, and manage open-source projects.`,
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long:  `Create a new open-source project in the OSM marketplace.`,
		Run: func(cmd *cobra.Command, args []string) {
			title, _ := cmd.Flags().GetString("title")
			repoURL, _ := cmd.Flags().GetString("repo-url")
			shortDesc, _ := cmd.Flags().GetString("short-desc")
			tagsStr, _ := cmd.Flags().GetString("tags")

			if title == "" {
				fmt.Println("Error: --title flag is required.")
				return
			}

			var tags []string
			if tagsStr != "" {
				err := json.Unmarshal([]byte(tagsStr), &tags)
				if err != nil {
					fmt.Printf("Error: Invalid tags JSON: %v. Please use format `[\"tag1\",\"tag2\"]`\n", err)
					return
				}
			}

			apiClient := NewAPIClient()
			payloadMap := map[string]interface{}{
				"title":      title,
				"short_desc": shortDesc,
				"repo_url":   repoURL,
			}
			if len(tags) > 0 {
				payloadMap["tags"] = tags
			}

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPost, "/projects", payloadMap)
			if err != nil {
				fmt.Printf("Error creating project: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error creating project: %s\n", string(body))
				return
			}

			fmt.Println("Project created successfully!")
		},
	}
	createCmd.Flags().StringP("title", "t", "", "Title of the project")
	createCmd.Flags().StringP("repo-url", "r", "", "GitHub/GitLab repository URL (optional)")
	createCmd.Flags().StringP("short-desc", "d", "", "Short description of the project (optional)")
	createCmd.Flags().String("tags", "", "JSON array of tags, e.g., '[\"go\",\"cli\"]' (optional)")
	createCmd.MarkFlagRequired("title")
	projectCmd.AddCommand(createCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Long:  `List all projects available in the OSM marketplace.`,
		Run: func(cmd *cobra.Command, args []string) {
			listProjects()
		},
	}
	projectCmd.AddCommand(listCmd)

	return projectCmd
}

func listProjects() {
	const serverURL = "http://localhost:8080/projects"

	resp, err := http.Get(serverURL)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", serverURL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Failed to list projects (Status: %s)\n", resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}

	var projects []models.Project
	if err := json.Unmarshal(body, &projects); err != nil {
		fmt.Printf("Error parsing server response: %v\n", err)
		return
	}

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return
	}

	fmt.Println("--- Projects ---")
	for _, p := range projects {
		fmt.Printf("ID: %d, Title: %s, Owner: %d, Repo: %s\n", p.ID, p.Title, p.OwnerID, p.RepoURL)
	}
}