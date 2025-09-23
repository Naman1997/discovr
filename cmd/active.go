package cmd

import (
	"github.com/spf13/cobra"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Scan network actively",
	Long:  `Sends network requests across the CIDR range with Arp-scan or ICMP echo requests`,
}

func init() {
	rootCmd.AddCommand(activeCmd)

}
