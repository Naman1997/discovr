package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	Interface   string
	ScanTime    int
	PathPassive string
)

var passiveCmd = &cobra.Command{
	Use:   "passive",
	Short: "Scan local network passively",
	Long:  `Reads incomming packets to determine devices present on the network`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.PassiveScan(Interface, ScanTime)
		internal.ShowResults(internal.Passive_results)
		internal.ExportCSV(PathPassive, internal.Passive_results)
		internal.UploadResults(UploadUrl, PathPassive, internal.Passive_results, "passive_")
	},
}

func init() {
	rootCmd.AddCommand(passiveCmd)
	passiveCmd.Flags().StringVarP(&Interface, "interface", "i", "any", "Interface to read packets from")
	passiveCmd.Flags().IntVarP(&ScanTime, "duration", "d", 10, "Number of seconds to run the scan")
	passiveCmd.Flags().StringVarP(&PathPassive, "export", "e", "", "Export results to CSV file")
}
