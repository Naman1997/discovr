package internal

import (
	"archive/zip"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/Naman1997/discovr/assets"
	"github.com/Naman1997/discovr/verbose"
	"github.com/Ullaakut/nmap/v3"
	osfamily "github.com/Ullaakut/nmap/v3/pkg/osfamilies"
)

var NmapVersion string = "7.92"

var Active_results []ScanResultActive

type ScanResultActive struct {
	Port     string
	Protocol string
	State    string
	Service  string
	Product  string
}

func NmapScan(targets string, ports string, osDetection bool) {

	//TODO: Scanning default scan if not using nmap

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

	// Configure the nmap scanner
	scanner, err := createScanner(targets, nmapPath)
	if ports == "" {
		scanner.AddOptions(nmap.WithMostCommonPorts(1000))
	} else {
		scanner.AddOptions(nmap.WithPorts(ports))
	}

	// Only enable OS detection if user is running with elevated privs
	if osDetection {
		isAdmin := false
		if runtime.GOOS == "windows" {
			isAdmin = isWindowsAdmin()
		} else {
			// For Unix-like systems (Linux, macOS, etc.)
			if os.Geteuid() == 0 {
				isAdmin = true
			}
		}

		if isAdmin {
			scanner.AddOptions(nmap.WithOSDetection())
			scanner.AddOptions(nmap.WithPrivileged())
		} else {
			log.Fatalf("Scan Failed: OS scan requires elevated privileges!")
		}
	}

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

		// Log host OS if OS detection is enabled
		if len(host.OS.Matches) > 0 {
			matchedHosts := []string{}
			for _, match := range host.OS.Matches {
				if !slices.Contains(matchedHosts, host.Addresses[0].Addr) {
					for _, class := range match.Classes {
						switch class.OSFamily() {
						case osfamily.Linux:
							verbose.VerbosePrintf("Discovered host running Linux: %q\n", host.Addresses[0])
							matchedHosts = append(matchedHosts, host.Addresses[0].Addr)
						case osfamily.Windows:
							verbose.VerbosePrintf("Discovered host running Windows: %q\n", host.Addresses[0])
							matchedHosts = append(matchedHosts, host.Addresses[0].Addr)
						}
					}
				}
			}
		} else {
			verbose.VerbosePrintf("Discovered host: %q\n", host.Addresses[0])
		}

		for _, port := range host.Ports {
			verbose.VerbosePrintf("\tPort %d/%s %s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name, port.Service.Product)

			// export SCRUM-94
			result := ScanResultActive{
				Port:     strconv.Itoa(int(port.ID)),
				Protocol: port.Protocol,
				State:    port.State.State,
				Service:  port.Service.Name,
				Product:  port.Service.Product,
			}
			Active_results = append(Active_results, result)
		}
	}

	verbose.VerbosePrintf("Nmap done: %d hosts up scanned in %.2f seconds\n", len(result.Hosts), result.Stats.Finished.Elapsed)

	// Remove the dir containing nmap
	defer os.RemoveAll(nmapDir)
}

func createScanner(targets string, nmapPath string) (*nmap.Scanner, error) {
	return nmap.NewScanner(
		context.Background(),
		nmap.WithTargets(targets),
		nmap.WithBinaryPath(nmapPath),
		nmap.WithServiceInfo(),
		nmap.WithUnprivileged(),
	)
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
			verbose.Printf("invalid file path")
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

// Source: https://gist.github.com/jerblack/d0eb182cc5a1c1d92d92a4c4fcc416c6
func isWindowsAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}
