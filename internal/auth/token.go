package auth

import (
	"encoding/json"
	"os"

	"github.com/ohhmaar/docmd/internal/config"
	"golang.org/x/oauth2"
)

type StoredAuth struct {
	Token       *oauth2.Token      `json:"token"`
	Credentials *GoogleCredentials `json:"credentials"`
}

func SaveAuth(auth *StoredAuth) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}

	tokenPath, err := config.GetTokenPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenPath, data, 0600)
}

func LoadAuth() (*StoredAuth, error) {
	tokenPath, err := config.GetTokenPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var auth StoredAuth
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, err
	}

	return &auth, nil
}

func LoadToken() (*oauth2.Token, error) {
	auth, err := LoadAuth()
	if err != nil {
		return nil, err
	}
	return auth.Token, nil
}

func LoadCredentialsFromStore() (*GoogleCredentials, error) {
	auth, err := LoadAuth()
	if err != nil {
		return nil, err
	}
	return auth.Credentials, nil
}

func TokenExists() bool {
	tokenPath, err := config.GetTokenPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(tokenPath)
	return err == nil
}

func DeleteToken() error {
	tokenPath, err := config.GetTokenPath()
	if err != nil {
		return err
	}
	return os.Remove(tokenPath)
}

// Backward compatibility - for reading old token format
func LoadTokenLegacy() (*oauth2.Token, error) {
	tokenPath, err := config.GetTokenPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}
