package cmd

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/gopacket/pcap"
)

// helper functions
func GetOsPathPlaceholder() string {
	switch runtime.GOOS {
	case "windows":
		placeholder = "default = no result | output.csv | C:\\Output\\output.csv"
	default:
		placeholder = "default = no result | ./output.csv"
	}
	return placeholder
}

type Adapter struct {
	transport   string
	description string
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

// validation functions

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

func validateIP(input string) error {
	if input == "" {
		return nil
	}
	ip := net.ParseIP(input)
	if ip == nil {
		return fmt.Errorf("invalid IP address")
	}
	return nil
}

// Allowed formats: "80,445,8080" or "22-100"
func ValidatePorts(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", nil
	}

	input = strings.ReplaceAll(input, " ", "")

	re := regexp.MustCompile(`^(\d{1,5}(-\d{1,5})?)(,(\d{1,5}(-\d{1,5})?))*$`)

	if !re.MatchString(input) {
		return "", fmt.Errorf("invalid format, use ports like 80,445 or ranges like 22-100")
	}

	parts := strings.Split(input, ",")
	for _, p := range parts {
		if strings.Contains(p, "-") {
			bounds := strings.Split(p, "-")
			start := atoi(bounds[0])
			end := atoi(bounds[1])
			if start <= 0 || end > 65535 || start >= end {
				return "", fmt.Errorf("invalid port range: %s", p)
			}
		} else {
			port := atoi(p)
			if port <= 0 || port > 65535 {
				return "", fmt.Errorf("invalid port number: %s", p)
			}
		}
	}

	return input, nil
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
