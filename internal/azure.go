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

var subID string
var err error

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

			// NIC info
			for _, nicRef := range vm.Properties.NetworkProfile.NetworkInterfaces {
				nicID, _ := arm.ParseResourceID(*nicRef.ID)
				nic, _ := nicClient.Get(ctx, nicID.ResourceGroupName, nicID.Name, nil)

				result.NIC = nicID.Name
				if nic.Properties.MacAddress != nil {
					result.MAC = *nic.Properties.MacAddress
				}

				// IP configs
				for _, ipConf := range nic.Properties.IPConfigurations {
					result.PrivateIP = *ipConf.Properties.PrivateIPAddress

					if ipConf.Properties.Subnet != nil {
						subnetID, _ := arm.ParseResourceID(*ipConf.Properties.Subnet.ID)
						result.Subnet = subnetID.Name
						result.Vnet = fmt.Sprint(subnetID.Parent.Parent.Name)
					}
					if ipConf.Properties.PublicIPAddress != nil {
						pipID, _ := arm.ParseResourceID(*ipConf.Properties.PublicIPAddress.ID)
						pip, _ := pipClient.Get(ctx, pipID.ResourceGroupName, pipID.Name, nil)
						result.PublicIP = *pip.Properties.IPAddress
					}
				}
			}

			fmt.Printf("\nName: %s\n", result.Name)
			fmt.Printf("ID: %s\n", result.UniqueID)
			fmt.Printf("Location: %s\n", result.Location)
			fmt.Printf("Resource Group: %s\n", result.ResourceGroup)
			fmt.Printf("NIC: %s\n", result.NIC)
			fmt.Printf("MAC Address: %s\n", result.MAC)
			fmt.Printf("Subnet: %s\n", result.Subnet)
			fmt.Printf("Vnet: %s\n", result.Vnet)
			fmt.Printf("Private IP: %s\n", result.PrivateIP)
			fmt.Printf("Public IP: %s\n", result.PublicIP)

			azure_results = append(azure_results, result)
		}
	}
}
