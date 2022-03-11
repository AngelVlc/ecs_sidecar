package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/stretchr/testify/mock"
)

type Ec2Api interface {
	DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error)
}

type AwsEc2Api struct {
	ec2Client *ec2.Client
}

func NewAwsEc2Api(cfg aws.Config) *AwsEc2Api {
	ec2Client := ec2.NewFromConfig(cfg)

	return &AwsEc2Api{ec2Client: ec2Client}
}

func (a *AwsEc2Api) DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return a.ec2Client.DescribeNetworkInterfaces(ctx, params)
}

type MockedEc2Api struct {
	mock.Mock
}

func NewMockedEc2Api() *MockedEc2Api {
	return &MockedEc2Api{}
}

func (m *MockedEc2Api) DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
	args := m.Called(ctx, params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeNetworkInterfacesOutput), args.Error(1)
}
