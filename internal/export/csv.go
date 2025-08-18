package export

import (
	"encoding/csv"
	"errors"
	"os"
	"strings"
	"time"
)

type Asset struct {
	IP, Hostname, OS, Role string
	Source, CloudProvider  string
	Account, Region        string
	Tags                   []string
	LastSeen               time.Time
}

func WriteCSV(path string, assets []Asset) error {
	if path == "" {
		return errors.New("csv path is required")
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// header
	if err := w.Write([]string{
		"ip", "hostname", "os", "role",
		"source", "cloud_provider", "account", "region",
		"tags", "last_seen",
	}); err != nil {
		return err
	}

	// rows
	for _, a := range assets {
		row := []string{
			a.IP, a.Hostname, a.OS, a.Role,
			a.Source, a.CloudProvider, a.Account, a.Region,
			strings.Join(a.Tags, " "),
			a.LastSeen.UTC().Format(time.RFC3339),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
