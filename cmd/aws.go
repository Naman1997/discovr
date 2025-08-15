package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

// awsCmd represents the aws command
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// awsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// awsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
