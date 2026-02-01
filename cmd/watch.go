package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/ohhmaar/docmd/internal/auth"
	"github.com/ohhmaar/docmd/internal/config"
	"github.com/ohhmaar/docmd/internal/convert"
	"github.com/ohhmaar/docmd/internal/gdrive"
	"github.com/ohhmaar/docmd/internal/sync"
)

var (
	watchDebounce int
	watchAll      bool
)

var watchCmd = &cobra.Command{
	Use:   "watch [file.md]",
	Short: "Watch for changes and auto-sync",
	Long: `Watch a markdown file for changes and automatically push to Google Docs.

Changes are debounced to avoid excessive API calls during rapid edits.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().IntVarP(&watchDebounce, "debounce", "d", 500, "Debounce delay in milliseconds")
	watchCmd.Flags().BoolVarP(&watchAll, "all", "a", false, "Watch all linked files")
}

func runWatch(cmd *cobra.Command, args []string) error {
	if !auth.TokenExists() {
		printError("Not authenticated!")
		fmt.Println("Run 'docmd init' first to authenticate with Google.")
		return fmt.Errorf("not authenticated")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var filesToWatch []string

	if watchAll {
		for filePath := range cfg.Links {
			if _, err := os.Stat(filePath); err == nil {
				filesToWatch = append(filesToWatch, filePath)
			}
		}
		if len(filesToWatch) == 0 {
			printWarning("No linked files found to watch.")
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
		filesToWatch = []string{absPath}
	} else {
		printError("No file specified!")
		fmt.Println("Usage: docmd watch <file.md>")
		fmt.Println("   or: docmd watch --all")
		return fmt.Errorf("no file specified")
	}

	if len(filesToWatch) == 1 {
		fmt.Printf("Watching %s for changes...\n", filepath.Base(filesToWatch[0]))
	} else {
		fmt.Printf("Watching %d files for changes...\n", len(filesToWatch))
		for _, f := range filesToWatch {
			fmt.Printf("  - %s\n", filepath.Base(f))
		}
	}
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	syncFunc := func(filePath string) error {
		return syncFile(cfg, filePath)
	}

	errChan := make(chan error, 1)
	go func() {
		watchConfig := sync.WatchConfig{
			DebounceMs: watchDebounce,
			OnChange:   syncFunc,
		}

		if len(filesToWatch) == 1 {
			errChan <- sync.WatchFile(filesToWatch[0], watchConfig)
		} else {
			errChan <- sync.WatchFiles(filesToWatch, watchConfig)
		}
	}()

	select {
	case <-sigChan:
		fmt.Println("\nStopping watch...")
		return nil
	case err := <-errChan:
		return err
	}
}

func syncFile(cfg *config.Config, filePath string) error {
	link, ok := cfg.GetLink(filePath)
	if !ok {
		return fmt.Errorf("file not linked")
	}

	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] Change detected in %s\n", timestamp, filepath.Base(filePath))
	fmt.Printf("[%s] Pushing to Google Docs...\n", timestamp)

	htmlContent, err := convert.FileToHTML(filePath)
	if err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	docInfo, err := gdrive.UpdateDoc(link.DocID, htmlContent)
	if err != nil {
		return fmt.Errorf("failed to update Google Doc: %w", err)
	}

	if err := cfg.UpdateSyncTime(filePath, docInfo.ModifiedTime.Format(time.RFC3339)); err != nil {
		fmt.Printf("[%s] Warning: failed to update sync time: %v\n", timestamp, err)
	}

	fmt.Printf("[%s] Synced successfully\n", timestamp)
	return nil
}
