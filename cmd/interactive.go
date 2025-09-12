package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"

	"github.com/Naman1997/discovr/internal"
	"github.com/charmbracelet/huh"
	"github.com/google/gopacket/pcap"
)

var (
	// TODO: Fix var names
	Scantype        string
	SelectInterface string
	Exportpath      string
	duration        int
	durationStr     string
	ip              string
	ports           string
	osdet           bool
	placeholder     string
	regionselect    string
)

const awsIPRangesURL = "https://ip-ranges.amazonaws.com/ip-ranges.json"

type Adapter struct {
	transport   string
	description string
}

type awsIPRanges struct {
	Prefixes []struct {
		Region string `json:"region"`
	} `json:"prefixes"`
}

func CheckOsPathPlaceholder() string {
	switch runtime.GOOS {
	case "windows":
		placeholder = "default = no result | output.csv | C:\\Output\\Nmap.csv"
	case "linux":
		placeholder = "default = no result | ./Nmap.csv"
	default:
		placeholder = "default = no result"
	}
	return placeholder
}

func GetAdapters() ([]Adapter, error) {
	devs, err := pcap.FindAllDevs()
	if err != nil {
		return nil, fmt.Errorf("pcap device enumeration failed: %w", err)
	}
	var adapters []Adapter
	for _, dev := range devs {
		if len(dev.Addresses) == 0 {
			continue
		}
		adapters = append(adapters, Adapter{
			transport:   dev.Name,
			description: dev.Description,
		})
	}

	return adapters, nil
}

func FetchAWSRegion() ([]string, error) {
	resp, err := http.Get(awsIPRangesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AWS ip-ranges.json: %w", err)
	}
	defer resp.Body.Close()

	var data awsIPRanges
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	regionSet := make(map[string]struct{})
	for _, p := range data.Prefixes {
		if p.Region != "" {
			regionSet[p.Region] = struct{}{}
		}
	}

	var regions []string
	for region := range regionSet {
		regions = append(regions, region)
	}

	return regions, nil
}

func errhandle(form *huh.Form) {
	if err := form.Run(); err != nil {
		log.Fatal(err)
	}
}

func RunTui() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pick a Scan Option").
				Options(huh.NewOptions("Active Scan", "Passive Scan", "Nmap Scan", "AWS Cloud Scan", "Azure Cloud Scan")...).
				Value(&Scantype),
		),
	)
	errhandle(form)

	switch {
	case Scantype == "Active Scan":
		pathplaceholder := CheckOsPathPlaceholder()
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&Exportpath),
			),
		)
		errhandle(form)
		internal.DefaultScan()
		header := []string{"Interface", "Dest_IP", "Dest_Mac"} //TODO: remove header for next ticket
		internal.ActiveExport(Exportpath, header, false)

	case Scantype == "Passive Scan":
		var options []huh.Option[string]
		pathplaceholder := CheckOsPathPlaceholder()
		adapters, err := GetAdapters()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if len(adapters) == 0 {
			fmt.Println("No active adapters found.")
			return
		}
		for _, a := range adapters {
			description_transport := fmt.Sprintf("%v, %v", a.description, a.transport)
			options = append(options, huh.Option[string]{
				Key:   description_transport,
				Value: a.transport,
			})
		}
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select an option").
					Options(options...).
					Value(&SelectInterface),
				huh.NewInput().
					Title("Enter a duration (sec)").
					Placeholder("default = 20").
					Value(&durationStr).
					Validate(func(s string) error {
						if s == "" {
							return nil
						}
						_, err := strconv.Atoi(s)
						return err
					}),
				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&Exportpath),
			),
		)
		errhandle(form)
		if durationStr == "" {
			duration = 20
		} else {
			duration, _ = strconv.Atoi(durationStr)
		}
		internal.PassiveScan(SelectInterface, duration)
		header := []string{"Source_IP", "Protocol", "Source_MAC", "Destination_Mac", "Ethernet_Type"} //TODO: remove header
		internal.PassiveExport(PathPassive, header)

	case Scantype == "Nmap Scan":
		pathplaceholder := CheckOsPathPlaceholder()
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter Target IP:").
					Placeholder("default = 127.0.0.1").
					Value(&ip),
				huh.NewInput().
					Title("Enter Ports").
					Placeholder("default = Top 1000 ports | specify ports eg 80,445").
					Value(&ports), // TODO: Remove spaces between numbers and make sure , after numbers and no space OR Leave it upto user?
				huh.NewConfirm().
					Title("OS detection:").
					Value(&osdet),

				// TODO: add flag for -Pn scan here??

				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&Exportpath),
			),
		)
		errhandle(form)
		if ip == "" {
			ip = "127.0.0.1"
		}
		internal.NmapScan(ip, ports, osdet)
		header := []string{"Port", "Protocol", "State", "Service", "Product"}
		internal.ActiveExport(Exportpath, header, true)

	case Scantype == "Azure Cloud Scan":
		fmt.Print("Azure Called")

	case Scantype == "AWS Cloud Scan":
		// Region, Config, Credential, Profile, export path, TODO: Do AWS scan internal.AwsScan(Region, Config, Credential, Profile)
		var regionOptions []huh.Option[string]
		region, err := FetchAWSRegion()
		if err != nil {
			fmt.Printf("%v", err)
		}
		for _, r := range region {
			regionOptions = append(regionOptions, huh.Option[string]{
				Key:   r,
				Value: r,
			})
		}
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select an AWS region").
					Options(regionOptions...).
					Value(&regionselect),
			),
		)
		errhandle(form)
	}
}
