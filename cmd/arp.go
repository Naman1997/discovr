package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	ExportPathActive string
	networkInterface string
	targetCIDR       string
)

var arpCmd = &cobra.Command{
	Use:   "arp-scan",
	Short: "Scan local network actively",
	Long:  `Sends network requests across the CIDR range to determine device ip, mac address and other details`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.DefaultScan(networkInterface, targetCIDR)
		internal.ActiveExport(ExportPathActive, false)
	},
}

func init() {
	activeCmd.AddCommand(arpCmd)
	arpCmd.Flags().StringVarP(&ExportPathActive, "export", "e", "", "Export results to CSV file")
	arpCmd.Flags().StringVarP(&networkInterface, "interface", "i", "any", "Network interface to use for scanning (default: any)")
	arpCmd.Flags().StringVarP(&targetCIDR, "cidr", "c", "", "Target CIDR to scan (default: interface network)")
}
