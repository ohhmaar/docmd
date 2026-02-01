package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ohhmaar/docmd/internal/auth"
	"github.com/ohhmaar/docmd/internal/config"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Authenticate with Google",
	Long: `Initialize docmd by authenticating with your Google account.

This will open a browser window for you to authorize docmd to
access your Google Drive (only files created by docmd).

Your credentials will be stored in ~/.docmd/`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Force re-authentication")
}

func runInit(cmd *cobra.Command, args []string) error {
	if !auth.IsConfigured() {
		printError("OAuth credentials not configured!")
		fmt.Println()
		fmt.Println("To use docmd, you need to set up Google OAuth credentials:")
		fmt.Println("1. Go to https://console.cloud.google.com/")
		fmt.Println("2. Create a new project (or select existing)")
		fmt.Println("3. Enable the Google Drive API")
		fmt.Println("4. Create OAuth 2.0 credentials (Desktop app type)")
		fmt.Println("5. Set CLIENT_ID and CLIENT_SECRET in .env")
		fmt.Println()
		return fmt.Errorf("oauth not configured")
	}

	if auth.TokenExists() && !initForce {
		printWarning("Already authenticated!")
		fmt.Println("Use --force to re-authenticate.")
		return nil
	}

	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	fmt.Println("Starting authentication...")
	fmt.Println()

	token, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := auth.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println()
	printSuccess("Authentication successful!")

	configDir, _ := config.GetConfigDir()
	fmt.Printf("Credentials saved to %s\n", configDir)
	fmt.Println()
	fmt.Println("You can now use 'docmd link <file.md>' to sync files.")

	return nil
}
