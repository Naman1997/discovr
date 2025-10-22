package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/Naman1997/discovr/verbose"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

var Gcp_results []GcpScanResult

type GcpScanResult struct {
	ProjectId     string
	InstanceName  string
	Hostname      string
	OsType        string
	InterfaceName string
	InternalIP    string
	ExternalIPs   string
	VPC           string
	Subnet        string
}

func GcpScan(credFile string, projectFilterStr string) {
	ctx := context.Background()
	var resourceManagerClient *cloudresourcemanager.Service
	var computeService *compute.Service
	var err error

	// Create clients based on auth type
	if credFile != "" {
		resourceManagerClient, err = cloudresourcemanager.NewService(ctx, option.WithCredentialsFile(credFile))
		if err != nil {
			verbose.VerboseFatalfMsg("Failed to create resource manager client: %v", err)
		}

		computeService, err = compute.NewService(ctx, option.WithCredentialsFile(credFile))
		if err != nil {
			verbose.VerboseFatalfMsg("Failed to create compute service: %v", err)
		}
	} else {
		resourceManagerClient, err = cloudresourcemanager.NewService(ctx)
		if err != nil {
			verbose.VerboseFatalfMsg("Failed to create resource manager client: %v", err)
		}

		computeService, err = compute.NewService(ctx)
		if err != nil {
			verbose.VerboseFatalfMsg("Failed to create compute service: %v", err)
		}
	}

	// List All Projects
	projects, err := resourceManagerClient.Projects.List().Do()
	if err != nil {
		verbose.VerboseFatalfMsg("Failed to list projects: %v", err)
	}

	// Process Projects
	if len(projects.Projects) == 0 {
		verbose.VerbosePrintln("No projects found.")
		return
	}

	// Convert project filter into a list
	filteredProjects := strings.Split(projectFilterStr, ",")

	for _, project := range projects.Projects {
		if contains(filteredProjects, project.Name) || projectFilterStr == "" {
			verbose.VerbosePrintf("Checking instances for project: %s\n", project.Name)

			listInstanceNetworkInfo(computeService, project.ProjectId)
			// TODO: Enable on debug
			// if err != nil {
			//     log.Printf("Error retrieving instance network info for project %s: %v", project.ProjectId, err)
			// }
			verbose.VerbosePrintln()
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// listInstanceNetworkInfo retrieves network details for instances in a specific project
func listInstanceNetworkInfo(computeService *compute.Service, projectID string) error {

	// TODO: Figure out pagination
	instanceList, err := computeService.Instances.AggregatedList(projectID).Do()
	if err != nil {
		return verbose.VerboseErrorf("failed to list instances: %v", err)
	}

	// Process instances from all zones
	for _, instancesScopedList := range instanceList.Items {

		// Skip zones with no instances
		if len(instancesScopedList.Instances) == 0 {
			continue
		}

		for _, instance := range instancesScopedList.Instances {
			for _, networkInterface := range instance.NetworkInterfaces {

				// Print item
				verbose.VerbosePrintln()
				verbose.VerbosePrintf("Project ID: %s\n", projectID)
				verbose.VerbosePrintf("Instance ID: %d\n", instance.Id)
				verbose.VerbosePrintf("Instance Name: %s\n", instance.Name)
				verbose.VerbosePrintf("Hostname: %s\n", instance.Hostname)

				// Get OS details
				var osType string
				for _, disk := range instance.Disks {
					if disk.Boot {
						if disk.Licenses != nil && len(disk.Licenses) > 0 {
							// Split the URL and get the last part
							parts := strings.Split(disk.Licenses[0], "/")
							if len(parts) > 0 {
								osType = parts[len(parts)-1]
								verbose.VerbosePrintf("OS: %s\n", osType)
								break
							}
						}
					}
				}

				verbose.VerbosePrintf("Interface Name: %s\n", networkInterface.Name)

				// Internal IP
				if networkInterface.NetworkIP != "" {
					verbose.VerbosePrintf("Internal IP: %s\n", networkInterface.NetworkIP)
				}

				// Collect all NatIP values (these are external Ips)
				var natIPs []string
				var natIPString string
				for _, accessConfig := range networkInterface.AccessConfigs {
					natIPs = append(natIPs, accessConfig.NatIP)
				}
				if len(natIPs) > 0 {
					natIPString = fmt.Sprintf("[%s]", strings.Join(natIPs, ","))
					verbose.VerbosePrintf("External IPs: %s\n", natIPString)
				}

				// VPC
				var vpcID string
				if networkInterface.Network != "" {
					networkParts := strings.Split(networkInterface.Network, "/")
					if len(networkParts) > 0 {
						vpcID = networkParts[len(networkParts)-1]
						verbose.VerbosePrintf("VPC: %s\n", vpcID)
					}
				}

				// Subnet
				var subnetID string
				if networkInterface.Subnetwork != "" {
					subnetParts := strings.Split(networkInterface.Subnetwork, "/")
					if len(subnetParts) > 0 {
						subnetID = subnetParts[len(subnetParts)-1]
						verbose.VerbosePrintf("Subnet: %s\n", subnetID)
					}
				}

				// Collect results
				result := GcpScanResult{
					ProjectId:     projectID,
					InstanceName:  instance.Name,
					Hostname:      instance.Hostname,
					OsType:        osType,
					InterfaceName: networkInterface.Name,
					InternalIP:    networkInterface.NetworkIP,
					ExternalIPs:   natIPString,
					VPC:           vpcID,
					Subnet:        subnetID,
				}
				Gcp_results = append(Gcp_results, result)
			}
		}
	}
	return nil
}
