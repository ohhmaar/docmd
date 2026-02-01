package auth

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"

	"github.com/ohhmaar/docmd/internal/config"
)

func SaveToken(token *oauth2.Token) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}

	tokenPath, err := config.GetTokenPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenPath, data, 0600)
}

func LoadToken() (*oauth2.Token, error) {
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
