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
	"context"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	probing "github.com/prometheus-community/pro-bing"
)

var (
	Defaultscan_results []ScanResultDfActive
	Icmpscan_results    []ScanResultICMP
	seenResults         = make(map[string]bool)
	mu                  sync.Mutex
	c                   = make(chan os.Signal, 1)
)

type ScanResultDfActive struct {
	Interface string
	Dest_IP   string
	Dest_Mac  string
}

type ScanResultICMP struct {
	IP  string
	RTT time.Duration
}

type HostnameResult struct {
	IP   string
	PTR  []string // reverse lookup names (may include trailing dot)
	FQDN string   // chosen FQDN (first PTR with trailing dot trimmed) or empty
	Err  string   // non-empty on error
}

// DefaultScan example: you can set desiredCIDR to "" to use interface mask,
// or "192.168.0.0/28" to request scanning that CIDR (must be inside interface network).
func DefaultScan(networkInterface string, targetCIDR string, ICMPMode bool, concurrency int, timeoutSec int, count int) {
	netiface, err := net.InterfaceByName(networkInterface)
	if err != nil {
		panic(err)
	}

	if ICMPMode {
		ICMPScan(netiface, targetCIDR, concurrency, timeoutSec, count)
	} else {
		ArpScan(netiface, targetCIDR, concurrency)
	}

	results := DiscoverHostnamesFromScanResults(Defaultscan_results, Icmpscan_results, 20, 2)
	if len(results) > 0 {
		fmt.Printf("\nDiscovered %d hostnames from scan results:\n", len(results))
	}
}

func ArpScan(networkInterface *net.Interface, targetCIDR string, concurrency int) {
	fmt.Println("Starting ARP scan...")
	var wg sync.WaitGroup
	// Find all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go func(netiface net.Interface) {
		defer wg.Done()
		if err := scan(&netiface, &devices, targetCIDR, concurrency); err != nil {
			log.Printf("interface %v: %v", netiface.Name, err)
		}
	}(*networkInterface)

	wg.Wait()
}

func ICMPScan(netiface *net.Interface, targetCIDR string, concurrency int, timeoutSec int, count int) {
	addr := parseNetIP(netiface)
	if addr == nil {
		fmt.Println("No valid IPv4 address found on the interface.")
		return
	}

	if targetCIDR == "" {
		targetCIDR = addr.String()
	}

	// Handle Ctrl+C
	signal.Notify(c, os.Interrupt)
	target := targetCIDR
	// --- Detect Single IP or CIDR ---
	if ip, ipnet, err := net.ParseCIDR(target); err == nil {
		fmt.Printf("Target is a CIDR: %s (network %s)\n", target, ipnet.String())
		runSweep(ip, ipnet, concurrency, count, time.Duration(timeoutSec)*time.Second)
		fmt.Println("Ping sweep complete.")

	} else if ip := net.ParseIP(target); ip != nil {
		fmt.Printf("Target is a single IP: %s\n", target)
		pingHost(target, count, time.Duration(timeoutSec)*time.Second)
	} else {
		fmt.Println("Invalid input: not a valid IP or CIDR")
	}
}

func runSweep(ip net.IP, ipNet *net.IPNet, concurrency int, count int, timeout time.Duration) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency) // limit parallel workers

	for currentIP := ip.Mask(ipNet.Mask); ipNet.Contains(currentIP); incIP(currentIP) {
		select {
		case <-c:
			fmt.Println("\nInterrupted.")
			return
		default:
			hostIP := currentIP.String()
			if hostIP == ipNet.IP.String() || isBroadcast(hostIP, ipNet) {
				continue
			}

			wg.Add(1)
			sem <- struct{}{} // acquire slot

			go func(target string) {
				defer wg.Done()
				defer func() { <-sem }() // release slot

				pinger, err := probing.NewPinger(target)
				if err != nil {
					return
				}
				pinger.SetPrivileged(true)

				pinger.Count = count
				pinger.Interval = time.Duration(100) * time.Millisecond
				pinger.Timeout = timeout
				if err := pinger.Run(); err == nil && pinger.Statistics().PacketsRecv > 0 {
					fmt.Printf("Host alive: %-15s (avg RTT: %v)\n",
						target, pinger.Statistics().AvgRtt)
					mu.Lock()
					Icmpscan_results = append(Icmpscan_results, ScanResultICMP{
						IP:  target,
						RTT: pinger.Statistics().AvgRtt,
					})
					mu.Unlock()
				}
			}(hostIP)
		}
	}

	wg.Wait()
}

// pingHost handles single-target ping with stats
func pingHost(target string, count int, timeout time.Duration) {
	pinger, err := probing.NewPinger(target)
	if err != nil {
		fmt.Printf("Cannot create pinger for %s: %v\n", target, err)
		return
	}
	pinger.SetPrivileged(true)

	pinger.Count = count
	pinger.Interval = time.Duration(100) * time.Millisecond
	pinger.Timeout = timeout
	pinger.OnRecv = func(pkt *probing.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.TTL)
	}
	if err := pinger.Run(); err != nil {
		fmt.Printf("Ping failed for %s: %v\n", target, err)
	}
}

