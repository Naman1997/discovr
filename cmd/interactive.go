package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Naman1997/discovr/internal"
	"github.com/Naman1997/discovr/verbose"
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
	projectfilter   string
	credPath        string
	verb            bool
)

func Runform(form *huh.Form) {
	if err := form.Run(); err != nil {
		log.Fatal(err)
	}
}

func RunTui() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pick a Scan Option").
				Options(huh.NewOptions("Active Scan", "Passive Scan", "Nmap Scan", "AWS Cloud Scan", "Azure Cloud Scan", "GCP Cloud Scan")...).
				Value(&scantype),
		),
	)
	Runform(form)

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
				huh.NewConfirm().
					Title("Enable Verbose Mode:").
					Value(&verbose.Verbose),
				huh.NewInput().
					Title("Enter the upload URL:").
					Value(&UploadUrl),
			),
		)
		Runform(form)
		VerboseEnabled()
		internal.DefaultScan(netInterface, tCIDR, icmpmode, concurrency, timeout, count)
		if !icmpmode {
			internal.ShowResults(internal.Defaultscan_results)
			internal.ExportCSV(exportpath, internal.Defaultscan_results)
			internal.UploadResults(UploadUrl, exportpath, internal.Defaultscan_results, "active_")
		} else {
			internal.ShowResults(internal.Icmpscan_results)
			internal.ExportCSV(exportpath, internal.Icmpscan_results)
			internal.UploadResults(UploadUrl, exportpath, internal.Defaultscan_results, "active_")
		}

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
				huh.NewConfirm().
					Title("Enable Verbose Mode:").
					Value(&verbose.Verbose),
				huh.NewInput().
					Title("Enter the upload URL:").
					Value(&UploadUrl),
			),
		)
		Runform(form)
		VerboseEnabled()
		if durationStr == "" {
			duration = 20
		} else {
			duration, _ = strconv.Atoi(durationStr)
		}
		internal.PassiveScan(selectinterface, duration)
		internal.ShowResults(internal.Passive_results)
		internal.ExportCSV(exportpath, internal.Passive_results)
		internal.UploadResults(UploadUrl, exportpath, internal.Passive_results, "passive_")

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
				huh.NewConfirm().
					Title("Enable Verbose Mode:").
					Value(&verbose.Verbose),
				huh.NewInput().
					Title("Enter the upload URL:").
					Value(&UploadUrl),
			),
		)
		Runform(form)
		VerboseEnabled()
		if ip == "" {
			ip = "127.0.0.1"
		}
		internal.NmapScan(ip, ports, osdet)
		internal.ShowResults(internal.Active_results)
		internal.ExportCSV(exportpath, internal.Active_results)
		internal.UploadResults(UploadUrl, exportpath, internal.Active_results, "nmap_")

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
				huh.NewConfirm().
					Title("Enable Verbose Mode:").
					Value(&verbose.Verbose),
				huh.NewInput().
					Title("Enter the upload URL:").
					Value(&UploadUrl),
			),
		)
		Runform(form)
		VerboseEnabled()
		if subID == "" {
			subID = "default"
		}
		internal.Azurescan(subID)
		internal.ShowResults(internal.Azure_results)
		internal.ExportCSV(exportpath, internal.Azure_results)
		internal.UploadResults(UploadUrl, exportpath, internal.Azure_results, "azure_")

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
				huh.NewConfirm().
					Title("Enable Verbose Mode:").
					Value(&verbose.Verbose),
				huh.NewInput().
					Title("Enter the upload URL:").
					Value(&UploadUrl),
			),
		)
		Runform(form)
		VerboseEnabled()
		internal.AwsScan(regionselect, []string{}, []string{}, "")
		internal.ShowResults(internal.Aws_results)
		internal.ExportCSV(exportpath, internal.Aws_results)
		internal.UploadResults(UploadUrl, exportpath, internal.Aws_results, "aws_")

	case "GCP Cloud Scan":
		var projectOptions []huh.Option[string]
		pathplaceholder := GetOsPathPlaceholder()
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter Custom Credential Path:").
					Placeholder("default = default credentials").
					Value(&credPath),
			),
		)
		Runform(form)
		projects, err := FetchGCPProjects(credPath)
		if err != nil {
			fmt.Printf("%v", err)
		}
		for _, r := range projects {
			projectOptions = append(projectOptions, huh.Option[string]{
				Key:   r,
				Value: r,
			})
		}
		form_1 := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a GCP project").
					Options(projectOptions...).
					Value(&projectfilter),
				huh.NewInput().
					Title("Enter an export path:").
					Placeholder(pathplaceholder).
					Value(&exportpath),
				huh.NewConfirm().
					Title("Enable Verbose Mode:").
					Value(&verbose.Verbose),
				huh.NewInput().
					Title("Enter the upload URL:").
					Value(&UploadUrl),
			),
		)
		Runform(form_1)
		VerboseEnabled()
		internal.GcpScan(credPath, projectfilter)
		internal.ShowResults(internal.Gcp_results)
		internal.ExportCSV(exportpath, internal.Gcp_results)
		internal.UploadResults(UploadUrl, exportpath, internal.Gcp_results, "gcp_")
	}
}
