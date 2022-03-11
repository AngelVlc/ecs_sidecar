package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Setenv("TESTING", "true")

	os.Exit(m.Run())
}

func Test_GetTaskEni_ListTasks_Error(t *testing.T) {
	ctx := context.TODO()
	mockedEcsApi := NewMockedEcsApi()

	mockedEcsApi.On("ListTasks", ctx, &ecs.ListTasksInput{Cluster: aws.String("cluster")}).Return(nil, fmt.Errorf("some error"))

	result, err := getTaskEni(ctx, mockedEcsApi, "cluster")

	assert.Empty(t, result)
	assert.EqualError(t, err, "error listing tasks of clusterName 'cluster': some error")

	mockedEcsApi.AssertExpectations(t)
}

func Test_GetTaskEni_DescribeTasks_Error(t *testing.T) {
	ctx := context.TODO()
	mockedEcsApi := NewMockedEcsApi()

	listOutput := &ecs.ListTasksOutput{
		TaskArns: []string{"taskArn"},
	}

	mockedEcsApi.On("ListTasks", ctx, &ecs.ListTasksInput{Cluster: aws.String("cluster")}).Return(listOutput, nil)

	describeTasksInput := &ecs.DescribeTasksInput{
		Cluster: aws.String("cluster"),
		Tasks:   []string{"taskArn"},
	}

	mockedEcsApi.On("DescribeTasks", ctx, describeTasksInput).Return(nil, fmt.Errorf("some error"))

	result, err := getTaskEni(ctx, mockedEcsApi, "cluster")

	assert.Empty(t, result)
	assert.EqualError(t, err, "error describing task with arn 'taskArn': some error")

	mockedEcsApi.AssertExpectations(t)
}

func Test_GetTaskEni_DescribeTasks_EniNotFound(t *testing.T) {
	ctx := context.TODO()
	mockedEcsApi := NewMockedEcsApi()

	listOutput := &ecs.ListTasksOutput{
		TaskArns: []string{"taskArn"},
	}

	mockedEcsApi.On("ListTasks", ctx, &ecs.ListTasksInput{Cluster: aws.String("cluster")}).Return(listOutput, nil)

	describeTasksInput := &ecs.DescribeTasksInput{
		Cluster: aws.String("cluster"),
		Tasks:   []string{"taskArn"},
	}

	describeTasksOutput := &ecs.DescribeTasksOutput{
		Tasks: []ecsTypes.Task{
			{},
		},
	}

	mockedEcsApi.On("DescribeTasks", ctx, describeTasksInput).Return(describeTasksOutput, nil)

	result, err := getTaskEni(ctx, mockedEcsApi, "cluster")

	assert.Empty(t, result)
	assert.EqualError(t, err, "eni not found")

	mockedEcsApi.AssertExpectations(t)
}

func Test_GetTaskEni_DescribeTasks_Ok(t *testing.T) {
	ctx := context.TODO()
	mockedEcsApi := NewMockedEcsApi()

	listOutput := &ecs.ListTasksOutput{
		TaskArns: []string{"taskArn"},
	}

	mockedEcsApi.On("ListTasks", ctx, &ecs.ListTasksInput{Cluster: aws.String("cluster")}).Return(listOutput, nil)

	describeTasksInput := &ecs.DescribeTasksInput{
		Cluster: aws.String("cluster"),
		Tasks:   []string{"taskArn"},
	}

	describeTasksOutput := &ecs.DescribeTasksOutput{
		Tasks: []ecsTypes.Task{
			{
				Attachments: []ecsTypes.Attachment{
					{
						Details: []ecsTypes.KeyValuePair{
							{
								Name:  aws.String("networkInterfaceId"),
								Value: aws.String("taskEni"),
							},
						},
					},
				},
			},
		},
	}

	mockedEcsApi.On("DescribeTasks", ctx, describeTasksInput).Return(describeTasksOutput, nil)

	result, err := getTaskEni(ctx, mockedEcsApi, "cluster")

	assert.Equal(t, "taskEni", result)
	assert.Nil(t, err)

	mockedEcsApi.AssertExpectations(t)
}

func Test_ChangeRoute53RecordSet_ListHostedZones_Error(t *testing.T) {
	ctx := context.TODO()
	mockedRoute53Api := NewMockedRoute53Api()

	mockedRoute53Api.On("ListHostedZones", ctx, &route53.ListHostedZonesInput{}).Return(nil, fmt.Errorf("some error")).Once()

	result, err := changeRoute53RecordSet(ctx, mockedRoute53Api, "subdomain", "ip")

	assert.Empty(t, result)
	assert.EqualError(t, err, "error listing hosted zones: some error")

	mockedRoute53Api.AssertExpectations(t)
}

