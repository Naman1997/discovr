package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

func ExportCSV[T any](filePath string, data []T) error {
	if filePath == "" {
		return nil
	}

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

	v := reflect.ValueOf(data)
	elemType := reflect.TypeOf(data).Elem()

	var headers []string
	for i := 0; i < elemType.NumField(); i++ {
		headers = append(headers, elemType.Field(i).Name)
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		var record []string
		for j := 0; j < elem.NumField(); j++ {
			val := elem.Field(j).Interface()
			record = append(record, fmt.Sprint(val))
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	fmt.Printf("Saved to: %v\n", filePath)
	return nil
}

func UploadResults[T any](url string, filePath string, data []T, filePrefix string) {
	if url == "" {
		return
	}
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// If file doesn't exist, create it in the temp directory
		tempFilePath := filepath.Join(os.TempDir(), filePrefix+time.Now().Format("20060102_150405")+".csv")
		err := ExportCSV(tempFilePath, data)
		if err != nil {
			panic(err)
		}

		filePath = tempFilePath
		defer os.Remove(filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		panic(err)
	}

	// Copy file content to form field
	if _, err = io.Copy(fw, file); err != nil {
		panic(err)
	}
	w.Close()

	// Create the request
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}