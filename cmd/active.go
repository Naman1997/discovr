package cmd

import (
	"fmt"

	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	ExportPathActive string
	networkInterface string
	targetCIDR       string
	ICMPMode         bool
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Scan network actively",
	Long:  `Sends network requests across the CIDR range to determine device ip, mac address and other details with arp requests or icmp requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.DefaultScan(networkInterface, targetCIDR, ICMPMode)
		internal.ShowActiveResults()
		internal.ActiveExport(ExportPathActive, false)
	},
}

func init() {
	rootCmd.AddCommand(activeCmd)
	activeCmd.Flags().StringVarP(&ExportPathActive, "export", "e", "", "Export results to CSV file")
	activeCmd.Flags().StringVarP(&networkInterface, "interface", "i", "", "Network interface to use for scanning (default: any)")
	activeCmd.Flags().StringVarP(&targetCIDR, "cidr", "c", "", "Target CIDR to scan (default: interface network)")
	activeCmd.Flags().BoolVarP(&ICMPMode, "icmp", "", false, "Use ICMP echo requests instead of ARP (may require root/admin)")

	err := activeCmd.MarkFlagRequired("interface")
	if err != nil {
		fmt.Println("Error marking flag as required:", err)
	}

}
