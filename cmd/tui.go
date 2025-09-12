package cmd

import (
	//	"github.com/Naman1997/discovr/internal" // will need when using active/passive functions
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Use TUI",
	Long:  "Use a TUI instead of CLI",
	Run: func(cmd *cobra.Command, args []string) {
		RunTui()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
