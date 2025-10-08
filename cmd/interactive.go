package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Naman1997/discovr/internal"
	"github.com/charmbracelet/huh"
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
	netInterface    string
	tCIDR           string
	subID           string
	icmpmode        bool
)

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
				Value(&scantype),
		),
	)
	errhandle(form)

	switch scantype {
	case "Active Scan":
		var options []huh.Option[string]
		pathplaceholder := GetOsPathPlaceholder()
		options, err := GetInterfaceOptions()
		if err != nil {
			panic(err)
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select an option").
					Options(options...).
					Value(&netInterface),
				huh.NewInput().
					Title("Enter a CIDR range:").
					Value(&tCIDR),
				huh.NewConfirm().
					Title("ICMP Scan type:").
					Value(&icmpmode),
				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&exportpath),
			),
		)
		errhandle(form)
		internal.DefaultScan(netInterface, tCIDR, icmpmode, concurrency, timeout, count)
		if !icmpmode {
			internal.ShowResults(internal.Defaultscan_results)
		} else {
			internal.ShowResults(internal.Icmpscan_results)
		}
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
		internal.ShowResults(internal.Passive_results)
		internal.PassiveExport(exportpath)

	case "Nmap Scan":
		pathplaceholder := GetOsPathPlaceholder()
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter Target IP:").
					Placeholder("default = 127.0.0.1").
					Value(&ip).
					Validate(validateIP),
				huh.NewInput().
					Title("Enter Ports").
					Placeholder("default = Top 1000 ports | specify ports eg: 80,445").
					Value(&ports).
					Validate(func(input string) error {
						normalized, err := ValidatePorts(input)
						if err != nil {
							return err
						}
						ports = normalized
						return nil
					}),
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
		internal.ShowResults(internal.Active_results)
		internal.NmapExport(exportpath)

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
		internal.ShowResults(internal.Azure_results)
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
