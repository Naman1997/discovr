package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	Target     string
	Ports      string
	PathActive string
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Scan local network actively",
	Long:  `Sends network requests across the CIDR range to determine device ip, mac address and other details`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.ActiveScan(Target, Ports)
		header := []string{"ID", "Protocol", "State", "Service", "Product"}
		internal.ActiveExport(PathActive, header)
	},
}

func init() {
	localCmd.AddCommand(activeCmd)

	activeCmd.Flags().StringVarP(&Target, "target", "t", "127.0.0.1", "Target CIDR range or IP address to scan")
	activeCmd.Flags().StringVarP(&Ports, "ports", "p", "", "Ports to scan on target systems (defaults to top 1000 most common ports)")
	activeCmd.Flags().StringVarP(&PathActive, "export", "e", "", "Export results to CSV file")
}
