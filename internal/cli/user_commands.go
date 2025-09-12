package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ossyne/internal/models"
	"strconv"
	"strings"
	"github.com/spf13/cobra"
)

func NewUserCmd() *cobra.Command {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  `The user command lets you create, view, and manage users.`,
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Long:  `Create a new user in the OSM marketplace.`,
		Run: func(cmd *cobra.Command, args []string) {
			username, _ := cmd.Flags().GetString("username")
			email, _ := cmd.Flags().GetString("email")

			if username == "" || email == "" {
				fmt.Println("Error: --username and --email flags are required.")
				return
			}

			createUser(username, email)
		},
	}
	createCmd.Flags().StringP("username", "u", "", "Username for the new user")
	createCmd.Flags().StringP("email", "e", "", "Email for the new user")
	createCmd.MarkFlagRequired("username")
	createCmd.MarkFlagRequired("email")
	userCmd.AddCommand(createCmd)
	viewCmd := &cobra.Command{
		Use:   "view [user-id]",
		Short: "View a user's profile, ratings, and skills",
		Long:  `View detailed profile information for a specific user, including their ratings and associated skills.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			userIDStr := args[0]
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid user ID: %v\n", err)
				return
			}
			viewUser(uint(userID))
		},
	}
	userCmd.AddCommand(viewCmd)

	return userCmd
}

func createUser(username, email string) {
	const serverURL = "http://localhost:8080/users"

	payload, err := json.Marshal(map[string]string{
		"username": username,
		"email":    email,
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
		fmt.Println("User created successfully!")
	} else {
		fmt.Printf("Error: Failed to create user (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func viewUser(userID uint) {
	userURL := fmt.Sprintf("http://localhost:8080/users/%d", userID)
	resp, err := http.Get(userURL)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server to fetch user %d. Is it running?\n", userID)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Failed to fetch user %d (Status: %s)\n", userID, resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body))
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading user response: %v\n", err)
		return
	}
	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		fmt.Printf("Error parsing user response: %v\n", err)
		return
	}

	fmt.Println("--- User Profile ---")
	fmt.Printf("ID: %d\n", user.ID)
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Ratings: %d\n", user.Ratings)
	fmt.Printf("Roles: %s\n", strings.Join(user.Roles, ", "))

	userSkillsURL := fmt.Sprintf("http://localhost:8080/users/%d/skills", userID)
	respSkills, err := http.Get(userSkillsURL)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server to fetch user skills for %d. Is it running?\n", userID)
		return
	}
	defer respSkills.Body.Close()
	if respSkills.StatusCode != http.StatusOK {
		fmt.Printf("Warning: Failed to fetch skills for user %d (Status: %s)\n", userID, respSkills.Status)
		return
	}
	bodySkills, err := io.ReadAll(respSkills.Body)
	if err != nil {
		fmt.Printf("Error reading user skills response: %v\n", err)
		return
	}
	var userSkills []models.UserSkill
	if err := json.Unmarshal(bodySkills, &userSkills); err != nil {
		fmt.Printf("Error parsing user skills response: %v\n", err)
		return
	}
	if len(userSkills) > 0 {
		fmt.Println("\n--- Skills ---")
		for _, us := range userSkills {
			if us.Skill.Name != "" {
				fmt.Printf("- %s (Level: %s)\n", us.Skill.Name, us.Level)
			} else {
				fmt.Printf("- Skill ID: %d (Level: %s) - Name not available\n", us.SkillID, us.Level)
			}
		}
	} else {
		fmt.Println("\nNo skills listed for this user.")
	}
}