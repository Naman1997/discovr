package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	ExportPathActive string
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Scan local network actively",
	Long:  `Sends network requests across the CIDR range to determine device ip, mac address and other details`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.DefaultScan()
		internal.ActiveExport(ExportPathActive, false)
	},
}

func init() {
	rootCmd.AddCommand(activeCmd)
	activeCmd.Flags().StringVarP(&ExportPathActive, "export", "e", "", "Export results to CSV file")
}
