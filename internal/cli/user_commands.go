package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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