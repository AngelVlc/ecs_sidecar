package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/stretchr/testify/mock"
)

type EcsApi interface {
	DescribeTasks(ctx context.Context, params *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error)
}

type AwsEcsApi struct {
	ecsClient *ecs.Client
}

func NewAwsEcsApi(cfg aws.Config) *AwsEcsApi {
	ecsClient := ecs.NewFromConfig(cfg)

	return &AwsEcsApi{ecsClient: ecsClient}
}

func (a *AwsEcsApi) DescribeTasks(ctx context.Context, params *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	return a.ecsClient.DescribeTasks(ctx, params)
}

type MockedEcsApi struct {
	mock.Mock
}

func NewMockedEcsApi() *MockedEcsApi {
	return &MockedEcsApi{}
}

func (m *MockedEcsApi) DescribeTasks(ctx context.Context, params *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	args := m.Called(ctx, params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ecs.DescribeTasksOutput), args.Error(1)
}
