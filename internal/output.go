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
		fmt.Println("Export path did not have .csv extension, saving as:", filePath)
	}

	// Check if file exists
	_, err := os.Stat(filePath)
	fileExists := err == nil

	// Open or create file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header only if file is new
	if !fileExists {
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("error writing header: %v", err)
		}
	}

	// Write rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %v", err)
		}
	}

	return nil
}

// Convert Passive results
func PassiveExport(path string, header []string) {
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
func ActiveExport(path string, header []string, nmapScan bool) {
	if path == "" {
		return
	}
	//Changes for scrum-136
	if nmapScan {
		rows := make([][]string, len(active_results))
		for i, r := range active_results {
			rows[i] = []string{r.Port, r.Protocol, r.State, r.Service, r.Product}
		}
		Export(path, header, rows)
	} else {
		// for default scan
		rows := make([][]string, len(defaultscan_results))
		for i, r := range defaultscan_results {
			rows[i] = []string{r.Interface, r.Dest_IP, r.Dest_Mac}
		}
		Export(path, header, rows)
		return
	}
}

// Convert AWS results
func AwsExport(path string, header []string) {
	if path == "" {
		return
	}
	rows := make([][]string, len(aws_results))
	for i, r := range aws_results {
		rows[i] = []string{r.InstanceId, r.PublicIp, r.PrivateIPs, r.MacAddress, r.VpcId, r.SubnetId, r.Hostname, r.Region}
	}
	Export(path, header, rows)
}
