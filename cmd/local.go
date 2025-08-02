/*
Copyright © 2025 Naman Arora aroranaman17@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"github.com/google/gopacket/dumpcommand"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/pcap"
	"log"
	"time"

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
		arp()
	},
}

func init() {
	localCmd.Flags().StringVarP(&Interface, "interface", "i", "eth0", "Interface to read packets from")
	rootCmd.AddCommand(localCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// localCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// localCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Need to modify this further for extracting IP and MAC addresses
// Probably will move this to another subcommand (passive)
func passive() {
	defer util.Run()()
	var handle *pcap.Handle
	var err error

	inactive, err := pcap.NewInactiveHandle(Interface)
	if err != nil {
		log.Fatalf("Could not create handle for interface: %v", err)
	}
	defer inactive.CleanUp()
	inactive.SetSnapLen(65536)
	inactive.SetPromisc(true)
	if err = inactive.SetTimeout(time.Second); err != nil {
		log.Fatalf("Could not set timeout: %v", err)
	}
	if handle, err = inactive.Activate(); err != nil {
		log.Fatal("PCAP Activate error:", err)
	}
	defer handle.Close()
	
	dumpcommand.Run(handle)
}
