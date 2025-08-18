package internal

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestWriteCSV(t *testing.T) {
	assets := []Asset{
		{
			IP:       "10.0.0.1",
			Hostname: "web-1",
			OS:       "Ubuntu",
			Role:     "server",
			Source:   "demo",
			Tags:     []string{"prod", "frontend"},
			LastSeen: time.Now().UTC(),
		},
	}

	out := t.TempDir() + "/out.csv"

	if err := WriteCSV(out, assets); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	s := string(b)

	// header + some cell values must appear
	for _, want := range []string{
		"ip,hostname,os,role", // header
		"10.0.0.1",
		"web-1",
		"Ubuntu",
		"server",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("csv missing %q\n--\n%s", want, s)
		}
	}
}
