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
	concurrency      int
	timeout          int
	count            int
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Scan network actively",
	Long:  `Sends network requests across the CIDR range to determine device ip, mac address and other details with arp requests or icmp requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.DefaultScan(networkInterface, targetCIDR, ICMPMode, concurrency, timeout, count)
		if !ICMPMode {
			internal.ShowActiveResults()
		} else {
			internal.ShowIcmpResults()
		}
		internal.ActiveExport(ExportPathActive, ICMPMode)
	},
}

func init() {
	rootCmd.AddCommand(activeCmd)
	activeCmd.Flags().BoolVarP(&ICMPMode, "mode", "m", false, "Use ICMP echo requests instead of ARP (true/false) (default false)")
	activeCmd.Flags().StringVarP(&networkInterface, "interface", "i", "", "Network interface to use for scanning (ARP)")
	activeCmd.Flags().StringVarP(&targetCIDR, "cidr", "r", "", "Target CIDR to scan (ARP, ICMP)")
	activeCmd.Flags().StringVarP(&ExportPathActive, "export", "e", "", "Export results to CSV file")
	activeCmd.Flags().IntVarP(&concurrency, "concurrency", "p", 50, "Number of concurrent workers (ICMP)")
	activeCmd.Flags().IntVarP(&timeout, "timeout", "t", 2, "Timeout in seconds to wait for each reply (ICMP)")
	activeCmd.Flags().IntVarP(&count, "count", "c", 1, "Number of requests to send to each IP (ICMP)")
	err := activeCmd.MarkFlagRequired("interface")
	if err != nil {
		fmt.Println("Error marking flag as required:", err)
	}

}
