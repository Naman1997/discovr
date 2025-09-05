package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// TODO: add configPath string, credentialPath string for auth options
func AwsScan(regionFilter string) {
	// Load the AWS config (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an EC2 client to list all regions
	svc := ec2.NewFromConfig(cfg)
	result, err := svc.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		log.Fatalf("failed to describe regions, %v", err)
	}

	// Loop through each region and describe instance in each one
	if(regionFilter == ""){
		for _, region := range result.Regions {
			regionName := aws.ToString(region.RegionName)
			fmt.Printf("Region: %s\n", regionName)
			ProcessInstancesForRegion(cfg, regionName)
		}
	} else {
		ProcessInstancesForRegion(cfg, regionFilter)
	}
}

func ProcessInstancesForRegion(cfg aws.Config, regionName string) {
	cfg.Region = regionName
	regionSvc := ec2.NewFromConfig(cfg)
	paginator := ec2.NewDescribeInstancesPaginator(regionSvc, &ec2.DescribeInstancesInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("failed to describe instances in region %s, %v", regionName, err)
		}

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				instanceID := aws.ToString(instance.InstanceId)
				fmt.Printf("  Instance ID: %s\n", instanceID)

				pageSize := int32(10)
				netPaginator := ec2.NewDescribeNetworkInterfacesPaginator(regionSvc, &ec2.DescribeNetworkInterfacesInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("attachment.instance-id"),
							Values: []string{instanceID},
						},
					},
				}, func(o *ec2.DescribeNetworkInterfacesPaginatorOptions) {
					o.Limit = pageSize
				})
				for netPaginator.HasMorePages() {
					netOutput, err := netPaginator.NextPage(context.TODO())
					if err != nil {
						log.Fatalf("failed to describe network interfaces in region %s, %v", regionName, err)
					}

					// Loop through each network interface and get the network details
					for _, netInterface := range netOutput.NetworkInterfaces {

						publicIP := ""
						if netInterface.Association != nil && netInterface.Association.PublicIp != nil {
							publicIP = aws.ToString(netInterface.Association.PublicIp)
						}

						macAddress := aws.ToString(netInterface.MacAddress)
						vpcID := aws.ToString(netInterface.VpcId)

						subnetID := ""
						if netInterface.SubnetId != nil {
							subnetID = aws.ToString(netInterface.SubnetId)
						}

						privateIPs := make([]string, len(netInterface.PrivateIpAddresses))
						for i, ip := range netInterface.PrivateIpAddresses {
							privateIPs[i] = aws.ToString(ip.PrivateIpAddress)
						}

						hostname := ""
						if netInterface.PrivateDnsName != nil {
							hostname = aws.ToString(netInterface.PrivateDnsName)
						}

						fmt.Printf("    Public IP: %s\n", publicIP)
						fmt.Printf("    MAC Address: %s\n", macAddress)
						fmt.Printf("    VPC ID: %s\n", vpcID)
						fmt.Printf("    Subnet ID: %s\n", subnetID)
						fmt.Printf("    Private IPs: %v\n", privateIPs)
						fmt.Printf("    Hostname: %s\n", hostname)
					}
				}
			}
		}
	}
}