package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// activeCmd represents the active command
var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Scan local network actively",
	Long: `Sends network requests across the CIDR range to determine device ip, mac address and other details`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("active called")
	},
}

func init() {
	localCmd.AddCommand(activeCmd)

	var cidr string
	activeCmd.Flags().StringVarP(&cidr, "cidr", "", "", "Custom CIDR range to scan")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// activeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// activeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