// scan now accepts targetCIDR. If targetCIDR == "" it uses interface network as before.
func scan(iface *net.Interface, devices *[]pcap.Interface, targetCIDR string, concurrency int) error {
	addr := parseNetIP(iface)
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
	if err := writeARP(handle, iface, scanNet, addr, concurrency); err != nil {
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
				Defaultscan_results = append(Defaultscan_results, result)
			}
			mu.Unlock()

			log.Printf("IP %v is at %v from interface: %v",
				result.Dest_IP, result.Dest_Mac, result.Interface)
		}
	}
}

// writeARP writes an ARP request for each address on our local network to the
// pcap handle.
// writeARPConcurrent writes ARP requests concurrently (bounded by `concurrency`).
// It is a drop-in replacement for writeARP but faster for large subnets.
func writeARP(handle *pcap.Handle, iface *net.Interface, addr *net.IPNet, intAddr *net.IPNet, concurrency int) error {
	// if concurrency <= 0 {
	// 	concurrency = 64 // sensible default, tune as needed
	// }

	// Mutex to protect WritePacketData in case the pcap handle implementation is not goroutine-safe.
	var writeMu sync.Mutex
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	errCh := make(chan error, 1) // capture first error

	// iterate all IPs to scan
	for _, ip := range ips(addr, intAddr) {
		// prepare per-ip values early
		ip := ip // capture loop variable

		// skip same as before (network/broadcast etc handled in ips)
		sem <- struct{}{}
		wg.Add(1)

		go func(targetIP net.IP) {
			defer wg.Done()
			defer func() { <-sem }()

			// Build ethernet + arp into a fresh buffer per goroutine (no shared buffer)
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
				SourceProtAddress: []byte(intAddr.IP),
				DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
				DstProtAddress:    []byte(targetIP.To4()),
			}

			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}
			if err := gopacket.SerializeLayers(buf, opts, &eth, &arp); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			out := buf.Bytes()

			// Serialize writes with mutex to be safe on non-threadsafe pcap handles
			writeMu.Lock()
			if err := handle.WritePacketData(out); err != nil {
				writeMu.Unlock()
				select {
				case errCh <- err:
				default:
				}
				return
			}
			writeMu.Unlock()
		}(ip)
	}

	// wait for all goroutines
	wg.Wait()
	close(errCh)
	// if any error occurred, return first one
	if err, ok := <-errCh; ok {
		return err
	}
	return nil
}

func DiscoverHostnamesFromScanResults(arp []ScanResultDfActive, icmp []ScanResultICMP, concurrency int, timeoutSec int) []HostnameResult {
	seen := make(map[string]struct{})
	var ips []string

	for _, r := range arp {
		if r.Dest_IP == "" {
			continue
		}
		if _, ok := seen[r.Dest_IP]; !ok {
			seen[r.Dest_IP] = struct{}{}
			ips = append(ips, r.Dest_IP)
		}
	}
	for _, r := range icmp {
		if r.IP == "" {
			continue
		}
		if _, ok := seen[r.IP]; !ok {
			seen[r.IP] = struct{}{}
			ips = append(ips, r.IP)
		}
	}

	if concurrency <= 0 {
		concurrency = 10
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	resChan := make(chan HostnameResult, len(ips))

	resolver := &net.Resolver{PreferGo: false}

	for _, ip := range ips {
		wg.Add(1)
		sem <- struct{}{}
		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
			defer cancel()

			var hr HostnameResult
			hr.IP = ip

			names, err := resolver.LookupAddr(ctx, ip)
			if err != nil {
				hr.Err = err.Error()
			} else if len(names) > 0 {
				for i := range names {
					names[i] = strings.TrimSuffix(names[i], ".")
				}
				hr.PTR = names
				hr.FQDN = names[0]
			}

			// --- Print immediately ---
			if hr.FQDN != "" {
				fmt.Printf("[+] %s -> %s\n", hr.IP, hr.FQDN)
			} else if hr.Err != "" {
				fmt.Printf("[-] %s lookup failed: %s\n", hr.IP, hr.Err)
			} else {
				fmt.Printf("[*] %s has no PTR record\n", hr.IP)
			}

			resChan <- hr
		}(ip)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	var results []HostnameResult
	for r := range resChan {
		results = append(results, r)
	}

	return results
}

//Helper functions for ARP

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

func parseNetIP(iface *net.Interface) *net.IPNet {
	var addr *net.IPNet
	if addrs, err := iface.Addrs(); err != nil {
		return nil
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
	return addr
}

// Helper functions for ICMP
// incIP increments an IP address by 1
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// isBroadcast checks if IP is the broadcast address of the subnet
func isBroadcast(ipStr string, ipNet *net.IPNet) bool {
	ip := net.ParseIP(ipStr).To4()
	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = ipNet.IP[i] | ^ipNet.Mask[i]
	}
	return ip.Equal(broadcast)
}

// Helper functions for reverse DNS and saving results
// SaveHostnamesCSV writes the results to a CSV file with columns: ip,fqdn,ptrs,error
func SaveHostnamesCSV(filename string, results []HostnameResult) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"ip", "fqdn", "ptrs", "error"}); err != nil {
		return err
	}

	for _, r := range results {
		ptrs := strings.Join(r.PTR, ";")
		if err := w.Write([]string{r.IP, r.FQDN, ptrs, r.Err}); err != nil {
			return err
		}
	}
	return nil
}
