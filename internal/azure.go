package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

var azure_results []AzureVMResult

type AzureVMResult struct {
	Name          string
	UniqueID      string
	Location      string
	ResourceGroup string
	NIC           string
	MAC           string
	Subnet        string
	Vnet          string
	PrivateIP     string
	PublicIP      string
}

type AzureProfile struct {
	Subscriptions []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		IsDefault bool   `json:"isDefault"`
	} `json:"subscriptions"`
}

// reads azureProfile.json and outputs default SubscriptionID
func GetDefaultSubscription() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	profilePath := filepath.Join(home, ".azure", "azureProfile.json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return "", err
	}

	dataStr := strings.TrimSpace(strings.TrimPrefix(string(data), "\uFEFF"))

	var profile AzureProfile
	if err := json.Unmarshal([]byte(dataStr), &profile); err != nil {
		return "", err
	}

	for _, sub := range profile.Subscriptions {
		if sub.IsDefault {
			return sub.ID, nil
		}
	}
	return "", errors.New("no default subscription found")
}

func Azurescan(subIdInput string) {
	var subID string
	var err error

	if subIdInput == "default" {
		subID, err = GetDefaultSubscription()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		subID = strings.TrimSpace(subIdInput)
	}

	fmt.Println("----------------------------------------")

	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	vmClient, err := armcompute.NewVirtualMachinesClient(subID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	// raises errors if Invalid SubID or no VMs in under Sub
	testPager := vmClient.NewListAllPager(nil)
	if testPager.More() {
		_, err := testPager.NextPage(ctx)
		if err != nil {
			log.Fatalf("Subscription ID %v does not Exist: %v", subID, err)
		}
	} else {
		log.Fatalf("Subscription %s is invalid or contains no VMs", subID)
	}

	nicClient, err := armnetwork.NewInterfacesClient(subID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	subnetClient, err := armnetwork.NewSubnetsClient(subID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	pipClient, err := armnetwork.NewPublicIPAddressesClient(subID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	// VM Client for creating pager and initial info
	pager := vmClient.NewListAllPager(nil)
	for pager.More() {
		page, _ := pager.NextPage(ctx)
		for _, vm := range page.Value {
			vmID, _ := arm.ParseResourceID(*vm.ID)

			result := AzureVMResult{
				Name:          *vm.Name,
				UniqueID:      *vm.Properties.VMID,
				Location:      *vm.Location,
				ResourceGroup: vmID.ResourceGroupName,
			}

			var (
				macs  []string
				nics  []string
				ips   []string
				ip    string
				pips  []string
				snets []string
				vnets []string
			)

			// NICs
			for _, nicRef := range vm.Properties.NetworkProfile.NetworkInterfaces {
				nicID, _ := arm.ParseResourceID(*nicRef.ID)
				nic, _ := nicClient.Get(ctx, nicID.ResourceGroupName, nicID.Name, nil)

				if nic.Name != nil {
					nics = append(nics, nicID.Name)
				}
				if nic.Properties.MacAddress != nil {
					macs = append(macs, *nic.Properties.MacAddress)
				}

				// IP configs
				for _, ipConf := range nic.Properties.IPConfigurations {
					if ipConf.Properties.PrivateIPAddress != nil {
						// ips = append(ips, *ipConf.Properties.PrivateIPAddress)
						ip = *ipConf.Properties.PrivateIPAddress
					}

					// Subnets and VNets
					if ipConf.Properties.Subnet != nil {
						subnetID, _ := arm.ParseResourceID(*ipConf.Properties.Subnet.ID)
						if subnetID.Parent.Parent.Name != "" {
							vnets = append(vnets, subnetID.Parent.Name)
							snets = append(snets, subnetID.Name)
						}
						subnetResp, err := subnetClient.Get(ctx, subnetID.ResourceGroupName, subnetID.Parent.Name, subnetID.Name, nil)
						cidr := "unknown" // default
						if err == nil && subnetResp.Properties != nil {
							if len(subnetResp.Properties.AddressPrefixes) > 0 && subnetResp.Properties.AddressPrefixes[0] != nil {
								cidr = *subnetResp.Properties.AddressPrefixes[0]
							} else if subnetResp.Properties.AddressPrefix != nil {
								cidr = *subnetResp.Properties.AddressPrefix
							}
						}

						// separate mask from cidr
						mask := ""
						if cidr != "unknown" || cidr != "" {
							parts := strings.Split(cidr, "/")
							if len(parts) == 2 {
								mask = "/" + parts[1]
							}
						}
						// joining mask with IP and appending to IPs slice
						ips = append(ips, fmt.Sprintf("%s%s", ip, mask))
					}
					// Public IP
					if ipConf.Properties.PublicIPAddress != nil {
						pipID, _ := arm.ParseResourceID(*ipConf.Properties.PublicIPAddress.ID)
						pip, _ := pipClient.Get(ctx, pipID.ResourceGroupName, pipID.Name, nil)
						if err == nil && pip.Properties.IPAddress != nil {
							pips = append(pips, *pip.Properties.IPAddress)
						}
					}
				}
			}

			result.NIC = strings.Join(nics, ", ")
			result.MAC = strings.Join(macs, ", ")
			result.PrivateIP = strings.Join(ips, ", ")
			result.Subnet = strings.Join(snets, ", ")
			result.Vnet = strings.Join(vnets, ", ")
			result.PublicIP = strings.Join(pips, ", ")

			fmt.Printf("\nName: %s\n", result.Name)
			fmt.Printf("ID: %s\n", result.UniqueID)
			fmt.Printf("Location: %s\n", result.Location)
			fmt.Printf("Resource Group: %s\n", result.ResourceGroup)
			fmt.Printf("Private IP: %s\n", result.PrivateIP)
			fmt.Printf("Public IP: %s\n", result.PublicIP)
			fmt.Printf("MAC Address: %s\n", result.MAC)
			fmt.Printf("Vnet: %s\n", result.Vnet)
			fmt.Printf("NIC: %s\n", result.NIC)
			fmt.Printf("Subnet: %s\n", result.Subnet)

			azure_results = append(azure_results, result)
		}
	}
}
