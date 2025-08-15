package cmd

import (
	"fmt"
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var Interface string

// passiveCmd represents the passive command
var passiveCmd = &cobra.Command{
	Use:   "passive",
	Short: "Scan local network passively",
	Long: `Reads incomming packets to determine devices present on the network`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("passive called")
		internal.PassiveScan(Interface)
	},
}

func init() {
	localCmd.AddCommand(passiveCmd)

	passiveCmd.Flags().StringVarP(&Interface, "interface", "i", "eth0", "Interface to read packets from")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// passiveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// passiveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
