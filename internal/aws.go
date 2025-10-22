package internal

import (
	"context"
	"os"
	"strings"

	"github.com/Naman1997/discovr/verbose"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var Aws_results []AwsScanResult

type AwsScanResult struct {
	InstanceId string
	PublicIp   string
	PrivateIPs string
	MacAddress string
	VpcId      string
	SubnetId   string
	Hostname   string
	Region     string
}

func AwsScan(regionFilter string, customConfigs []string, customCredentials []string, customProfile string) {

	// Set custom profile if provided
	if customProfile != "" {
		os.Setenv("AWS_PROFILE", customProfile)
	}

	var cfg aws.Config
	var err error

	// Load custom config options
	if len(customConfigs) != 0 && len(customCredentials) != 0 {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithSharedCredentialsFiles(
				customCredentials,
			),
			config.WithSharedConfigFiles(
				customConfigs,
			),
		)
	} else if len(customConfigs) != 0 {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigFiles(
				customConfigs,
			),
		)
	} else if len(customCredentials) != 0 {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithSharedCredentialsFiles(
				customCredentials,
			),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO())
		if err != nil {
			verbose.VerboseFatalfMsg("unable to load SDK config, %v", err)
		}
	}

	// Create an EC2 client to list all regions
	svc := ec2.NewFromConfig(cfg)
	result, err := svc.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		verbose.VerboseFatalfMsg("failed to describe regions, %v", err)
	}

	// Loop through each region and describe instance in each one
	if regionFilter == "" {
		for _, region := range result.Regions {
			regionName := aws.ToString(region.RegionName)
			verbose.VerbosePrintf("Scanning region: %s\n", regionName)
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
			verbose.VerboseFatalfMsg("failed to describe instances in region %s, %v", regionName, err)
		}

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				instanceID := aws.ToString(instance.InstanceId)
				verbose.VerbosePrintf("  Instance ID: %s\n", instanceID)

				pageSize := int32(50)
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
						verbose.VerboseFatalfMsg("failed to describe network interfaces in region %s, %v", regionName, err)
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

						verbose.VerbosePrintf("	Instance Id: %s\n", instanceID)
						verbose.VerbosePrintf("    Public IP: %s\n", publicIP)
						verbose.VerbosePrintf("    MAC Address: %s\n", macAddress)
						verbose.VerbosePrintf("    VPC ID: %s\n", vpcID)
						verbose.VerbosePrintf("    Subnet ID: %s\n", subnetID)
						verbose.VerbosePrintf("    Private IPs: %v\n", privateIPs)
						verbose.VerbosePrintf("    Hostname: %s\n", hostname)
						verbose.VerbosePrintf("    Region: %s\n", regionName)

						result := AwsScanResult{
							InstanceId: instanceID,
							PublicIp:   publicIP,
							PrivateIPs: strings.Join(privateIPs, " "),
							MacAddress: macAddress,
							VpcId:      vpcID,
							SubnetId:   subnetID,
							Hostname:   hostname,
							Region:     regionName,
						}
						Aws_results = append(Aws_results, result)
					}
				}
			}
		}
	}
}
