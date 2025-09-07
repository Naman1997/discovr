package internal

import (
	"fmt"
)

func ActiveScan(targets string, ports string, osDetection bool, nmapScan bool) []string {

	if nmapScan {
		//It will run nmap scan
		fmt.Println("It will run nmap scan")
		NmapScan(targets, ports, osDetection)

		//export header for nmap scan
		return []string{"ID", "Protocol", "State", "Service", "Product"}
	} else {

		//It will run default scan
		fmt.Println("It will run default scan")
		DefaultScan()

		//export header for default scan
		return []string{"Date", "Time", "Interface", "Dest_IP", "Dest_Mac"}
	}

}
