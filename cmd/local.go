package cmd

import (
	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Scan your local environment with the specified method",
	Long:  `Scan your local environment for live IT assets. Select 'active' or 'passive' mode to initiate scan.`,
}

func init() {
	rootCmd.AddCommand(localCmd)
}
