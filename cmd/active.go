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
		internal.ActiveExport(ExportPathActive, ICMPMode)
	},
}

func init() {
	rootCmd.AddCommand(activeCmd)
	activeCmd.Flags().BoolVarP(&ICMPMode, "mode", "m", false, "Use ICMP echo requests instead of ARP (may require root/admin)")
	activeCmd.Flags().StringVarP(&networkInterface, "interface", "i", "", "Network interface to use for scanning (default: any)")
	activeCmd.Flags().StringVarP(&targetCIDR, "cidr", "r", "", "Target CIDR to scan (default: interface network)")
	activeCmd.Flags().StringVarP(&ExportPathActive, "export", "e", "", "Export results to CSV file")
	activeCmd.Flags().IntVarP(&concurrency, "concurrency", "p", 50, "Number of concurrent workers (default 50)")
	activeCmd.Flags().IntVarP(&timeout, "timeout", "", 2, "Timeout in seconds to wait for each reply (default 2)")
	activeCmd.Flags().IntVarP(&count, "count", "", 1, "Number of requests to send to each IP (default 1)")
	err := activeCmd.MarkFlagRequired("interface")
	if err != nil {
		fmt.Println("Error marking flag as required:", err)
	}

}
