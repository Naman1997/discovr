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
	icmpscan_results    []ScanResultICMP
	seenResults         = make(map[string]bool)
	mu                  sync.Mutex
	stats               SweepStats
	timeout             time.Duration
	wg                  sync.WaitGroup
)

type ScanResultDfActive struct {
	Interface string
	Dest_IP   string
	Dest_Mac  string
}

// SweepStats holds sweep statistics
type SweepStats struct {
	PacketsSent     int
	PacketsReceived int
	TotalRTT        time.Duration
}

type ScanResultICMP struct {
	IP  string
	RTT time.Duration
}

// DefaultScan example: you can set desiredCIDR to "" to use interface mask,
// or "192.168.0.0/28" to request scanning that CIDR (must be inside interface network).
func DefaultScan(networkInterface string, targetCIDR string, ICMPMode bool) {

	if ICMPMode {
		ICMPScan(targetCIDR)
	} else {
		ArpScan(networkInterface, targetCIDR)
	}

}

func ArpScan(networkInterface string, targetCIDR string) {
	fmt.Println("Starting ARP scan...")

	// Find all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

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

	wg.Wait()
}

func ICMPScan(targetCIDR string) {
	if targetCIDR == "" {
		fmt.Println("No CIDR provided for ICMP scan, please provide a valid CIDR using the -c flag.")
		return
	}

	timeout = 2 * time.Second
	target := targetCIDR
	// --- Detect Single IP or CIDR ---
	if _, ipnet, err := net.ParseCIDR(target); err == nil {
		fmt.Printf("Target is a CIDR: %s (network %s)\n", target, ipnet.String())
		runSweep(target)
		fmt.Println("Ping sweep complete.")
	} else if ip := net.ParseIP(target); ip != nil {
		fmt.Printf("Target is a single IP: %s\n", target)
		wg.Add(1)
		go pingHost(target, 4)
		wg.Wait()
	} else {
		fmt.Println("Invalid input: not a valid IP or CIDR")
		return
	}
	printStats()
}

// runSweep handles CIDR sweeps
func runSweep(cidr string) {
	ips, err := hostsInCIDR(cidr)
	if err != nil {
		log.Fatalf("Error parsing CIDR: %v", err)
	}

	fmt.Printf("Starting ping sweep on %s ...\n", cidr)
	for _, ip := range ips {
		wg.Add(1)
		go pingHost(ip, 1)
	}

	wg.Wait()

}

// pingHost sends a single ICMP Echo Request
func pingHost(ip string, count int) {
	defer wg.Done()

	remoteAddr, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		return
	}

	conn, err := net.DialIP("ip4:icmp", nil, remoteAddr) // nil => let OS pick interface
	if err != nil {
		fmt.Println("Unable to open raw socket. Run as Administrator/Root.")
		return
	}
	defer conn.Close()

	for i := 1; i <= count; i++ {
		mu.Lock()
		stats.PacketsSent++
		mu.Unlock()
		// Build ICMP Echo Request
		icmp := &layers.ICMPv4{
			TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, 0),
			Id:       54321,
			Seq:      uint16(i),
		}

		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
		if err := icmp.SerializeTo(buf, opts); err != nil {
			log.Fatalf("Failed to serialize ICMP packet: %v", err)
		}

		start := time.Now()
		if _, err := conn.Write(buf.Bytes()); err != nil {
			log.Printf("Request %d failed to send: %v\n", i, err)
			continue
		}

		reply := make([]byte, 1500)
		_ = conn.SetReadDeadline(time.Now().Add(timeout))
		if _, _, err := conn.ReadFrom(reply); err == nil {
			rtt := time.Since(start)
			fmt.Printf("Host up: %-15s RTT=%v Seq=%v\n", ip, rtt, icmp.Seq)

			mu.Lock()
			result := ScanResultICMP{
				IP:  ip,
				RTT: rtt,
			}
			icmpscan_results = append(icmpscan_results, result)
			stats.PacketsReceived++
			stats.TotalRTT += rtt
			mu.Unlock()
		}

		time.Sleep(700 * time.Millisecond)
	}
}

// hostsInCIDR expands CIDR to a list of host IPs (excluding network/broadcast)
func hostsInCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network and broadcast
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}
	return ips, nil
}

// inc increments an IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// printStats shows summary of packet statistics
func printStats() {
	sent := stats.PacketsSent
	received := stats.PacketsReceived
	loss := 0.0
	if sent > 0 {
		loss = float64(sent-received) / float64(sent) * 100
	}

	fmt.Println("📊 --- Sweep Statistics ---")
	fmt.Printf("Packets Sent:     %d\n", sent)
	fmt.Printf("Packets Received: %d\n", received)
	fmt.Printf("Packet Loss:      %.1f%%\n", loss)

	if received > 0 {
		avgRTT := stats.TotalRTT / time.Duration(received)
		fmt.Printf("Average RTT:      %v\n", avgRTT)
	}
	fmt.Println("---------------------------")
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

	time.Sleep(3 * time.Second)
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
