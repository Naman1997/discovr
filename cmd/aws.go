package cmd

import (
	"runtime"

	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var (
	Region           string
	Profile          string
	AwsCsvExportPath string
	Config           []string
	Credential       []string
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Scan your AWS environment for EC2 instances",
	Long: `Scan your AWS environment for EC2 instances
`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.AwsScan(Region, Config, Credential, Profile)
		internal.ExportCSV(AwsCsvExportPath, internal.Aws_results)
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)
	awsCmd.Flags().StringVarP(&Region, "region", "r", "", "Region for filtering results")
	awsCmd.Flags().StringVarP(&Profile, "profile", "p", "", "AWS profile for fetching results")
	awsCmd.Flags().StringVarP(&AwsCsvExportPath, "export", "e", "", "Export results to CSV file")
	awsCmd.Flags().StringSliceVarP(&Credential, "credential", "x", []string{}, "Custom AWS credential file(s)")
	switch runtime.GOOS {
	case "windows":
		awsCmd.Flags().StringSliceVarP(&Config, "config", "c", []string{}, "Custom AWS config file(s)")
	default:
		awsCmd.Flags().StringSliceVarP(&Config, "config", "c", []string{}, "Custom AWS config file(s)")
	}
}
