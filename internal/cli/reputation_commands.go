package cli

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"github.com/spf13/cobra"
)

func NewRatingsCmd() *cobra.Command {
	ratingsCmd := &cobra.Command{
		Use:   "ratings",
		Short: "Manage ratings and endorsements",
		Long:  `The ratings command provides tools for managing user ratings and mentor endorsements.`,
	}

	endorseCmd := &cobra.Command{
		Use:   "endorse",
		Short: "Endorse a user as a mentor",
		Long:  `A mentor can endorse a user for their work on a specific task/claim, boosting their ratings.`,
		Run: func(cmd *cobra.Command, args []string) {
			userIDStr, _ := cmd.Flags().GetString("user-id")
			relatedIDStr, _ := cmd.Flags().GetString("related-id")
			notes, _ := cmd.Flags().GetString("notes")

			if userIDStr == "" || relatedIDStr == "" {
				fmt.Println("Error: --user-id and --related-id are required.")
				return
			}

			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid user-id: %v\n", err)
				return
			}
			relatedID, err := strconv.ParseUint(relatedIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid related-id: %v\n", err)
				return
			}

			apiClient := NewAPIClient()
			payloadMap := map[string]interface{}{
				"user_id":    uint(userID),
				"related_id": uint(relatedID),
				"notes":      notes,
			}

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPost, "/mentor/endorse", payloadMap)
			if err != nil {
				fmt.Printf("Error endorsing user: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error endorsing user: %s\n", string(body))
				return
			}

			fmt.Printf("Successfully endorsed user %d\n", userID)
		},
	}
	endorseCmd.Flags().StringP("user-id", "u", "", "ID of the user being endorsed")
	endorseCmd.Flags().StringP("related-id", "r", "", "ID of the related task or claim (e.g., Task ID)")
	endorseCmd.Flags().StringP("notes", "n", "", "Optional notes for the endorsement")
	endorseCmd.MarkFlagRequired("user-id")
	endorseCmd.MarkFlagRequired("related-id")
	ratingsCmd.AddCommand(endorseCmd)

	return ratingsCmd
}

func NewAdminCmd() *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "Admin actions for platform owners",
		Long:  `The admin command provides powerful tools for platform owners to manage users, content, and resolve disputes.`,
	}

	acceptContribCmd := &cobra.Command{
		Use:   "accept-contribution [contribution-id]",
		Short: "Manually accept a contribution (e.g., after verification)",
		Long:  `Accepts a contribution, marking it as verified and crediting the contributor.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			contribIDStr := args[0]
			contribID, err := strconv.ParseUint(contribIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid contribution ID: %v\n", err)
				return
			}

			apiClient := NewAPIClient()
			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPut, fmt.Sprintf("/contributions/%d/accept", contribID), nil)
			if err != nil {
				fmt.Printf("Error accepting contribution: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error accepting contribution: %s\n", string(body))
				return
			}

			fmt.Println("Contribution accepted successfully!")
		},
	}
	adminCmd.AddCommand(acceptContribCmd)

	rejectContribCmd := &cobra.Command{
		Use:   "reject-contribution [contribution-id]",
		Short: "Reject a contribution",
		Long:  `Rejects a contribution, setting its status to rejected.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			contribIDStr := args[0]
			reason, _ := cmd.Flags().GetString("reason")

			if reason == "" {
				fmt.Println("Error: --reason flag is required to reject a contribution.")
				return
			}

			contribID, err := strconv.ParseUint(contribIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid contribution ID: %v\n", err)
				return
			}

			apiClient := NewAPIClient()
			payloadMap := map[string]interface{}{
				"reason": reason,
			}

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPut, fmt.Sprintf("/contributions/%d/reject", contribID), payloadMap)
			if err != nil {
				fmt.Printf("Error rejecting contribution: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error rejecting contribution: %s\n", string(body))
				return
			}

			fmt.Println("Contribution rejected successfully!")
		},
	}
	rejectContribCmd.Flags().StringP("reason", "r", "", "Reason for rejecting the contribution")
	rejectContribCmd.MarkFlagRequired("reason")
	adminCmd.AddCommand(rejectContribCmd)

	createSkillCmd := &cobra.Command{
		Use:   "create-skill",
		Short: "Add a new skill to the marketplace",
		Long:  `Allows platform admins to add new skills that can be associated with users and tasks.`,
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			description, _ := cmd.Flags().GetString("description")
			if name == "" {
				fmt.Println("Error: --name flag is required.")
				return
			}

			apiClient := NewAPIClient()
			payloadMap := map[string]interface{}{
				"name":        name,
				"description": description,
			}

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPost, "/admin/skills", payloadMap)
			if err != nil {
				fmt.Printf("Error creating skill: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error creating skill: %s\n", string(body))
				return
			}

			fmt.Printf("Successfully created skill '%s'\n", name)
		},
	}
	createSkillCmd.Flags().StringP("name", "n", "", "Name of the skill (e.g., 'Go', 'React', 'Documentation')")
	createSkillCmd.Flags().StringP("description", "d", "", "Description of the skill (optional)")
	createSkillCmd.MarkFlagRequired("name")
	adminCmd.AddCommand(createSkillCmd)

	addUserSkillCmd := &cobra.Command{
		Use:   "add-user-skill",
		Short: "Associate a skill with a user",
		Long:  `Platform admins can manually assign skills to users with a proficiency level.`,
		Run: func(cmd *cobra.Command, args []string) {
			userIDStr, _ := cmd.Flags().GetString("user-id")
			skillIDStr, _ := cmd.Flags().GetString("skill-id")
			level, _ := cmd.Flags().GetString("level")

			if userIDStr == "" || skillIDStr == "" {
				fmt.Println("Error: --user-id and --skill-id are required.")
				return
			}
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid user-id: %v\n", err)
				return
			}
			skillID, err := strconv.ParseUint(skillIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid skill-id: %v\n", err)
				return
			}

			apiClient := NewAPIClient()
			payloadMap := map[string]interface{}{
				"user_id":           uint(userID),
				"skill_id":          uint(skillID),
				"proficiency_level": level,
			}

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPost, "/admin/users/skills", payloadMap)
			if err != nil {
				fmt.Printf("Error adding user skill: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error adding user skill: %s\n", string(body))
				return
			}

			fmt.Printf("Successfully added skill %d to user %d with level %s\n", skillID, userID, level)
		},
	}
	addUserSkillCmd.Flags().StringP("user-id", "u", "", "ID of the user to add the skill to")
	addUserSkillCmd.Flags().StringP("skill-id", "s", "", "ID of the skill to add")
	addUserSkillCmd.Flags().StringP("level", "l", "beginner", "Proficiency level (beginner, intermediate, expert)")
	addUserSkillCmd.MarkFlagRequired("user-id")
	addUserSkillCmd.MarkFlagRequired("skill-id")
	adminCmd.AddCommand(addUserSkillCmd)

	listSkillsCmd := &cobra.Command{
		Use:   "list-skills",
		Short: "List all skills",
		Long:  `Lists all defined skills in the marketplace.`,
		Run: func(cmd *cobra.Command, args []string) {
			apiClient := NewAPIClient()
			resp, err := apiClient.DoAuthenticatedRequest(http.MethodGet, "/admin/skills", nil)
			if err != nil {
				fmt.Printf("Error listing skills: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error listing skills: %s\n", string(body))
				return
			}

			fmt.Println("Skills:")
			fmt.Println(string(body))
		},
	}
	adminCmd.AddCommand(listSkillsCmd)

	return adminCmd
}