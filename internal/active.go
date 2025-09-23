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

// DefaultScan example: you can set desiredCIDR to "" to use interface mask,
// or "192.168.0.0/28" to request scanning that CIDR (must be inside interface network).
func DefaultScan(networkInterface string, targetCIDR string) {
	// var networkInterface string = "Wi-Fi"
	var wg sync.WaitGroup

	// user-specified target CIDR; set to "" to use interface default
	// desiredCIDR := "192.168.0.16/28" // <-- change this or set to ""
	// desiredCIDR := ""
	// Get a list of all device
	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	if networkInterface == "any" {
		ifaces, err := net.Interfaces()
		if err != nil {
			panic(err)
		}
		for _, iface := range ifaces {
			wg.Add(1)
			go func(iface net.Interface) {
				defer wg.Done()
				if err := scan(&iface, &devices, targetCIDR); err != nil {
					log.Printf("interface %v: %v", iface.Name, err)
				}
			}(iface)
		}
	} else {
		netiface, err := net.InterfaceByName(networkInterface)
		if err != nil {
			panic(err)
		}
		wg.Add(1)
		go func(netiface net.Interface) {
			defer wg.Done()
			if err := scan(&netiface, &devices, targetCIDR); err != nil {
				log.Printf("interface %v: %v", netiface.Name, err)
			}
		}(*netiface)
	}
	wg.Wait()
}

// scan now accepts targetCIDR. If targetCIDR == "" it uses interface network as before.
func scan(iface *net.Interface, devices *[]pcap.Interface, targetCIDR string) error {

	// find interface IPv4 (same as before)
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
	// addr is connected IPv4 network from interface
	if addr == nil {
		return errors.New("no good IP network found")
	} else if addr.IP[0] == 127 {
		return errors.New("skipping localhost")
	} else if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
		return errors.New("mask means network is too large")
	}

	// If user supplied a targetCIDR, parse and verify it is inside addr.
	var scanNet *net.IPNet

	if targetCIDR != "" {
		_, tNet, err := net.ParseCIDR(targetCIDR)
		if err != nil {
			return fmt.Errorf("invalid target CIDR %q: %v", targetCIDR, err)
		}
		// align target network IP to its network base
		tNet = alignToNetwork(tNet)
		if !isSubnetWithin(addr, tNet) {
			return fmt.Errorf("requested CIDR %v is outside connected interface network %v", tNet.String(), addr.String())
		}
		scanNet = tNet
	} else {
		// use interface network (align IP to network base)
		scanNet = addr
	}

	log.Printf("Using network range %v for interface %v", scanNet, iface.Name)

	// find device name (same)
	var deviceName string
	for _, d := range *devices {
		if strings.Contains(fmt.Sprint(d.Addresses), fmt.Sprint(addr.IP)) {
			deviceName = d.Name
		}
	}
	if deviceName == "" {
		return fmt.Errorf("cannot find the corresponding device for the interface %s", iface.Name)
	}

	handle, err := pcap.OpenLive(deviceName, 65536, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	// Start read goroutine with stop channel
	stop := make(chan struct{})
	go readARP(handle, iface, stop)
	defer close(stop)

	// write ARP only for scanNet (which may be the requested /28, /30, or the full iface /24)
	if err := writeARP(handle, iface, scanNet, addr); err != nil {
		return err
	}

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
				log.Printf("Duplicate detected for %s", key)
				mu.Unlock()
				return
			} else {
				seenResults[key] = true
				defaultscan_results = append(defaultscan_results, result)
			}

			mu.Unlock()

			log.Printf("IP %v is at %v from interface: %v",
				result.Dest_IP, result.Dest_Mac, result.Interface)
		}
	}
}

// writeARP writes an ARP request for each address on our local network to the
// pcap handle.
func writeARP(handle *pcap.Handle, iface *net.Interface, addr *net.IPNet, intAddr *net.IPNet) error {
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
	for _, ip := range ips(addr, intAddr) {
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
func ips(n *net.IPNet, o *net.IPNet) (out []net.IP) {

	// n and o are different

	orinum := binary.BigEndian.Uint32([]byte(o.IP))
	orimask := binary.BigEndian.Uint32([]byte(o.Mask))
	oribroadcast := orinum | ^orimask
	orinetwork := orinum & orimask

	num := binary.BigEndian.Uint32([]byte(n.IP))
	mask := binary.BigEndian.Uint32([]byte(n.Mask))
	network := num & mask
	broadcast := network | ^mask

	for ip := network; ip <= broadcast; ip++ {
		if ip == orinetwork || ip == oribroadcast {
			continue // skip network and broadcast addresses
		}
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], ip)
		out = append(out, net.IP(buf[:]))
	}

	return
}

// ipToUint32 converts a 4-byte IPv4 to uint32.
func ipToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32([]byte(ip.To4()))
}

// networkRange returns network and broadcast (uint32) for a net.IPNet
func networkRange(n *net.IPNet) (network uint32, broadcast uint32) {
	num := ipToUint32(n.IP)
	mask := ipToUint32(net.IP(n.Mask)) // n.Mask is 4 bytes for IPv4
	network = num & mask
	broadcast = network | ^mask
	return
}

// isSubnetWithin returns true if targetNet range is fully inside ifaceNet range.
func isSubnetWithin(ifaceNet, targetNet *net.IPNet) bool {
	if ifaceNet == nil || targetNet == nil {
		return false
	}
	if ifaceNet.IP.To4() == nil || targetNet.IP.To4() == nil {
		return false
	}
	ifaceNetStart, ifaceNetEnd := networkRange(ifaceNet)
	targetStart, targetEnd := networkRange(targetNet)
	// target start must be >= iface start and target end <= iface end
	return targetStart >= ifaceNetStart && targetEnd <= ifaceNetEnd
}

// alignToNetwork returns targetNet with IP aligned to its network address (useful after parsing).
func alignToNetwork(n *net.IPNet) *net.IPNet {
	if n == nil {
		return nil
	}
	mask := n.Mask
	ip := n.IP.Mask(mask)
	return &net.IPNet{IP: ip, Mask: mask}
}
