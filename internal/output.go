package internal

import (
	"encoding/csv"
	"fmt"
	"os"
)

func Storedata(filepath string, head []string, data [][]string) error {

	// Check if file exists
	_, err := os.Stat(filepath)
	fileExists := err == nil

	// Append to file if it exists, create new file if it does not exist, raise error if errors
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if file is new
	if !fileExists {
		if err := writer.Write(head); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write new data rows
	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}
	return nil
}
