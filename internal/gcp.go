package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func GcpScan(credFile string, projectFilterStr string) {
	ctx := context.Background()
	var resourceManagerClient *cloudresourcemanager.Service
	var computeService *compute.Service
	var err error

	// Create clients based on auth type
	if credFile != "" {
		resourceManagerClient, err = cloudresourcemanager.NewService(ctx, option.WithCredentialsFile(credFile))
		if err != nil {
			log.Fatalf("Failed to create resource manager client: %v", err)
		}

		computeService, err = compute.NewService(ctx, option.WithCredentialsFile(credFile))
		if err != nil {
			log.Fatalf("Failed to create compute service: %v", err)
		}
	} else {
		resourceManagerClient, err = cloudresourcemanager.NewService(ctx)
		if err != nil {
			log.Fatalf("Failed to create resource manager client: %v", err)
		}

		computeService, err = compute.NewService(ctx)
		if err != nil {
			log.Fatalf("Failed to create compute service: %v", err)
		}
	}

	

	// List All Projects
	projects, err := resourceManagerClient.Projects.List().Do()
	if err != nil {
		log.Fatalf("Failed to list projects: %v", err)
	}

	// Process Projects
	if len(projects.Projects) == 0 {
		fmt.Println("No projects found.")
		return
	}

	// Convert project filter into a list
	filteredProjects := strings.Split(projectFilterStr, ",")

	for _, project := range projects.Projects {
		if(contains(filteredProjects, project.Name) || projectFilterStr == ""){
			fmt.Printf("Checking instances for project: %s\n", project.Name)

			listInstanceNetworkInfo(computeService, project.ProjectId)
			// TODO: Enable on debug
			// if err != nil {
			//     log.Printf("Error retrieving instance network info for project %s: %v", project.ProjectId, err)
			// }
			fmt.Println()
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
		return fmt.Errorf("failed to list instances: %v", err)
	}

	// Process instances from all zones
	for zoneName, instancesScopedList := range instanceList.Items {

		// Skip zones with no instances
		if len(instancesScopedList.Instances) == 0 {
			continue
		}

		for _, instance := range instancesScopedList.Instances {
			fmt.Println()
			fmt.Printf("Project ID: %s\n", projectID)
			fmt.Printf("Hostname: %s\n", instance.Name)
			fmt.Printf("Zone: %s\n", strings.TrimPrefix(zoneName, "zones/"))

			for idx, networkInterface := range instance.NetworkInterfaces {
				fmt.Printf("\nNetwork Interface %d:\n", idx+1)
				fmt.Printf("Interface Name: %s\n", networkInterface.Name)

				// Internal IP
				if networkInterface.NetworkIP != "" {
					fmt.Printf("Internal IP: %s\n", networkInterface.NetworkIP)
				}

				// Collect all NatIP values (these are external Ips)
				var natIPs []string
				for _, accessConfig := range networkInterface.AccessConfigs {
					natIPs = append(natIPs, accessConfig.NatIP)
				}
				if len(natIPs) > 0 {
					natIPString := fmt.Sprintf("[%s]", strings.Join(natIPs, ","))
					fmt.Printf("External IPs: %s\n", natIPString)
				}

				// VPC
				if networkInterface.Network != "" {
					networkParts := strings.Split(networkInterface.Network, "/")
					if len(networkParts) > 0 {
						vpcID := networkParts[len(networkParts)-1]
						fmt.Printf("VPC: %s\n", vpcID)
					}
				}

				// Subnet
				if networkInterface.Subnetwork != "" {
					subnetParts := strings.Split(networkInterface.Subnetwork, "/")
					if len(subnetParts) > 0 {
						subnetID := subnetParts[len(subnetParts)-1]
						fmt.Printf("Subnet: %s\n", subnetID)
					}
				}
			}
		}
	}
	return nil
}
