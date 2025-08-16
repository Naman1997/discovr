package internal

import (
	"context"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
	"strings"
	"time"
	"sync"
)

var (
	wg      = &sync.WaitGroup{}
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

	// Creating a ticker to manually stop the for loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	timeout := time.After(scanDuration)

	packets := packets(ctx, wg, networkInterface)
	for {
		select {
			case packet := <-packets:
				printPacketInfo(packet)
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

/* 
Source: https://github.com/tgogos/gopacket_pcap/blob/master/decode/decode.go
TODO: Need to modify the decode example to extract source addresses of packets
*/
func printPacketInfo(packet gopacket.Packet) {
	// Let's see if the packet is an ethernet packet
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		fmt.Println("Ethernet layer detected.")
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
		fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
		fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
		// Ethernet type is typically IPv4 but could be ARP or other
		fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
		fmt.Println()
	}

	// Let's see if the packet is IP (even though the ether type told us)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		fmt.Println("IPv4 layer detected.")
		ip, _ := ipLayer.(*layers.IPv4)

		// IP layer variables:
		// Version (Either 4 or 6)
		// IHL (IP Header Length in 32-bit words)
		// TOS, Length, Id, Flags, FragOffset, TTL, Protocol (TCP?),
		// Checksum, SrcIP, DstIP
		fmt.Printf("From %s to %s\n", ip.SrcIP, ip.DstIP)
		fmt.Println("Protocol: ", ip.Protocol)
		fmt.Println()
	}

	// Let's see if the packet is TCP
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		fmt.Println("TCP layer detected.")
		tcp, _ := tcpLayer.(*layers.TCP)

		// TCP layer variables:
		// SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum, Urgent
		// Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
		fmt.Printf("From port %d to %d\n", tcp.SrcPort, tcp.DstPort)
		fmt.Println("Sequence number: ", tcp.Seq)
		fmt.Println()
	}

	// Iterate over all layers, printing out each layer type
	fmt.Println("All packet layers:")
	for _, layer := range packet.Layers() {
		fmt.Println("- ", layer.LayerType())
	}

	// When iterating through packet.Layers() above,
	// if it lists Payload layer then that is the same as
	// this applicationLayer. applicationLayer contains the payload
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		fmt.Println("Application layer/Payload found.")
		fmt.Printf("%s\n", applicationLayer.Payload())

		// Search for a string inside the payload
		if strings.Contains(string(applicationLayer.Payload()), "HTTP") {
			fmt.Println("HTTP found!")
		}
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("Error decoding some part of the packet:", err)
	}
}