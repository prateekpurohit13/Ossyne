package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ossyne/internal/models"
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
			mentorIDStr, _ := cmd.Flags().GetString("mentor-id")
			userIDStr, _ := cmd.Flags().GetString("user-id")
			relatedIDStr, _ := cmd.Flags().GetString("related-id")
			notes, _ := cmd.Flags().GetString("notes")

			if mentorIDStr == "" || userIDStr == "" || relatedIDStr == "" {
				fmt.Println("Error: --mentor-id, --user-id, and --related-id are required.")
				return
			}

			mentorID, err := strconv.ParseUint(mentorIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid mentor-id: %v\n", err)
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

			mentorEndorse(uint(mentorID), uint(userID), uint(relatedID), notes)
		},
	}
	endorseCmd.Flags().StringP("mentor-id", "m", "", "ID of the mentor performing the endorsement")
	endorseCmd.Flags().StringP("user-id", "u", "", "ID of the user being endorsed")
	endorseCmd.Flags().StringP("related-id", "r", "", "ID of the related task or claim (e.g., Task ID)")
	endorseCmd.Flags().StringP("notes", "n", "", "Optional notes for the endorsement")
	endorseCmd.MarkFlagRequired("mentor-id")
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
			acceptContribution(uint(contribID))
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
			rejectContribution(uint(contribID), reason)
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
			createSkill(name, description)
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
			if err != nil { fmt.Printf("Error: Invalid user-id: %v\n", err); return }
			skillID, err := strconv.ParseUint(skillIDStr, 10, 64)
			if err != nil { fmt.Printf("Error: Invalid skill-id: %v\n", err); return }

			addUserSkill(uint(userID), uint(skillID), level)
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
			listSkills()
		},
	}
	adminCmd.AddCommand(listSkillsCmd)

	return adminCmd
}

func mentorEndorse(mentorID, userID, relatedID uint, notes string) {
	const serverURL = "http://localhost:8080/mentor/endorse"

	payload, err := json.Marshal(map[string]interface{}{
		"mentor_id":  mentorID,
		"user_id":    userID,
		"related_id": relatedID,
		"notes":      notes,
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
	if resp.StatusCode == http.StatusOK {
		fmt.Println("User endorsed successfully!")
	} else {
		fmt.Printf("Error: Failed to endorse user (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func acceptContribution(contributionID uint) {
	url := fmt.Sprintf("http://localhost:8080/contributions/%d/accept", contributionID)

	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", url)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Contribution accepted successfully!")
	} else {
		fmt.Printf("Error: Failed to accept contribution (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func rejectContribution(contributionID uint, reason string) {
	url := fmt.Sprintf("http://localhost:8080/contributions/%d/reject", contributionID)

	payload, err := json.Marshal(map[string]string{
		"reason": reason,
	})
	if err != nil {
		fmt.Printf("Error creating request payload: %v\n", err)
		return
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", url)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Contribution rejected successfully!")
	} else {
		fmt.Printf("Error: Failed to reject contribution (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func createSkill(name, description string) {
	const serverURL = "http://localhost:8080/skills"

	payload, err := json.Marshal(map[string]string{
		"name":        name,
		"description": description,
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
		fmt.Println("Skill created successfully!")
	} else {
		fmt.Printf("Error: Failed to create skill (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func listSkills() {
	const serverURL = "http://localhost:8080/skills"

	resp, err := http.Get(serverURL)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", serverURL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Failed to list skills (Status: %s)\n", resp.Status)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}
	var skills []models.Skill
	if err := json.Unmarshal(body, &skills); err != nil {
		fmt.Printf("Error parsing server response: %v\n", err)
		return
	}
	if len(skills) == 0 {
		fmt.Println("No skills found.")
		return
	}

	fmt.Println("--- Skills ---")
	for _, s := range skills {
		fmt.Printf("ID: %d, Name: %s, Description: %s\n", s.ID, s.Name, s.Description)
	}
}

func addUserSkill(userID, skillID uint, level string) {
	const serverURL = "http://localhost:8080/users/skills"

	payload, err := json.Marshal(map[string]interface{}{
		"user_id":  userID,
		"skill_id": skillID,
		"level":    level,
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
		fmt.Println("User skill added successfully!")
	} else {
		fmt.Printf("Error: Failed to add user skill (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}