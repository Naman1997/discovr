package cmd

import (
	"github.com/Naman1997/discovr/internal"
	"github.com/spf13/cobra"
	"time"
)

var (
	Interface string
	scanTime  time.Duration = 10000 * time.Millisecond
)

var passiveCmd = &cobra.Command{
	Use:   "passive",
	Short: "Scan local network passively",
	Long:  `Reads incomming packets to determine devices present on the network`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.PassiveScan(Interface, scanTime)
	},
}

func init() {
	localCmd.AddCommand(passiveCmd)

	passiveCmd.Flags().StringVarP(&Interface, "interface", "i", "eth0", "Interface to read packets from")
}
