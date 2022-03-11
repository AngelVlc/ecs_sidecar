package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/stretchr/testify/mock"
)

type Route53Api interface {
	ListHostedZones(ctx context.Context, params *route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error)
	ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
}

type AwsRoute53Api struct {
	route53Client *route53.Client
}

func NewAwsRoute53Api(cfg aws.Config) *AwsRoute53Api {
	route53Client := route53.NewFromConfig(cfg)

	return &AwsRoute53Api{route53Client: route53Client}
}

func (a *AwsRoute53Api) ListHostedZones(ctx context.Context, params *route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error) {
	return a.route53Client.ListHostedZones(ctx, params)
}

func (a *AwsRoute53Api) ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	return a.route53Client.ChangeResourceRecordSets(ctx, params)
}

type MockedRoute53Api struct {
	mock.Mock
}

func NewMockedRoute53Api() *MockedRoute53Api {
	return &MockedRoute53Api{}
}

func (m *MockedRoute53Api) ListHostedZones(ctx context.Context, params *route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error) {
	args := m.Called(ctx, params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*route53.ListHostedZonesOutput), args.Error(1)
}

func (m *MockedRoute53Api) ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	args := m.Called(ctx, params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*route53.ChangeResourceRecordSetsOutput), args.Error(1)
}
