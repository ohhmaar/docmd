package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ohhmaar/docmd/internal/auth"
	"github.com/ohhmaar/docmd/internal/config"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init <credentials.json>",
	Short: "Authenticate with Google",
	Long: `Initialize docmd by authenticating with your Google account.

This will open a browser window for you to authorize docmd to
access your Google Drive (only files created by docmd).

Your credentials will be stored in ~/.docmd/`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Force re-authentication")
}

func runInit(cmd *cobra.Command, args []string) error {
	credsPath := args[0]

	creds, err := auth.LoadCredentials(credsPath)
	if err != nil {
		printError("Invalid credentials file")
		return err
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

	token, err := auth.Authenticate(creds)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	storedAuth := &auth.StoredAuth{
		Token:       token,
		Credentials: creds,
	}

	if err := auth.SaveAuth(storedAuth); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println()
	printSuccess("Authentication successful!")

	configDir, _ := config.GetConfigDir()
	fmt.Printf("Credentials saved to %s\n", configDir)
	fmt.Println()
	fmt.Println("You can now use 'docmd link <file.md>' to sync files.")

	return nil
}
