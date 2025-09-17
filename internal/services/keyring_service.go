package services

import (
	"fmt"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = "ossyne-cli"
	keyringUser    = "github_access_token"
)

type KeyringService struct{}

func (s *KeyringService) SetToken(token string) error {
	err := keyring.Set(keyringService, keyringUser, token)
	if err != nil {
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}
	return nil
}

func (s *KeyringService) GetToken() (string, error) {
	token, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", fmt.Errorf("not logged in. Please run 'osm login --provider github'")
		}
		return "", fmt.Errorf("failed to retrieve token from keyring: %w", err)
	}
	return token, nil
}

func (s *KeyringService) ClearToken() error {
	err := keyring.Delete(keyringService, keyringUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil
		}
		return fmt.Errorf("failed to clear token from keyring: %w", err)
	}
	return nil
}