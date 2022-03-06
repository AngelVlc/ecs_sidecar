package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
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

	publicIp := getPublicIp(context.TODO(), cfg, clusterName)
	log.Printf("Public ip: %v", publicIp)

	status := changeRoute53RecordSet(ctx, cfg, subdomain, publicIp)
	log.Printf("Change Route53 recordset status: %v\n", status)
}

func getPublicIp(ctx context.Context, cfg aws.Config, clusterName string) *string {
	ecsClient := ecs.NewFromConfig(cfg)

	listTaskOutput, err := ecsClient.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster: aws.String(clusterName),
	})
	if err != nil {
		log.Fatalf("error listing tasks of clusterName '%v': %v", clusterName, err)
	}

	log.Printf("Found tasks: %v\n", len(listTaskOutput.TaskArns))

	taskArn := listTaskOutput.TaskArns[0]
	log.Printf("First tasks arn: '%v'\n", taskArn)

	describeTasksOutput, err := ecsClient.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterName),
		Tasks:   []string{taskArn},
	})
	if err != nil {
		log.Fatalf("error describing task with arn '%v': %v", taskArn, err)
	}

	log.Printf("Found %v tasks'\n", len(describeTasksOutput.Tasks))
	taskEni := getTaskEni(describeTasksOutput.Tasks[0])

	log.Printf("The eni of the first task is '%v'", *taskEni)

	ec2Client := ec2.NewFromConfig(cfg)

	describeNetworkInterfacesOutput, err := ec2Client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{*taskEni},
	})
	if err != nil {
		log.Fatalf("error describing network interface with id '%v': %v", *taskEni, err)
	}

	return describeNetworkInterfacesOutput.NetworkInterfaces[0].Association.PublicIp
}

func getTaskEni(task types.Task) *string {
	for attachmentIndex, attachment := range task.Attachments {
		for detailIndex, detail := range attachment.Details {
			if *detail.Name == "networkInterfaceId" {
				return detail.Value
			}
			log.Printf("Attachment %v detail %v: name: '%v' - value: '%v'", attachmentIndex, detailIndex, *detail.Name, *detail.Value)
		}
	}

	return nil
}

func changeRoute53RecordSet(ctx context.Context, cfg aws.Config, subdomain string, publicIp *string) route53Types.ChangeStatus {
	route53Client := route53.NewFromConfig(cfg)

	listHostedZonesOutput, err := route53Client.ListHostedZones(ctx, &route53.ListHostedZonesInput{})
	if err != nil {
		log.Fatalf("error listing hosted zones: %v", err)
	}

	hostedZoneId := listHostedZonesOutput.HostedZones[0].Id

	changeResourceRecordSetsOutput, err := route53Client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53Types.ChangeBatch{
			Changes: []route53Types.Change{
				{
					Action: "UPSERT",
					ResourceRecordSet: &route53Types.ResourceRecordSet{
						Type: route53Types.RRTypeA,
						Name: aws.String(subdomain),
						TTL:  aws.Int64(300),
						ResourceRecords: []route53Types.ResourceRecord{
							{Value: publicIp},
						},
					},
				},
			},
		},
		HostedZoneId: hostedZoneId,
	})

	if err != nil {
		log.Fatalf("error changing the resouce set in Route53 hosted zone '%v' with subdomain '%v': %v", hostedZoneId, subdomain, err)
	}

	return changeResourceRecordSetsOutput.ChangeInfo.Status
}
