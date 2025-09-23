package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strconv"

	"github.com/Naman1997/discovr/internal"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/charmbracelet/huh"
	"github.com/google/gopacket/pcap"
)

var (
	scantype        string
	selectinterface string
	exportpath      string
	duration        int
	durationStr     string
	ip              string
	ports           string
	osdet           bool
	placeholder     string
	regionselect    string
	subID           string
)

type Adapter struct {
	transport   string
	description string
}

func GetOsPathPlaceholder() string {
	switch runtime.GOOS {
	case "windows":
		placeholder = "default = no result | output.csv | C:\\Output\\output.csv"
	default:
		placeholder = "default = no result | ./output.csv"
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
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	svc := ec2.NewFromConfig(cfg)
	result, err := svc.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}
	var regions []string
	for _, region := range result.Regions {
		regions = append(regions, *region.RegionName)
	}
	return regions, nil
}

func errhandle(form *huh.Form) {
	if err := form.Run(); err != nil {
		log.Fatal(err)
	}
}

var guidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func validateSubscriptionID(input string) error {
	if input == "" {
		return nil
	}
	if !guidRegex.MatchString(input) {
		return fmt.Errorf("invalid subscription ID format")
	}
	return nil
}

func RunTui() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pick a Scan Option").
				Options(huh.NewOptions("Active Scan", "Passive Scan", "Nmap Scan", "AWS Cloud Scan", "Azure Cloud Scan")...).
				Value(&scantype),
		),
	)
	errhandle(form)

	switch scantype {
	case "Active Scan":
		pathplaceholder := GetOsPathPlaceholder()
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&exportpath),
			),
		)
		errhandle(form)
		internal.DefaultScan()
		internal.ActiveExport(exportpath, false)

	case "Passive Scan":
		var options []huh.Option[string]
		pathplaceholder := GetOsPathPlaceholder()
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
					Value(&selectinterface),
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
					Value(&exportpath),
			),
		)
		errhandle(form)
		if durationStr == "" {
			duration = 20
		} else {
			duration, _ = strconv.Atoi(durationStr)
		}
		internal.PassiveScan(selectinterface, duration)
		internal.PassiveExport(exportpath)

	case "Nmap Scan":
		pathplaceholder := GetOsPathPlaceholder()
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
				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&exportpath),
			),
		)
		errhandle(form)
		if ip == "" {
			ip = "127.0.0.1"
		}
		internal.NmapScan(ip, ports, osdet)
		internal.ActiveExport(exportpath, true)

	case "Azure Cloud Scan":
		pathplaceholder := GetOsPathPlaceholder()
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter Subscription ID:").
					Description("Use AzureCLI command 'az account list -o table' to list subscription ID\nor leave empty for default subscriptionID\n").
					Placeholder("00000000-0000-0000-0000-000000000000 | default = default subscriptionID").
					Value(&subID).
					Validate(validateSubscriptionID),
				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&exportpath),
			),
		)
		errhandle(form)
		if subID == "" {
			subID = "default"
		}
		internal.Azurescan(subID)
		internal.AzureExport(exportpath)

	case "AWS Cloud Scan":
		var regionOptions []huh.Option[string]
		pathplaceholder := GetOsPathPlaceholder()
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
				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&exportpath),
			),
		)
		errhandle(form)
		internal.AwsScan(regionselect, []string{}, []string{}, "")
		internal.AwsExport(exportpath)
	}
}