func Test_ChangeRoute53RecordSet_ChangeResourceRecordSets_Error(t *testing.T) {
	ctx := context.TODO()
	mockedRoute53Api := NewMockedRoute53Api()

	output := &route53.ListHostedZonesOutput{
		HostedZones: []route53Types.HostedZone{
			{
				Id: aws.String("hostedZoneId"),
			},
		},
	}

	mockedRoute53Api.On("ListHostedZones", ctx, &route53.ListHostedZonesInput{}).Return(output, nil).Once()

	changeInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53Types.ChangeBatch{
			Changes: []route53Types.Change{
				{
					Action: "UPSERT",
					ResourceRecordSet: &route53Types.ResourceRecordSet{
						Type: route53Types.RRTypeA,
						Name: aws.String("subdomain"),
						TTL:  aws.Int64(300),
						ResourceRecords: []route53Types.ResourceRecord{
							{Value: aws.String("ip")},
						},
					},
				},
			},
		},
		HostedZoneId: aws.String("hostedZoneId"),
	}

	mockedRoute53Api.On("ChangeResourceRecordSets", ctx, changeInput).Return(nil, fmt.Errorf("some error")).Once()

	result, err := changeRoute53RecordSet(ctx, mockedRoute53Api, "subdomain", "ip")

	assert.Empty(t, result)
	assert.EqualError(t, err, "error changing the resouce set in Route53 hosted zone 'hostedZoneId' with subdomain 'subdomain': some error")

	mockedRoute53Api.AssertExpectations(t)
}

func Test_ChangeRoute53RecordSet_ChangeResourceRecordSets_Ok(t *testing.T) {
	ctx := context.TODO()
	mockedRoute53Api := NewMockedRoute53Api()

	listHostedZonesOutput := &route53.ListHostedZonesOutput{
		HostedZones: []route53Types.HostedZone{
			{
				Id: aws.String("hostedZoneId"),
			},
		},
	}

	mockedRoute53Api.On("ListHostedZones", ctx, &route53.ListHostedZonesInput{}).Return(listHostedZonesOutput, nil).Once()

	changeInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53Types.ChangeBatch{
			Changes: []route53Types.Change{
				{
					Action: "UPSERT",
					ResourceRecordSet: &route53Types.ResourceRecordSet{
						Type: route53Types.RRTypeA,
						Name: aws.String("subdomain"),
						TTL:  aws.Int64(300),
						ResourceRecords: []route53Types.ResourceRecord{
							{Value: aws.String("ip")},
						},
					},
				},
			},
		},
		HostedZoneId: aws.String("hostedZoneId"),
	}

	changeResourceRecordSetsOutput := &route53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &route53Types.ChangeInfo{
			Status: route53Types.ChangeStatusPending,
		},
	}

	mockedRoute53Api.On("ChangeResourceRecordSets", ctx, changeInput).Return(changeResourceRecordSetsOutput, nil).Once()

	result, err := changeRoute53RecordSet(ctx, mockedRoute53Api, "subdomain", "ip")

	assert.Equal(t, route53Types.ChangeStatusPending, result)
	assert.Nil(t, err)

	mockedRoute53Api.AssertExpectations(t)
}

func Test_GetPublicIpFromTaskEni_DescribeNetworkInterfaces_Error(t *testing.T) {
	ctx := context.TODO()
	mockedEc2Api := NewMockedEc2Api()

	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{"taskEni"},
	}

	mockedEc2Api.On("DescribeNetworkInterfaces", ctx, input).Return(nil, fmt.Errorf("some error"))

	result, err := getPublicIpFromTaskEni(ctx, mockedEc2Api, "taskEni")

	assert.Empty(t, result)
	assert.EqualError(t, err, "error describing network interface with id 'taskEni': some error")

	mockedEc2Api.AssertExpectations(t)
}

func Test_GetPublicIpFromTaskEni_DescribeNetworkInterfaces_Ok(t *testing.T) {
	ctx := context.TODO()
	mockedEc2Api := NewMockedEc2Api()

	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{"taskEni"},
	}

	output := &ec2.DescribeNetworkInterfacesOutput{
		NetworkInterfaces: []ec2Types.NetworkInterface{
			{
				Association: &ec2Types.NetworkInterfaceAssociation{
					PublicIp: aws.String("publicIp"),
				},
			},
		},
	}

	mockedEc2Api.On("DescribeNetworkInterfaces", ctx, input).Return(output, nil)

	result, err := getPublicIpFromTaskEni(ctx, mockedEc2Api, "taskEni")

	assert.Equal(t, "publicIp", result)
	assert.Nil(t, err)

	mockedEc2Api.AssertExpectations(t)
}
