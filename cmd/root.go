package cmd

import (
	"os"

	"github.com/Naman1997/discovr/verbose"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "discovr",
	Short: "Portable asset discovery tool for mapping your networks",
	Long:  `Find more information at: https://github.com/Naman1997/discovr`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose.Verbose, "verbose", "v", false, "Enable verbose")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func VerboseEnabled() bool {
	return verbose.Verbose
}
