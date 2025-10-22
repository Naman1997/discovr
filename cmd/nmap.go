package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	Target      string
	Ports       string
	OsDetection bool
	PathActive  string
)

var nmapCmd = &cobra.Command{
	Use:   "nmap",
	Short: "Scan network with nmap",
	Long:  `Sends network requests with NMAP tool across the CIDR range to determine device ip, mac address and other details`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.NmapScan(Target, Ports, OsDetection)
		internal.ShowResults(internal.Active_results)
		internal.ExportCSV(PathActive, internal.Active_results)
		internal.UploadResults(UploadUrl, PathActive, internal.Active_results, "nmap_")
	},
}

func init() {
	rootCmd.AddCommand(nmapCmd)
	nmapCmd.Flags().StringVarP(&Target, "target", "t", "127.0.0.1", "Target CIDR range or IP address to scan")
	nmapCmd.Flags().StringVarP(&Ports, "ports", "p", "", "Ports to scan on target systems (defaults to top 1000 most common ports)")
	nmapCmd.Flags().BoolVarP(&OsDetection, "detect-os", "d", false, "Enable OS detection (requires sudo)")
	nmapCmd.Flags().StringVarP(&PathActive, "export", "e", "", "Export results to CSV file")
}
