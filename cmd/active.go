package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Scan local network actively",
	Long:  `Sends network requests across the CIDR range to determine device ip, mac address and other details`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.ActiveScan()
	},
}

func init() {
	localCmd.AddCommand(activeCmd)

	var cidr string
	activeCmd.Flags().StringVarP(&cidr, "cidr", "", "", "Custom CIDR range to scan")
}
