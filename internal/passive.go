package internal

import (
	"context"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/Naman1997/discovr/verbose"
	Verbose "github.com/Naman1997/discovr/verbose"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"golang.org/x/sync/semaphore"
)

var (
	sem              = semaphore.NewWeighted(2)
	discoveredAssets = []string{}
	Passive_results  []ScanResultPassive
)

// export vars
type ScanResultPassive struct {
	SrcIP        string
	Protocol     string
	SrcMAC       string
	DstMAC       string
	EthernetType string
}

func PassiveScan(device string, scanSeconds int) {

	// Initialize context and define scanDuration
	var scanDuration time.Duration = time.Duration(scanSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), scanDuration)
	go capturePackets(ctx, sem, device, scanDuration)

	// Wait for the scanDuration and wg to finish
	time.Sleep(scanDuration)
	defer cancel()

	err := sem.Acquire(ctx, 2)
	if err != nil {
		fmt.Println("")
		return
	}
}

func capturePackets(ctx context.Context, sem *semaphore.Weighted, networkInterface string, scanDuration time.Duration) {
	err := sem.Acquire(context.Background(), 1)
	if err != nil {
		panic(err)
	}
	defer sem.Release(1)

	// Fetch local ip addresses
	localIPs, err := getLocalIPs()
	if err != nil {
		panic(err)
	}

	// Creating a ticker to manually stop the for loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	timeout := time.After(scanDuration)

	packets := packets(ctx, sem, networkInterface)
	for {
		select {
		case packet := <-packets:
			printPacketInfo(packet, localIPs)
		case <-timeout:
			return
		}
	}
}

func packets(ctx context.Context, sem *semaphore.Weighted, networkInterface string) chan gopacket.Packet {
	if handle, err := pcap.OpenLive(networkInterface, 1024, false, pcap.BlockForever); err != nil {
		panic(err)
	} else {
		ps := gopacket.NewPacketSource(handle, handle.LinkType())
		err := sem.Acquire(context.Background(), 1)
		if err != nil {
			panic(err)
		}
		defer sem.Release(1)
		go func() {
			<-ctx.Done()
			handle.Close()
		}()
		return ps.Packets()
	}
}

// TODO: Wait for SRUM-8 and implement the method to export this information to a csv file
func printPacketInfo(packet gopacket.Packet, localIPs []string) {
	if packet == nil {
		return
	}
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		if slices.Contains(localIPs, ip.DstIP.String()) && !slices.Contains(discoveredAssets, ip.SrcIP.String()) {
			discoveredAssets = append(discoveredAssets, ip.SrcIP.String())
			verbose.VerbosePrintf("Discovered new asset: %s\n", ip.SrcIP)
			verbose.VerbosePrintln("Protocol: ", ip.Protocol)
			verbose.VerbosePrintln()

			if ethernetLayer != nil {
				verbose.VerbosePrintln("Ethernet layer detected.\n")
				ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
				verbose.VerbosePrintln("Source MAC: ", ethernetPacket.SrcMAC)
				verbose.VerbosePrintln("Destination MAC: ", ethernetPacket.DstMAC)
				verbose.VerbosePrintln("Ethernet type: ", ethernetPacket.EthernetType)
				verbose.VerbosePrintln()

				//export SCRUM-94
				result := ScanResultPassive{
					SrcIP:        ip.SrcIP.String(),
					Protocol:     ip.Protocol.String(),
					SrcMAC:       ethernetPacket.SrcMAC.String(),
					DstMAC:       ethernetPacket.DstMAC.String(),
					EthernetType: ethernetPacket.EthernetType.String(),
				}
				Passive_results = append(Passive_results, result)

			}
			Verbose.VerbosePrintln("==========================================================================================")
		}
	}
}

func getLocalIPs() ([]string, error) {
	var localIPs []string
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error getting interfaces: %v\n", err)
		return localIPs, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Printf("Error getting addresses for interface %s: %v\nContinuing...\n", i.Name, err)
			continue
		}

		// fmt.Printf("Interface: %s (Flags: %s, HardwareAddr: %s)\n", i.Name, i.Flags, i.HardwareAddr)
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				localIPs = append(localIPs, ipnet.IP.String())
			}
		}
	}
	return localIPs, nil
}
