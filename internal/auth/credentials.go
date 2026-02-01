package auth

import (
	"encoding/json"
	"fmt"
	"os"
)

type GoogleCredentials struct {
	Installed struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"installed"`
}

func LoadCredentials(path string) (*GoogleCredentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds GoogleCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if creds.Installed.ClientID == "" || creds.Installed.ClientSecret == "" {
		return nil, fmt.Errorf("credentials file missing client_id or client_secret")
	}

	return &creds, nil
}
