package cmd

import (
	"fmt"
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var Interface string

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Scan your local environment",
	Long: `Scan your local environment for live IT assets. For example:

Usage:
discovr local passive [--interface INTERFACE]
discovr local active [--cidr CIDR (--ping|--arp)]
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("local called")
		internal.PassiveScan(Interface)
	},
}

func init() {
	localCmd.Flags().StringVarP(&Interface, "interface", "i", "wlp0s20f3", "Interface to read packets from")
	rootCmd.AddCommand(localCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// localCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// localCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
