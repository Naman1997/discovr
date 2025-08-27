package internal

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/Naman1997/discovr/assets"
	"github.com/Ullaakut/nmap/v3"
	osfamily "github.com/Ullaakut/nmap/v3/pkg/osfamilies"
)

var NmapVersion string

func ActiveScan(targets string, ports string) {

	// TODO: Use the system nmap if it is present
	// import "os/exec"
	// nmapPath, err = exec.LookPath("nmap")
	// if err != nil {
	// 	// extract nmap
	// }

	nmapDir, nmapPath := extractNmap()

	// Keeping this around for debugging
	// fmt.Printf(nmapPath)
	// fmt.Println("")

	scanner, err := createScanner(targets, ports, nmapPath)

	result, warnings, err := scanner.Run()
	if len(*warnings) > 0 {
		log.Printf("run finished with warnings: %s\n", *warnings) // Warnings are non-critical errors from nmap.
	}
	if err != nil {
		log.Fatalf("nmap scan failed: %v", err)
	}

	// TODO: Wait for SRUM-8 and implement the method to export this information to a csv file
	// Use the results to get the OS and ports open
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		matchedHosts := []string{}

		for _, match := range host.OS.Matches {
			if !slices.Contains(matchedHosts, host.Addresses[0].Addr) {
				for _, class := range match.Classes {
					switch class.OSFamily() {
					case osfamily.Linux:
						fmt.Printf("Discovered host running Linux: %q\n", host.Addresses[0])
						matchedHosts = append(matchedHosts, host.Addresses[0].Addr)
					case osfamily.Windows:
						fmt.Printf("Discovered host running Windows: %q\n", host.Addresses[0])
						matchedHosts = append(matchedHosts, host.Addresses[0].Addr)
					}
				}
			}
		}

		for _, port := range host.Ports {
			fmt.Printf("\tPort %d/%s %s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name, port.Service.Product)
			path := "output_passive.csv"
			header := []string{"Asset_IP", "Protocol", "Sourece_MAC", "Destination_MAC", "Ethernet_Type"}
			row := [][]string{{port.ID}}
			Storedata(path, header, row)
		}
	}

	fmt.Printf("Nmap done: %d hosts up scanned in %.2f seconds\n", len(result.Hosts), result.Stats.Finished.Elapsed)

	// Remove the dir containing nmap
	_ = os.RemoveAll(nmapDir)
}

func createScanner(targets string, ports string, nmapPath string) (*nmap.Scanner, error) {
	if ports == "" {
		return nmap.NewScanner(
			context.Background(),
			nmap.WithTargets(targets),
			nmap.WithOSDetection(), // TODO: Needs to run with sudo, need to make this a flag
			nmap.WithBinaryPath(nmapPath),
			nmap.WithMostCommonPorts(1000),
			nmap.WithServiceInfo(),
		)

	} else {
		return nmap.NewScanner(
			context.Background(),
			nmap.WithTargets(targets),
			nmap.WithOSDetection(), // TODO: Needs to run with sudo, need to make this a flag
			nmap.WithBinaryPath(nmapPath),
			nmap.WithPorts(ports),
			nmap.WithServiceInfo(),
		)
	}
}

func extractNmap() (string, string) {
	nmapBinaryName := "nmap"
	nmapExeName := nmapBinaryName + ".exe"
	nmapVersionedZip := "nmap-" + NmapVersion + "-win32.zip"
	extractedFolderName := nmapBinaryName + "-" + NmapVersion

	nmapWinZipFile, _ := assets.Assets.ReadFile(nmapVersionedZip)
	tmpDir, _ := os.MkdirTemp("", "discovr-embedded-bin-*")
	tmpPath := filepath.Join(tmpDir, nmapVersionedZip)
	_ = os.WriteFile(tmpPath, nmapWinZipFile, 0644)

	// Unzip and delete zip file
	unzip(tmpDir, tmpPath)
	_ = os.Remove(tmpDir + string(os.PathSeparator) + nmapVersionedZip)

	// If linux or macos, copy the nmap binary to the extracted folder
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		nmapLinuxBinary, _ := assets.Assets.ReadFile(nmapBinaryName)
		binPath := tmpDir + string(os.PathSeparator) + extractedFolderName + string(os.PathSeparator) + nmapBinaryName
		tmpPath = filepath.Join(binPath)
		_ = os.WriteFile(tmpPath, nmapLinuxBinary, 0755)
		return tmpDir, binPath
	}

	return tmpDir, tmpDir + string(os.PathSeparator) + extractedFolderName + string(os.PathSeparator) + nmapExeName
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
