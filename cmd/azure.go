package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// azureCmd represents the azure command
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// azureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// azureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
