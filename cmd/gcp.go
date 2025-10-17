package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	GcpCsvExportPath string
	ProjectFilterStr string
	CredFile         string
)

var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Scan your GCP environment for Virtual machines",
	Long: `Scan your GCP environment for Virtual machines
`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.GcpScan(CredFile, ProjectFilterStr)
		internal.ShowResults(internal.Gcp_results)
		internal.ExportCSV(GcpCsvExportPath, internal.Gcp_results)
	},
}

func init() {
	rootCmd.AddCommand(gcpCmd)
	gcpCmd.Flags().StringVarP(&ProjectFilterStr, "project", "p", "", "Comma separated project names to use as a filter")
	gcpCmd.Flags().StringVarP(&CredFile, "cred", "c", "", "Path to service account json file to use for auth")
	gcpCmd.Flags().StringVarP(&GcpCsvExportPath, "export", "e", "", "Export results to CSV file")
}
