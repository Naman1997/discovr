package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

func Export(filePath string, header []string, rows [][]string) error {
	if filepath.Ext(filePath) != ".csv" {
		filePath = filePath + ".csv"
		fmt.Println("\nExport path did not have .csv extension, saving as:", filePath)
	}

	originalPath := filePath
	count := 1
	for {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		ext := filepath.Ext(originalPath)
		name := originalPath[:len(originalPath)-len(ext)]
		filePath = fmt.Sprintf("%s_%d%s", name, count, ext)
		count++
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %v", err)
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %v", err)
		}
	}
	fmt.Printf("Saved to: %v\n", filePath)
	return nil
}

// Convert Passive results
func PassiveExport(path string) {
	header := []string{"Source_IP", "Protocol", "Source_MAC", "Destination_Mac", "Ethernet_Type"}
	if path == "" {
		return
	}
	rows := make([][]string, len(passive_results))
	for i, r := range passive_results {
		rows[i] = []string{r.SrcIP, r.Protocol, r.SrcMAC, r.DstMAC, r.EthernetType}
	}
	Export(path, header, rows)
}

// Convert Active results
func ActiveExport(path string, mode bool) {
	fmt.Println(mode)
	if path == "" {
		return
	}
	//Changes for scrum-136
	if mode {
		// for icmp scan
		header := []string{"IP", "RTT"}
		rows := make([][]string, len(icmpscan_results))
		for i, r := range icmpscan_results {
			rows[i] = []string{r.IP, r.RTT.String()}
		}
		Export(path, header, rows)
		return
	} else {
		// for default scan
		header := []string{"Interface", "Dest_IP", "Dest_Mac"}
		rows := make([][]string, len(defaultscan_results))
		for i, r := range defaultscan_results {
			rows[i] = []string{r.Interface, r.Dest_IP, r.Dest_Mac}
		}
		Export(path, header, rows)
	}
}

func NmapExport(path string) {
	if path == "" {
		return
	}
	header := []string{"Port", "Protocol", "State", "Service", "Product"}
	rows := make([][]string, len(active_results))
	for i, r := range active_results {
		rows[i] = []string{r.Port, r.Protocol, r.State, r.Service, r.Product}
	}
	Export(path, header, rows)
}

// Convert AWS results
func AwsExport(path string) {
	header := []string{"InstanceId", "PublicIp", "PrivateIPs", "MacAddress", "VpcId", "SubnetId", "Hostname", "Region"}
	if path == "" {
		return
	}
	rows := make([][]string, len(aws_results))
	for i, r := range aws_results {
		rows[i] = []string{r.InstanceId, r.PublicIp, r.PrivateIPs, r.MacAddress, r.VpcId, r.SubnetId, r.Hostname, r.Region}
	}
	Export(path, header, rows)
}

// Convert Azure results
func AzureExport(path string) {
	header := []string{"VM Name", "ID", "Location", "Resource Group", "NIC", "MAC", "SubnetID", "Vnet", "PrivateIP", "Public IP"}
	if path == "" {
		return
	}
	rows := make([][]string, len(azure_results))
	for i, r := range azure_results {
		rows[i] = []string{r.Name, r.UniqueID, r.Location, r.ResourceGroup, r.MAC, r.Subnet, r.Vnet, r.PrivateIP, r.PublicIP}
	}
	Export(path, header, rows)
}

// Convert GCP results
func GcpExport(path string) {
	header := []string{"ProjectId", "InstanceId", "InstanceName", "Hostname", "OS", "Zone", "InterfaceName", "InternalIp", "ExternalIps", "VPC", "Subnet"}
	if path == "" {
		return
	}
	rows := make([][]string, len(gcp_results))
	for i, r := range gcp_results {
		rows[i] = []string{r.ProjectId, r.InstanceId, r.InstanceName, r.Hostname, r.OsType, r.Zone, r.InterfaceName, r.InternalIP, r.ExternalIPs, r.VPC, r.Subnet}
	}
	Export(path, header, rows)
}
