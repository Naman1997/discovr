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
	"golang.org/x/net/ipv4"
	"math/rand"
	"net"
)

var (
	wg      = &sync.WaitGroup{}
)

/* The handler.OpenLive function is broken in upstream. The handler refuses to close.
Fixing by sending a fake response to the network card of the specified device
Issue: https://github.com/google/gopacket/issues/862
Fix: https://github.com/google/gopacket/issues/862#issuecomment-2490795103
TODO try alternative solution: https://github.com/google/gopacket/issues/1089#issuecomment-1501148868
*/
func PassiveScan(device string, scanTime time.Duration) {
	log.Println("Started pcap open:", device)
	defer log.Println("Finished")

	ctx, cancel := context.WithCancel(context.Background())
	go capturePackets(ctx, wg, device)

	time.Sleep(scanTime)
	log.Println("Attempting to close the pcap handle and expecting the packets channel to be closed soon.")
	cancel()

	wg.Wait()
}

func capturePackets(ctx context.Context, wg *sync.WaitGroup, networkInterface string) {
	wg.Add(1)
	defer wg.Done()
	defer log.Println("The gopacket.PacketSources.Packets() channel was closed.")

	for packet := range packets(ctx, wg, networkInterface) {
		printPacketInfo(packet)
	}
}

func packets(ctx context.Context, wg *sync.WaitGroup, networkInterface string) chan gopacket.Packet {
	localIP := getLocalIP(networkInterface)
	bpfFilter := "host " + localIP
	log.Printf("Listening on '%s', with '%s'", networkInterface, bpfFilter)
	if handle, err := pcap.OpenLive(networkInterface, 1024, false, pcap.BlockForever); err != nil {
		panic(err)
	} else if err = handle.SetBPFFilter(bpfFilter); err != nil {
		panic(err)
	} else {
		ps := gopacket.NewPacketSource(handle, handle.LinkType())

		closed := false
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			log.Println("Closing the pcap handle.")
			handle.Close()
			closed = true
			log.Println("Closed the pcap handle.")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				localIP, remoteIP net.IP
			)
			src := strings.Fields(bpfFilter)[1]
			remoteIP = net.ParseIP(src)
			if remoteIP == nil {
				panic(fmt.Errorf("src:%s parse failed", src))
			}
			ifaces, err := net.Interfaces()
			if err != nil {
				panic(err)
			}

			// TODO: Clean up this mess
			// get localIP with device
			{
			LOOP:
				for _, iface := range ifaces {
					if iface.Name == networkInterface {
						addrs, err := iface.Addrs()
						if err != nil {
							panic(err)
						}
						for _, addr := range addrs {
							switch addr := addr.(type) {
							case *net.IPNet:
								if remoteIP.To4() != nil && addr.IP.To4() != nil {
									localIP = addr.IP
								}
								if remoteIP.To16() != nil && addr.IP.To16() != nil {
									localIP = addr.IP
								}
								break LOOP
							}
						}
					}
				}
			}
			if localIP == nil {
				panic(fmt.Errorf("can not get localIP with remote:%v", remoteIP))
			}
			// Another goroutine is preparing to close the pcap and send a fake response packet later.
			<-ctx.Done()
			time.Sleep(1000 * time.Millisecond)
			if closed {
				return
			}

			tcpConn, err := net.ListenPacket("ip4:tcp", localIP.String())
			if err != nil {
				panic(err)
			}
			defer tcpConn.Close()

			// mock local send request
			localPort := 1000
			remotePort := 80
			log.Printf("mock request, %v:%d -> %v:%d", localIP, localPort, remoteIP, remotePort)
			err = sendMessage(tcpConn, 10, uint32(rand.Int()), localPort, remotePort, localIP, remoteIP)
			if err != nil {
				panic(err)
			}

			// mock remote response
			log.Printf("mock response, %v:%d -> %v:%d", remoteIP, remotePort, localIP, localPort)
			err = sendMessage(tcpConn, 10, uint32(rand.Int()), remotePort, localPort, remoteIP, localIP)
			if err != nil {
				panic(err)
			}
		}()
		return ps.Packets()
	}
}

func sendMessage(tcpConn net.PacketConn, ttl uint16, sequenceNumber uint32, srcPort, dstPort int, srcIP, dstIP net.IP) error {
	ipHeader := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: layers.IPProtocolTCP,
		TTL:      uint8(ttl),
	}

	tcpHeader := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		Seq:     sequenceNumber,
		SYN:     true,
		Window:  14600,
	}
	_ = tcpHeader.SetNetworkLayerForChecksum(ipHeader)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buf, opts, tcpHeader); err != nil {
		return err
	}

	if err := ipv4.NewPacketConn(tcpConn).SetTTL(int(ttl)); err != nil {
		return err
	}

	if _, err := tcpConn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstIP}); err != nil {
		return err
	}

	return nil
}

/* 
Source: https://github.com/tgogos/gopacket_pcap/blob/master/decode/decode.go
Need to modify the decode example to extract source addresses of packets
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

// TODO: Clean up this mess
func getLocalIP(device string) string {
	
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error getting interfaces: %v\n", err)
		return ""
	}

	for _, i := range ifaces {
		if i.Name == device {
			addrs, err := i.Addrs()
			if err != nil {
				fmt.Printf("Error getting addresses for interface %s: %v\n", i.Name, err)
				continue
			}

			fmt.Printf("Interface: %s (Flags: %s, HardwareAddr: %s)\n", i.Name, i.Flags, i.HardwareAddr)
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					fmt.Printf("  IP Address: %s\n", ipnet.IP.String())
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}