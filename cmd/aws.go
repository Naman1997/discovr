package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
	"runtime"
)

var (
	Region      string
	Credential      string
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Scan your aws environment",
	Long: `Scan your environment for EC2 instances. For example:

Usage:
discovr aws --region REGION --config CONFIG_PATH --credential CREDENTIAL_PATH
`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.AwsScan(Region)
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().StringVarP(&Region, "region", "r", "", "Region for filtering results")
	awsCmd.Flags().StringVarP(&Credential, "credential", "x", "", "Path to aws credential file")

	// TODO: Use this for auth options
	var config string
	switch runtime.GOOS {
	case "windows":
		awsCmd.Flags().StringVarP(&config, "config", "c", "\\%USERPROFILE%\\.aws\\config", "Path to aws config file")
	default:
		awsCmd.Flags().StringVarP(&config, "config", "c", "~/.aws/config", "Path to aws config file")
	}
}
