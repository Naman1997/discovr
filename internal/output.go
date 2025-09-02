package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

func Storedata(file_path string, head []string, data []ScanResultPassive) error {
	fmt.Println("\n")
	if filepath.Ext(file_path) != ".csv" {
		file_path = file_path + ".csv" // or return error instead
		fmt.Println("Export path did not have .csv extension, saving as:", file_path)
	}

	// Check if file exists
	_, err := os.Stat(file_path)
	fileExists := err == nil

	// Append to file if it exists, create new file if it does not exist, raise error if errors
	file, err := os.OpenFile(file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error! %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if file is new
	if !fileExists {
		if err := writer.Write(head); err != nil {
			fmt.Printf("Error! %v", err)
		}
	}

	// Write new data rows
	for _, r := range data {
		if err := writer.Write([]string{r.SrcIP, r.Protocol, r.SrcMAC, r.DstMAC, r.EthernetType}); err != nil {
			fmt.Printf("Error! %v", err)
		}
	}
	return nil
}

func PassiveExport(path string) {
	if path != "" {
		passiveheader := []string{"Source_IP", "Protocol", "Source_MAC", "Destination_Mac", "Ethernet_Type"}
		Storedata(path, passiveheader, passive_results)
	} else {
		return
	}
}

func Storedata_A(file_path string, head []string, data []ScanResultActive) error {
	fmt.Println("\n")
	if filepath.Ext(file_path) != ".csv" {
		file_path = file_path + ".csv" // or return error instead
		fmt.Println("Export path did not have .csv extension, saving as:", file_path)
	}

	// Check if file exists
	_, err := os.Stat(file_path)
	fileExists := err == nil

	// Append to file if it exists, create new file if it does not exist, raise error if errors
	file, err := os.OpenFile(file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error! %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if file is new
	if !fileExists {
		if err := writer.Write(head); err != nil {
			fmt.Printf("Error! %v", err)
		}
	}

	// Write new data rows
	for _, r := range data {
		if err := writer.Write([]string{r.ID, r.Protocol, r.State, r.Service, r.Product}); err != nil {
			fmt.Printf("Error! %v", err)
		}
	}
	return nil
}

func ActiveExport(path string) {
	if path != "" {
		passiveheader := []string{"ID", "Protocol", "Source_MAC", "Destination_Mac", "Ethernet_Type"}
		Storedata_A(path, passiveheader, active_results)
	} else {
		return
	}
}
