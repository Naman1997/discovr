package cmd

import (
	"fmt"
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
)

var Interface string

var passiveCmd = &cobra.Command{
	Use:   "passive",
	Short: "Scan local network passively",
	Long:  `Reads incomming packets to determine devices present on the network`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("passive called")
		internal.PassiveScan(Interface)
	},
}

func init() {
	localCmd.AddCommand(passiveCmd)

	passiveCmd.Flags().StringVarP(&Interface, "interface", "i", "eth0", "Interface to read packets from")
}
