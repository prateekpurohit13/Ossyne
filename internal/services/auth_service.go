package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"ossyne/internal/config"
	"ossyne/internal/db"
	"ossyne/internal/models"
	"strconv"
	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"
)

type AuthService struct {
	GitHubOAuthConfig *oauth2.Config
}

func NewAuthService(cfg config.Config) *AuthService {
	return &AuthService{
		GitHubOAuthConfig: &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.GitHubRedirectURL,
			Endpoint:     oauth2_github.Endpoint,
			Scopes:       []string{"read:user", "user:email"},
		},
	}
}

func (s *AuthService) HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state := "random-state-string"
	url := s.GitHubOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *AuthService) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if state != "random-state-string" {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	token, err := s.GitHubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	client := github.NewClient(s.GitHubOAuthConfig.Client(context.Background(), token))
	githubUser, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		http.Error(w, "Failed to get user from GitHub: "+err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := s.findOrCreateUser(githubUser, token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to find or create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	cliRedirectURL := "http://localhost:9999/auth/cli/callback"

	params := url.Values{}
	params.Add("token", token.AccessToken)
	redirectURLWithToken := fmt.Sprintf("%s?%s", cliRedirectURL, params.Encode())
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, fmt.Sprintf(`
		<html>
			<head>
				<title>Authentication Successful</title>
				<script>
					setTimeout(function() {
						window.location.href = "%s";
					}, 2000); // 2-second delay to show user info
				</script>
			</head>
			<body>
				<h1>Authentication Successful!</h1>
				<p>Welcome, %s!</p>
				<p>Your user ID is: %d</p>
				<p>You can now close this window and return to your terminal.</p>
				<p>Redirecting back to the CLI...</p>
			</body>
		</html>
	`, redirectURLWithToken, user.Username, user.ID))
}

func (s *AuthService) findOrCreateUser(githubUser *github.User, accessToken string) (*models.User, error) {
	var user models.User
	githubIDStr := strconv.FormatInt(*githubUser.ID, 10)
	result := db.DB.Where("github_id = ?", githubIDStr).First(&user)
	if result.Error != nil && result.Error.Error() != "record not found" {
		return nil, result.Error
	}
	user.GitHubAccessToken = &accessToken
	user.GithubID = &githubIDStr
	if githubUser.Login != nil {
		user.Username = *githubUser.Login
	}
	if githubUser.AvatarURL != nil {
		user.AvatarURL = *githubUser.AvatarURL
	}
	if githubUser.Email != nil && *githubUser.Email != "" {
		user.Email = *githubUser.Email
	}

	if result.RowsAffected > 0 {
		fmt.Printf("Updating existing user: %s (ID: %d)\n", user.Username, user.ID)
		if err := db.DB.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		fmt.Printf("Creating new user: %s\n", user.Username)
		if user.Email == "" {
			user.Email = fmt.Sprintf("%s@github.placeholder.com", user.Username)
		}
		if err := db.DB.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		fmt.Printf("Successfully created user with ID: %d\n", user.ID)
	}

	return &user, nil
}