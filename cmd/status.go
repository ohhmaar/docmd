package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ohhmaar/docmd/internal/auth"
	"github.com/ohhmaar/docmd/internal/config"
	"github.com/ohhmaar/docmd/internal/gdrive"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status of linked files",
	Long:  `Display all linked markdown files and their sync status.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if !auth.TokenExists() {
		printError("Not authenticated!")
		fmt.Println("Run 'docmd init' first to authenticate with Google.")
		return fmt.Errorf("not authenticated")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Links) == 0 {
		printWarning("No linked files.")
		fmt.Println("Use 'docmd link <file.md>' to link a file to Google Docs.")
		return nil
	}

	fmt.Println("Linked files:")
	fmt.Println()

	for filePath, link := range cfg.Links {
		displayPath := filePath
		if cwd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(cwd, filePath); err == nil && !filepath.IsAbs(rel) {
				displayPath = "./" + rel
			}
		}

		fmt.Printf("  %s\n", displayPath)
		fmt.Printf("    -> %s\n", link.DocURL)

		status := getFileStatus(filePath, link, cfg)
		fmt.Printf("    Status: %s\n", status)

		if !link.LastSync.IsZero() {
			fmt.Printf("    Last push: %s\n", link.LastSync.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d file(s) linked\n", len(cfg.Links))

	return nil
}

func getFileStatus(filePath string, link *config.Link, cfg *config.Config) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "Local file missing"
	}

	if !gdrive.DocExists(link.DocID) {
		return "Google Doc not found"
	}

	hasLocalChanges, err := cfg.HasLocalChanges(filePath)
	if err != nil {
		return "Unknown (could not check)"
	}

	if hasLocalChanges {
		return "Local changes pending"
	}

	return "In sync"
}
