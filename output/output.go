package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func storedata(data [][]string) error {

	file, err := os.Create("data.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.WriteAll(data)
	if err != nil {
		fmt.Println("Error writing CSV:", err)
	}
	return nil
}

func main() {

	scandata := [][]string{{"HostName", "IP"}, {"PC_12", "192.168.0.2"}}

	if err := storedata(scandata); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("CSV file 'output.csv' created successfully!")
	}
}
