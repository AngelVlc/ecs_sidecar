package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func main() {
	ctx := context.TODO()

	clusterName := os.Getenv("CLUSTER_NAME")
	log.Printf("Cluster name: '%v'\n", clusterName)

	subdomain := os.Getenv("SUBDOMAIN")
	log.Printf("Subdomain: '%v'\n", subdomain)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("error loading the default config: %v", err)
	}

	ecsApi := InitEcsApi(cfg)

	eni, err := getTaskEni(ctx, ecsApi, clusterName)
	if err != nil {
		log.Fatal(err.Error())
	}

	ec2Api := InitEc2Api(cfg)

	publicIp, err := getPublicIpFromTaskEni(ctx, ec2Api, eni)
	if err != nil {
		log.Fatal(err.Error())
	}

	route53Api := InitRoute53Api(cfg)

	status, err := changeRoute53RecordSet(ctx, route53Api, subdomain, publicIp)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Change Route53 recordset status: %v\n", status)
}

func getTaskEni(ctx context.Context, ecsApi EcsApi, clusterName string) (string, error) {
	listTaskOutput, err := ecsApi.ListTasks(ctx, &ecs.ListTasksInput{Cluster: aws.String(clusterName)})
	if err != nil {
		return "", fmt.Errorf("error listing tasks of clusterName '%v': %v", clusterName, err)
	}
	log.Printf("Found tasks: %v\n", len(listTaskOutput.TaskArns))

	taskArn := listTaskOutput.TaskArns[0]
	log.Printf("First tasks arn: '%v'\n", taskArn)

	describeTasksInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterName),
		Tasks:   []string{taskArn},
	}

	describeTasksOutput, err := ecsApi.DescribeTasks(ctx, describeTasksInput)
	if err != nil {
		return "", fmt.Errorf("error describing task with arn '%v': %v", taskArn, err)
	}

	log.Printf("Found %v tasks'\n", len(describeTasksOutput.Tasks))

	for _, attachment := range describeTasksOutput.Tasks[0].Attachments {
		for _, detail := range attachment.Details {
			if *detail.Name == "networkInterfaceId" {
				log.Printf("The eni of the first task is '%v'", *detail.Value)

				return *detail.Value, nil
			}
		}
	}

	return "", fmt.Errorf("eni not found")
}

func getPublicIpFromTaskEni(ctx context.Context, ec2Api Ec2Api, taskEni string) (string, error) {
	describeNetworkInterfaceInput := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{taskEni},
	}

	describeNetworkInterfacesOutput, err := ec2Api.DescribeNetworkInterfaces(ctx, describeNetworkInterfaceInput)
	if err != nil {
		return "", fmt.Errorf("error describing network interface with id '%v': %v", taskEni, err)
	}

	return *describeNetworkInterfacesOutput.NetworkInterfaces[0].Association.PublicIp, nil
}

func changeRoute53RecordSet(ctx context.Context, route53Api Route53Api, subdomain string, publicIp string) (route53Types.ChangeStatus, error) {
	listHostedZonesOutput, err := route53Api.ListHostedZones(ctx, &route53.ListHostedZonesInput{})
	if err != nil {
		return "", fmt.Errorf("error listing hosted zones: %v", err)
	}

	hostedZoneId := listHostedZonesOutput.HostedZones[0].Id

	changeResourceRecordSetsInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53Types.ChangeBatch{
			Changes: []route53Types.Change{
				{
					Action: "UPSERT",
					ResourceRecordSet: &route53Types.ResourceRecordSet{
						Type: route53Types.RRTypeA,
						Name: aws.String(subdomain),
						TTL:  aws.Int64(300),
						ResourceRecords: []route53Types.ResourceRecord{
							{Value: aws.String(publicIp)},
						},
					},
				},
			},
		},
		HostedZoneId: hostedZoneId,
	}

	changeResourceRecordSetsOutput, err := route53Api.ChangeResourceRecordSets(ctx, changeResourceRecordSetsInput)
	if err != nil {
		return "", fmt.Errorf("error changing the resouce set in Route53 hosted zone '%v' with subdomain '%v': %v", *hostedZoneId, subdomain, err)
	}

	return changeResourceRecordSetsOutput.ChangeInfo.Status, nil
}
