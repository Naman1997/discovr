// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// arpscan implements ARP scanning of all interfaces' local networks using
// gopacket and its subpackages.  This example shows, among other things:
//   - Generating and sending packet data
//   - Reading in packet data and interpreting it
//   - Use of the 'pcap' subpackage for reading/writing
package internal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	defaultscan_results []ScanResultDfActive
	seenResults         = make(map[string]bool)
	mu                  sync.Mutex
)

type ScanResultDfActive struct {
	Interface string
	Dest_IP   string
	Dest_Mac  string
}

func DefaultScan() {
	var wg sync.WaitGroup
	// Get a list of all interfaces.
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	// Get a list of all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	// For each interface scan the subnet range
	for _, iface := range ifaces {
		// TODO: Add a interface filter like this
		// if (iface.Name == "virbr0") {
		wg.Add(1)
		// Start up a scan on each interface.
		go func(iface net.Interface) {
			defer wg.Done()

			if err := scan(&iface, &devices); err != nil {
				log.Printf("interface %v: %v", iface.Name, err)
			}

		}(iface)
		// }
	}
	wg.Wait()

	// TODO: Scan a cutom subnet range (ex: 192.168.0.0/28)
}

// scan scans an individual interface's local network for machines using ARP requests/replies.
// scan loops forever, sending packets out regularly.  It returns an error if
// it's ever unable to write a packet.
func scan(iface *net.Interface, devices *[]pcap.Interface) error {

	// We just look for IPv4 addresses, so try to find if the interface has one.
	var addr *net.IPNet
	if addrs, err := iface.Addrs(); err != nil {
		return err
	} else {
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					addr = &net.IPNet{
						IP:   ip4,
						Mask: ipnet.Mask[len(ipnet.Mask)-4:],
					}
					break
				}
			}
		}
	}
	// Sanity-check that the interface has a good address.
	if addr == nil {
		return errors.New("no good IP network found")
	} else if addr.IP[0] == 127 {
		return errors.New("skipping localhost")
	} else if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
		return errors.New("mask means network is too large")
	}
	log.Printf("Using network range %v for interface %v", addr, iface.Name)

	// Try to find a match between device and interface
	var deviceName string
	for _, d := range *devices {
		if strings.Contains(fmt.Sprint(d.Addresses), fmt.Sprint(addr.IP)) {
			deviceName = d.Name
		}
	}

	if deviceName == "" {
		return fmt.Errorf("cannot find the corresponding device for the interface %s", iface.Name)
	}

	// Open up a pcap handle for packet reads/writes.
	handle, err := pcap.OpenLive(deviceName, 65536, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	// Start up a goroutine to read in packet data.
	stop := make(chan struct{})
	go readARP(handle, iface, stop)
	defer close(stop)

	// send packets only once
	if err := writeARP(handle, iface, addr); err != nil {
		//log.Printf("error writing packets on %v: %v", iface.Name, err)
		return err
	}

	// give some time to collect ARP replies before exiting
	time.Sleep(10 * time.Second)
	return nil
}

// readARP reads in packets from the pcap handle, looking for ARP replies.
func readARP(handle *pcap.Handle, iface *net.Interface, stop chan struct{}) {
	src := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	in := src.Packets()

	for {
		var packet gopacket.Packet
		select {
		case <-stop:
			return
		case packet = <-in:
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer == nil {
				continue
			}
			arp := arpLayer.(*layers.ARP)

			if arp.Operation != layers.ARPReply || bytes.Equal([]byte(iface.HardwareAddr), arp.SourceHwAddress) {
				continue
			}

			result := ScanResultDfActive{
				Interface: iface.Name,
				Dest_IP:   net.IP(arp.SourceProtAddress).String(),
				Dest_Mac:  net.HardwareAddr(arp.SourceHwAddress).String(),
			}

			key := result.Interface + "_" + result.Dest_IP + "_" + result.Dest_Mac
			mu.Lock()
			if seenResults[key] {
				log.Printf("Duplicate detected for %s, stopping all scans...", key)
				mu.Unlock()
				close(stop) // stop this goroutine
				// close(globalStop)
				return
			}
			seenResults[key] = true
			defaultscan_results = append(defaultscan_results, result)
			mu.Unlock()

			log.Printf("IP %v is at %v from interface: %v",
				result.Dest_IP, result.Dest_Mac, result.Interface)
		}
	}
}

// writeARP writes an ARP request for each address on our local network to the
// pcap handle.
func writeARP(handle *pcap.Handle, iface *net.Interface, addr *net.IPNet) error {
	// Set up all the layers' fields we can.
	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(addr.IP),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
	}
	// Set up buffer and options for serialization.
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	// Send one packet for every address.
	for _, ip := range ips(addr) {
		arp.DstProtAddress = []byte(ip)
		gopacket.SerializeLayers(buf, opts, &eth, &arp)
		if err := handle.WritePacketData(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// ips is a simple and not very good method for getting all IPv4 addresses from a
// net.IPNet.  It returns all IPs it can over the channel it sends back, closing
// the channel when done.
func ips(n *net.IPNet) (out []net.IP) {
	num := binary.BigEndian.Uint32([]byte(n.IP))
	mask := binary.BigEndian.Uint32([]byte(n.Mask))
	network := num & mask
	broadcast := network | ^mask
	for network++; network < broadcast; network++ {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], network)
		out = append(out, net.IP(buf[:]))
	}
	return
}
