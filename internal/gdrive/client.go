package gdrive

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/ohhmaar/docmd/internal/auth"
)

func GetClient() (*http.Client, error) {
	storedAuth, err := auth.LoadAuth()
	if err != nil {
		return nil, fmt.Errorf("not authenticated - run 'docmd init' first: %w", err)
	}

	config := auth.GetOAuthConfig(storedAuth.Credentials, "")

	tokenSource := config.TokenSource(context.Background(), storedAuth.Token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if newToken.AccessToken != storedAuth.Token.AccessToken {
		storedAuth.Token = newToken
		if err := auth.SaveAuth(storedAuth); err != nil {
			fmt.Printf("Warning: failed to save refreshed token: %v\n", err)
		}
	}

	return oauth2.NewClient(context.Background(), tokenSource), nil
}

func GetDriveService() (*drive.Service, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return srv, nil
}
