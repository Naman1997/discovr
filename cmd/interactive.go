package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Naman1997/discovr/internal"
	"github.com/charmbracelet/huh"
)

var (
	scantype        string
	localtype       string
	SelectInterface string
)

type Adapter struct {
	MAC       string
	Transport string
}

func GetAdapters() ([]Adapter, error) {
	// runs getmac /fo csv, replace Tcpip with NPF and appends it to Adapter struct
	cmd := exec.Command("cmd", "/C", "getmac /fo csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running getmac: %w", err)
	}

	reader := csv.NewReader(strings.NewReader(string(output)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing CSV: %w", err)
	}

	var adapters []Adapter
	for _, row := range records[1:] {
		if len(row) < 2 {
			continue
		}
		mac := row[0]
		transport := row[1]

		transportNFP := strings.Replace(transport, `\Device\Tcpip_`, `\Device\NPF_`, 1)

		adapters = append(adapters, Adapter{
			MAC:       mac,
			Transport: transportNFP,
		})
	}
	return adapters, nil
}

func errhandle(form *huh.Form) {
	if err := form.Run(); err != nil {
		log.Fatal(err)
	}
}

func tui() string {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose an option").
				Description("Select Scan Type").
				Options(
					huh.NewOption("Local Scan", "local"),
					huh.NewOption("Azure Cloud Scan", "azure"),
					huh.NewOption("AWS Cloud Scan", "aws"),
				).
				Value(&scantype),
		),
	)
	errhandle(form)
	return scantype
}

func RunTui() {
	// runs form with options for local, azure, AWS scan
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose an option").
				Description("Select Scan Type").
				Options(
					huh.NewOption("Local Scan", "local"),
					huh.NewOption("Azure Cloud Scan", "azure"),
					huh.NewOption("AWS Cloud Scan", "aws"),
				).
				Value(&scantype),
		),
	)
	errhandle(form)

	switch {
	case scantype == "local":
		// runs form for passive or active option
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose an option").
					Description("Select Local Scan Type").
					Options(
						huh.NewOption("Passive Scan", "passive"),
						huh.NewOption("Active Scan", "active"),
					).
					Value(&localtype),
			),
		)
		errhandle(form)

		switch {
		case localtype == "passive":
			// runs GetAdapters() function if OS is Windows and appends it into var options
			var options []huh.Option[string]
			if runtime.GOOS == "windows" {
				adapters, err := GetAdapters()
				if err != nil {
					fmt.Println("Error", err)
					return
				}
				for _, a := range adapters {
					MacTransport := fmt.Sprintf("%v,%v", a.MAC, a.Transport)
					options = append(options, huh.Option[string]{
						Key:   MacTransport,
						Value: a.Transport,
					})
				}
			} else {
				return // put linux or macos code here
			}

			// form displays options for selection
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Select an option").
						Options(options...).
						Value(&SelectInterface),
				),
			)
			// add duration input
			// add export input
			errhandle(form)

			fmt.Println("Running Local active scan on Interface: ", SelectInterface)
			internal.PassiveScan(SelectInterface, 20)

		case localtype == "active":
			fmt.Print("Active Scan Called")
			// target local if empty huh.text
			// ports(top 1000, all(-p-), choose via input) huh.input/text
			// os(true, false) huh.confirm
			// Export(path) huh.text
			// internal.ActiveScan
		}

	case scantype == "azure":
		fmt.Print("Azure Called")
	case scantype == "aws":
		fmt.Print("AWS Called")
	}
}

// if OS windows run getmac command within this program and printout as options
// use the selected one to do passive scans
// for active scans require ip address for scans. Do input filtering so that anything besides that is not put
// Allow subnet range in another form
