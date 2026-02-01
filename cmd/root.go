package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "docmd",
	Short: "Sync markdown files to Google Docs",
	Long: `docmd is a CLI tool that keeps your local markdown files
in sync with Google Docs.

Write in your favorite editor, sync to Google Docs for sharing.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func printSuccess(msg string) {
	fmt.Printf("OK: %s\n", msg)
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
}

func printWarning(msg string) {
	fmt.Printf("WARN: %s\n", msg)
}

func printInfo(msg string) {
	fmt.Printf("INFO: %s\n", msg)
}
