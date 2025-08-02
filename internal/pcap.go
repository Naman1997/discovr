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
