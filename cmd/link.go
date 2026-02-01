package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ohhmaar/docmd/internal/auth"
	"github.com/ohhmaar/docmd/internal/config"
	"github.com/ohhmaar/docmd/internal/convert"
	"github.com/ohhmaar/docmd/internal/gdrive"
)

var (
	linkTitle    string
	linkFolderID string
)

var linkCmd = &cobra.Command{
	Use:   "link <file.md>",
	Short: "Link a markdown file to a new Google Doc",
	Long: `Create a new Google Doc from a markdown file and link them.

The markdown file will be converted to HTML and uploaded to Google Docs.
Future changes can be synced using 'docmd push'.`,
	Args: cobra.ExactArgs(1),
	RunE: runLink,
}

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().StringVarP(&linkTitle, "title", "t", "", "Custom title for the Google Doc (default: filename)")
	linkCmd.Flags().StringVarP(&linkFolderID, "folder", "f", "", "Google Drive folder ID to create the doc in")
}

func runLink(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	if !auth.TokenExists() {
		printError("Not authenticated!")
		fmt.Println("Run 'docmd init' first to authenticate with Google.")
		return fmt.Errorf("not authenticated")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	absPath, _ := filepath.Abs(filePath)
	if link, exists := cfg.GetLink(absPath); exists {
		printWarning("File is already linked!")
		fmt.Printf("  Doc URL: %s\n", link.DocURL)
		fmt.Println()
		fmt.Println("Use 'docmd push' to sync changes, or 'docmd unlink' first to create a new doc.")
		return nil
	}

	title := linkTitle
	if title == "" {
		base := filepath.Base(filePath)
		title = strings.TrimSuffix(base, filepath.Ext(base))
	}

	fmt.Printf("Creating Google Doc from %s...\n", filePath)

	htmlContent, err := convert.FileToHTML(filePath)
	if err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	docInfo, err := gdrive.CreateDoc(title, htmlContent, linkFolderID)
	if err != nil {
		return fmt.Errorf("failed to create Google Doc: %w", err)
	}

	hash, _ := config.HashFile(absPath)

	link := &config.Link{
		DocID:           docInfo.ID,
		DocURL:          docInfo.URL,
		Title:           docInfo.Title,
		CreatedAt:       time.Now(),
		LastSync:        time.Now(),
		LocalHashAtSync: hash,
	}

	if err := cfg.AddLink(absPath, link); err != nil {
		return fmt.Errorf("failed to save link: %w", err)
	}

	fmt.Println()
	printSuccess(fmt.Sprintf("Created: \"%s\"", docInfo.Title))
	fmt.Printf("  URL: %s\n", docInfo.URL)
	fmt.Println()
	fmt.Println("File linked! Use 'docmd push' to sync future changes.")

	return nil
}
