package internal

import (
	"context"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"net"
	"slices"
	"sync"
	"time"
)

var (
	wg               = &sync.WaitGroup{}
	discoveredAssets = []string{}
)

func PassiveScan(device string, scanSeconds int) {

	// Initialize context and define scanDuration
	var scanDuration time.Duration = time.Duration(scanSeconds) * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	go capturePackets(ctx, wg, device, scanDuration)

	// Wait for the scanDuration and wg to finish
	time.Sleep(scanDuration)
	cancel()
	wg.Wait()
}

func capturePackets(ctx context.Context, wg *sync.WaitGroup, networkInterface string, scanDuration time.Duration) {
	wg.Add(1)
	defer wg.Done()

	// Fetch local ip addresses
	localIPs, err := getLocalIPs()
	if err != nil {
		panic(err)
	}

	// Creating a ticker to manually stop the for loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	timeout := time.After(scanDuration)

	packets := packets(ctx, wg, networkInterface)
	for {
		select {
		case packet := <-packets:
			printPacketInfo(packet, localIPs)
		case <-timeout:
			return
		}
	}
}

func packets(ctx context.Context, wg *sync.WaitGroup, networkInterface string) chan gopacket.Packet {
	if handle, err := pcap.OpenLive(networkInterface, 1024, false, pcap.BlockForever); err != nil {
		panic(err)
	} else {
		ps := gopacket.NewPacketSource(handle, handle.LinkType())
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			handle.Close()
		}()
		return ps.Packets()
	}
}

// TODO: Wait for SRUM-8 and implement the method to export this information to a csv file
func printPacketInfo(packet gopacket.Packet, localIPs []string) {
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		if slices.Contains(localIPs, ip.DstIP.String()) && !slices.Contains(discoveredAssets, ip.SrcIP.String()) {
			discoveredAssets = append(discoveredAssets, ip.SrcIP.String())
			fmt.Printf("Discovered new asset: %s\n", ip.SrcIP)
			fmt.Println("Protocol: ", ip.Protocol)
			fmt.Println()

			if ethernetLayer != nil {
				fmt.Println("Ethernet layer detected.")
				ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
				fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
				fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
				fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
				fmt.Println()
			}
			fmt.Println("==========================================================================================")
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
