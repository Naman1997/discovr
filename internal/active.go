package internal

import (
	"archive/zip"
	"fmt"
	"github.com/Naman1997/discovr/assets"
	"github.com/joho/godotenv"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func ActiveScan() {

	nmapPath := extractNmap()
	fmt.Printf(nmapPath)
}

func extractNmap() string {
	nmapVersion := getEnvVariable("NMAP_VERSION")
	nmapBinaryName := "nmap"
	nmapExeName := nmapBinaryName + ".exe"
	nmapVersionedZip := "nmap-" + nmapVersion + "-win32.zip"
	extractedFolderName := nmapBinaryName + "-" + nmapVersion

	bin2, _ := assets.Assets.ReadFile(nmapVersionedZip)
	tmpDir, _ := os.MkdirTemp("", "discovr-embedded-bin-*")
	tmpPath := filepath.Join(tmpDir, nmapVersionedZip)
	_ = os.WriteFile(tmpPath, bin2, 0644)

	// Unzip and delete zip file
	unzip(tmpDir, tmpPath)
	_ = os.Remove(tmpDir + string(os.PathSeparator) + nmapVersionedZip)

	// If linux, copy the linux binary to the extracted folder
	if runtime.GOOS == "linux" {
		nmapLinuxBinary, _ := assets.Assets.ReadFile(nmapBinaryName)
		binPath := tmpDir + string(os.PathSeparator) + extractedFolderName + string(os.PathSeparator) + nmapBinaryName
		tmpPath = filepath.Join(binPath)
		_ = os.WriteFile(tmpPath, nmapLinuxBinary, 0755)
		return binPath
	}

	return tmpDir + string(os.PathSeparator) + extractedFolderName + string(os.PathSeparator) + nmapExeName
}

func unzip(destination string, zipFilePath string) {
	archive, err := zip.OpenReader(zipFilePath)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(destination, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(string(os.PathSeparator))) {
			fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}

func getEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	// TODO: Fix this handling
	if err != nil {
		fmt.Printf("Error loading .env file")
	}

	return os.Getenv(key)
}
