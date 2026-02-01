package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ohhmaar/docmd/internal/auth"
	"github.com/ohhmaar/docmd/internal/config"
	"github.com/ohhmaar/docmd/internal/gdrive"
)

var (
	unlinkDelete bool
	unlinkYes    bool
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink <file.md>",
	Short: "Remove link between file and Google Doc",
	Long: `Remove the link between a local markdown file and its Google Doc.

By default, the Google Doc is NOT deleted. Use --delete to also
delete the Google Doc.`,
	Args: cobra.ExactArgs(1),
	RunE: runUnlink,
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
	unlinkCmd.Flags().BoolVarP(&unlinkDelete, "delete", "d", false, "Also delete the Google Doc")
	unlinkCmd.Flags().BoolVarP(&unlinkYes, "yes", "y", false, "Skip confirmation prompt")
}

func runUnlink(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	if !auth.TokenExists() {
		printError("Not authenticated!")
		fmt.Println("Run 'docmd init' first to authenticate with Google.")
		return fmt.Errorf("not authenticated")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	absPath, _ := filepath.Abs(filePath)
	link, exists := cfg.GetLink(absPath)
	if !exists {
		printWarning("File is not linked.")
		return nil
	}

	if !unlinkYes {
		fmt.Printf("Unlink %s from Google Docs?\n", filepath.Base(filePath))
		if unlinkDelete {
			printWarning("The Google Doc WILL be deleted!")
		} else {
			fmt.Println("The Google Doc will NOT be deleted.")
		}
		fmt.Println()
		fmt.Print("[Y/n]: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "" && input != "y" && input != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if unlinkDelete {
		fmt.Println("Deleting Google Doc...")
		if err := gdrive.DeleteDoc(link.DocID); err != nil {
			printWarning(fmt.Sprintf("Failed to delete Google Doc: %v", err))
			fmt.Println("The link will still be removed.")
		} else {
			printSuccess("Google Doc deleted.")
		}
	}

	if err := cfg.RemoveLink(absPath); err != nil {
		return fmt.Errorf("failed to remove link: %w", err)
	}

	printSuccess(fmt.Sprintf("Unlinked %s", filepath.Base(filePath)))

	return nil
}
