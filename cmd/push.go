package cmd

import (
	"bufio"
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
	pushForce bool
	pushAll   bool
)

var pushCmd = &cobra.Command{
	Use:   "push [file.md]",
	Short: "Push local changes to Google Docs",
	Long: `Sync local markdown changes to the linked Google Doc.

By default, checks for conflicts (remote changes since last sync).
Use --force to overwrite without checking.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, "Skip conflict check and overwrite")
	pushCmd.Flags().BoolVarP(&pushAll, "all", "a", false, "Push all linked files")
}

func runPush(cmd *cobra.Command, args []string) error {
	if !auth.TokenExists() {
		printError("Not authenticated!")
		fmt.Println("Run 'docmd init' first to authenticate with Google.")
		return fmt.Errorf("not authenticated")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var filesToPush []string

	if pushAll {
		for filePath := range cfg.Links {
			filesToPush = append(filesToPush, filePath)
		}
		if len(filesToPush) == 0 {
			printWarning("No linked files found.")
			fmt.Println("Use 'docmd link <file.md>' to link a file first.")
			return nil
		}
	} else if len(args) == 1 {
		absPath, _ := filepath.Abs(args[0])
		if _, exists := cfg.GetLink(absPath); !exists {
			printError("File is not linked!")
			fmt.Println("Use 'docmd link' to link this file first.")
			return fmt.Errorf("file not linked")
		}
		filesToPush = []string{absPath}
	} else {
		printError("No file specified!")
		fmt.Println("Usage: docmd push <file.md>")
		fmt.Println("   or: docmd push --all")
		return fmt.Errorf("no file specified")
	}

	for _, filePath := range filesToPush {
		if err := pushFile(cfg, filePath); err != nil {
			printError(fmt.Sprintf("Failed to push %s: %v", filepath.Base(filePath), err))
			if !pushAll {
				return err
			}
		}
	}

	return nil
}

func pushFile(cfg *config.Config, filePath string) error {
	link, ok := cfg.GetLink(filePath)
	if !ok {
		return fmt.Errorf("file not linked")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	if !pushForce {
		hasConflict, err := checkConflict(link)
		if err != nil {
			printWarning(fmt.Sprintf("Could not check for conflicts: %v", err))
		} else if hasConflict {
			resolved, err := handleConflict(link, filePath)
			if err != nil {
				return err
			}
			if !resolved {
				return nil
			}
		}
	}

	fmt.Printf("Syncing %s -> Google Docs...\n", filepath.Base(filePath))

	htmlContent, err := convert.FileToHTML(filePath)
	if err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	docInfo, err := gdrive.UpdateDoc(link.DocID, htmlContent)
	if err != nil {
		return fmt.Errorf("failed to update Google Doc: %w", err)
	}

	if err := cfg.UpdateSyncTime(filePath, docInfo.ModifiedTime.Format(time.RFC3339)); err != nil {
		printWarning(fmt.Sprintf("Failed to update sync time: %v", err))
	}

	printSuccess("Pushed successfully!")
	fmt.Printf("  Last synced: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return nil
}

func checkConflict(link *config.Link) (bool, error) {
	if link.LastSync.IsZero() {
		return false, nil
	}

	docInfo, err := gdrive.GetDocInfo(link.DocID)
	if err != nil {
		return false, err
	}

	return docInfo.ModifiedTime.After(link.LastSync), nil
}

func handleConflict(link *config.Link, filePath string) (bool, error) {
	docInfo, err := gdrive.GetDocInfo(link.DocID)
	if err != nil {
		return false, err
	}

	fileInfo, _ := os.Stat(filePath)

	fmt.Println()
	printWarning("Conflict detected!")
	fmt.Println("  The Google Doc has been modified since your last sync.")
	fmt.Println()
	fmt.Printf("  Local file:  modified %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))
	fmt.Printf("  Google Doc:  modified %s", docInfo.ModifiedTime.Format("2006-01-02 15:04:05"))
	if docInfo.ModifiedBy != "" {
		fmt.Printf(" (by %s)", docInfo.ModifiedBy)
	}
	fmt.Println()
	fmt.Printf("  Last sync:   %s\n", link.LastSync.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("What would you like to do?")
	fmt.Println("  [L] Push local (overwrite Google Doc)")
	fmt.Println("  [R] Keep remote (skip this push)")
	fmt.Println("  [A] Abort")
	fmt.Println()
	fmt.Print("Choice [L/R/A]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToUpper(input))

	switch input {
	case "L":
		return true, nil
	case "R":
		fmt.Println("Skipping push.")
		return false, nil
	default:
		fmt.Println("Aborted.")
		return false, nil
	}
}
