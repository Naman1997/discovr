package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Scan your azure environment",
	Long: `Scan your azure environment for IT assets. For example:

Usage:
discovr azure --config FILENAME
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("azure called")
	},
}

func init() {
	rootCmd.AddCommand(azureCmd)

	var config string
	switch runtime.GOOS {
	case "windows":
		azureCmd.Flags().StringVarP(&config, "config", "c", "~/.azure/config", "Path to azure config file")
	default:
		azureCmd.Flags().StringVarP(&config, "config", "c", "\\%USERPROFILE%\\.azure\\config", "Path to azure config file")
	}
}
