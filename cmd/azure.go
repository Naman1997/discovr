package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	SubscriptionID     string
	AzureCsvExportPath string
)

var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Scan your azure environment",
	Long: `Scan your azure environment for IT assets. For example:

Usage:
discovr azure --config FILENAME
`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.Azurescan(SubscriptionID)
		internal.ShowResults(internal.Azure_results)
		internal.ExportCSV(AzureCsvExportPath, internal.Azure_results)
	},
}

func init() {
	rootCmd.AddCommand(azureCmd)
	azureCmd.Flags().StringVarP(&SubscriptionID, "SubID", "s", "default", "Subscription ID for creating clients for API calls")
	azureCmd.Flags().StringVarP(&AzureCsvExportPath, "export", "e", "", "Export results to CSV file")
}
