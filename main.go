package main

import (
	"github.com/Naman1997/discovr/cmd"
)

func main() {
	cmd.Execute()
}


// package main


// import (
// 	"context"
// 	"flag"
// 	"fmt"
// 	"github.com/google/gopacket/layers"
// 	"golang.org/x/net/ipv4"
// 	"log"
// 	"math/rand"
// 	"net"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/google/gopacket"
// 	"github.com/google/gopacket/pcap"
// )

// var (
// 	wg      = &sync.WaitGroup{}
// 	device  = flag.String("i", "wlp0s20f3", "bind interface")
// 	express = flag.String("r", "host 10.136.148.255", "pcap expression")
// )

// func main() {
// 	flag.Parse()
// 	log.Println("Started pcap open:", *device)
// 	defer log.Println("Finished")

// 	ctx, cancel := context.WithCancel(context.Background())
// 	go capturePackets(ctx, wg, *device, *express)

// 	time.Sleep(100000 * time.Millisecond)
// 	log.Println("Attempting to close the pcap handle and expecting the packets channel to be closed soon.")
// 	cancel()

// 	wg.Wait()
// }

// func capturePackets(ctx context.Context, wg *sync.WaitGroup, networkInterface, bpfFilter string) {
// 	wg.Add(1)
// 	defer wg.Done()
// 	defer log.Println("The gopacket.PacketSources.Packets() channel was closed.")

// 	for packet := range packets(ctx, wg, networkInterface, bpfFilter) {
// 		log.Print(packet)
// 	}
// }

// func packets(ctx context.Context, wg *sync.WaitGroup, networkInterface, bpfFilter string) chan gopacket.Packet {
// 	log.Printf("Listening on '%s', with '%s'", networkInterface, bpfFilter)
// 	if handle, err := pcap.OpenLive(networkInterface, 1024, false, pcap.BlockForever); err != nil {
// 		panic(err)
// 	} else if err = handle.SetBPFFilter(bpfFilter); err != nil {
// 		panic(err)
// 	} else {
// 		ps := gopacket.NewPacketSource(handle, handle.LinkType())

// 		closed := false
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			<-ctx.Done()
// 			log.Println("Closing the pcap handle.")
// 			handle.Close()
// 			closed = true
// 			log.Println("Closed the pcap handle.")
// 		}()

// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			var (
// 				localIP, remoteIP net.IP
// 			)
// 			src := strings.Fields(bpfFilter)[1]
// 			remoteIP = net.ParseIP(src)
// 			if remoteIP == nil {
// 				panic(fmt.Errorf("src:%s parse failed", src))
// 			}
// 			ifaces, err := net.Interfaces()
// 			if err != nil {
// 				panic(err)
// 			}

// 			// get localIP with device
// 			{
// 			LOOP:
// 				for _, iface := range ifaces {
// 					if iface.Name == *device {
// 						addrs, err := iface.Addrs()
// 						if err != nil {
// 							panic(err)
// 						}
// 						for _, addr := range addrs {
// 							switch addr := addr.(type) {
// 							case *net.IPNet:
// 								if remoteIP.To4() != nil && addr.IP.To4() != nil {
// 									localIP = addr.IP
// 								}
// 								if remoteIP.To16() != nil && addr.IP.To16() != nil {
// 									localIP = addr.IP
// 								}
// 								break LOOP
// 							}
// 						}
// 					}
// 				}
// 			}
// 			if localIP == nil {
// 				panic(fmt.Errorf("can not get localIP with remote:%v", remoteIP))
// 			}
// 			// Another goroutine is preparing to close the pcap and send a fake response packet later.
// 			<-ctx.Done()
// 			time.Sleep(1000 * time.Millisecond)
// 			if closed {
// 				return
// 			}

// 			tcpConn, err := net.ListenPacket("ip4:tcp", localIP.String())
// 			if err != nil {
// 				panic(err)
// 			}
// 			defer tcpConn.Close()

// 			// mock local send request
// 			localPort := 1000
// 			remotePort := 80
// 			log.Printf("mock request, %v:%d -> %v:%d", localIP, localPort, remoteIP, remotePort)
// 			err = sendMessage(tcpConn, 10, uint32(rand.Int()), localPort, remotePort, localIP, remoteIP)
// 			if err != nil {
// 				panic(err)
// 			}

// 			// mock remote response
// 			log.Printf("mock response, %v:%d -> %v:%d", remoteIP, remotePort, localIP, localPort)
// 			err = sendMessage(tcpConn, 10, uint32(rand.Int()), remotePort, localPort, remoteIP, localIP)
// 			if err != nil {
// 				panic(err)
// 			}
// 		}()
// 		return ps.Packets()
// 	}
// }

// func sendMessage(tcpConn net.PacketConn, ttl uint16, sequenceNumber uint32, srcPort, dstPort int, srcIP, dstIP net.IP) error {
// 	ipHeader := &layers.IPv4{
// 		SrcIP:    srcIP,
// 		DstIP:    dstIP,
// 		Protocol: layers.IPProtocolTCP,
// 		TTL:      uint8(ttl),
// 	}

// 	tcpHeader := &layers.TCP{
// 		SrcPort: layers.TCPPort(srcPort),
// 		DstPort: layers.TCPPort(dstPort),
// 		Seq:     sequenceNumber,
// 		SYN:     true,
// 		Window:  14600,
// 	}
// 	_ = tcpHeader.SetNetworkLayerForChecksum(ipHeader)

// 	buf := gopacket.NewSerializeBuffer()
// 	opts := gopacket.SerializeOptions{
// 		ComputeChecksums: true,
// 		FixLengths:       true,
// 	}
// 	if err := gopacket.SerializeLayers(buf, opts, tcpHeader); err != nil {
// 		return err
// 	}

// 	if err := ipv4.NewPacketConn(tcpConn).SetTTL(int(ttl)); err != nil {
// 		return err
// 	}

// 	if _, err := tcpConn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstIP}); err != nil {
// 		return err
// 	}

// 	return nil
// }