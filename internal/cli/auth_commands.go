package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"ossyne/internal/services"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  `The auth command provides tools to log in and log out of OSSYNE.`,
	}

	keyring := services.KeyringService{}
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to OSSYNE using a provider",
		Long:  `Launches a web browser to authenticate with an external provider like GitHub.`,
		Run: func(cmd *cobra.Command, args []string) {
			provider, _ := cmd.Flags().GetString("provider")
			if provider != "github" {
				fmt.Println("Error: Currently, only 'github' is a supported provider.")
				return
			}
			tokenChan := make(chan string)
			errChan := make(chan error)			
			server := &http.Server{Addr: ":9999"}			
			http.HandleFunc("/auth/cli/callback", func(w http.ResponseWriter, r *http.Request) {
				token := r.URL.Query().Get("token")
				if token == "" {
					errChan <- fmt.Errorf("did not receive token in callback")
					return
				}
				tokenChan <- token
				fmt.Fprintln(w, "Authentication successful! You can close this browser window now.")
			})
			
			go func() {
				if err := server.ListenAndServe(); err != http.ErrServerClosed {
					errChan <- fmt.Errorf("failed to start local callback server: %w", err)
				}
			}()

			const remoteAuthURL = "http://localhost:8080/auth/github"
			fmt.Println("Your browser should open for authentication.")
			fmt.Printf("If it doesn't, please navigate to this URL: %s\n", remoteAuthURL)
			err := browser.OpenURL(remoteAuthURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening browser: %v\n", err)
			}
			
			select {
			case token := <-tokenChan:
				fmt.Println("Successfully received access token.")
				if err := keyring.SetToken(token); err != nil {
					fmt.Fprintf(os.Stderr, "Error storing token: %v\n", err)
				} else {
					fmt.Println("Authentication successful. You are now logged in.")
				}
			case err := <-errChan:
				fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
			case <-time.After(2 * time.Minute):
				fmt.Fprintf(os.Stderr, "Authentication timed out.\n")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		},
	}
	loginCmd.Flags().StringP("provider", "p", "github", "The authentication provider to use (e.g., 'github')")
	authCmd.AddCommand(loginCmd)

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of OSSYNE",
		Run: func(cmd *cobra.Command, args []string) {
			if err := keyring.ClearToken(); err != nil {
				fmt.Fprintf(os.Stderr, "Error logging out: %v\n", err)
			} else {
				fmt.Println("You have been logged out.")
			}
		},
	}
	authCmd.AddCommand(logoutCmd)

	return authCmd
}