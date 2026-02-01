package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func VersionString() string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}

	commitStr := strings.TrimSpace(commit)
	dateStr := strings.TrimSpace(date)

	if commitStr == "" && dateStr == "" {
		return v
	}
	if commitStr == "" {
		return fmt.Sprintf("%s (%s)", v, dateStr)
	}
	if dateStr == "" {
		return fmt.Sprintf("%s (%s)", v, commitStr)
	}
	return fmt.Sprintf("%s (%s %s)", v, commitStr, dateStr)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version number, commit hash, and build date of docmd.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(VersionString())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
