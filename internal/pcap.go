package internal

import (
	"github.com/google/gopacket/dumpcommand"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/pcap"
	"log"
	"time"
)

// Need to modify this further for extracting IP and MAC addresses
// Probably will move this to another subcommand (passive)
func PassiveScan(interfaceName string) {
	defer util.Run()()
	var handle *pcap.Handle
	var err error

	inactive, err := pcap.NewInactiveHandle(interfaceName)
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
