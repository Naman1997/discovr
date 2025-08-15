package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Scan your aws environment",
	Long: `Scan your aws environment for IT assets. For example:

Usage:
discovr aws --config FILENAME
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aws called")
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)

	var config string
	switch runtime.GOOS {
	case "windows":
		awsCmd.Flags().StringVarP(&config, "config", "c", "~/.aws/config", "Path to aws config file")
	default:
		awsCmd.Flags().StringVarP(&config, "config", "c", "\\%USERPROFILE%\\.aws\\config", "Path to aws config file")
	}
}
