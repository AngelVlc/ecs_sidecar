//go:build wireinject

package main

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/wire"
)

func InitEc2Api(cfg aws.Config) Ec2Api {
	if inTestingMode() {
		return initMockedEc2Api()
	} else {
		return initAwsEc2Api(cfg)
	}
}

func initAwsEc2Api(cfg aws.Config) Ec2Api {
	wire.Build(AwsEc2ApiSet)
	return nil
}

func initMockedEc2Api() Ec2Api {
	wire.Build(MockedEc2ApiSet)
	return nil
}

func InitEcsApi(cfg aws.Config) EcsApi {
	if inTestingMode() {
		return initMockedEcsApi()
	} else {
		return initAwsEcsApi(cfg)
	}
}

func initAwsEcsApi(cfg aws.Config) EcsApi {
	wire.Build(AwsEcsApiSet)
	return nil
}

func initMockedEcsApi() EcsApi {
	wire.Build(MockedEcsApiSet)
	return nil
}

func InitRoute53Api(cfg aws.Config) Route53Api {
	if inTestingMode() {
		return initMockedRoute53Api()
	} else {
		return initAwsRoute53Api(cfg)
	}
}

func initAwsRoute53Api(cfg aws.Config) Route53Api {
	wire.Build(AwsRoute53ApiSet)
	return nil
}

func initMockedRoute53Api() Route53Api {
	wire.Build(MockedRoute53ApiSet)
	return nil
}

func inTestingMode() bool {
	return len(os.Getenv("TESTING")) > 0
}

var MockedEc2ApiSet = wire.NewSet(
	NewMockedEc2Api,
	wire.Bind(new(Ec2Api), new(*MockedEc2Api)),
)

var AwsEc2ApiSet = wire.NewSet(
	NewAwsEc2Api,
	wire.Bind(new(Ec2Api), new(*AwsEc2Api)),
)

var MockedEcsApiSet = wire.NewSet(
	NewMockedEcsApi,
	wire.Bind(new(EcsApi), new(*MockedEcsApi)),
)

var AwsEcsApiSet = wire.NewSet(
	NewAwsEcsApi,
	wire.Bind(new(EcsApi), new(*AwsEcsApi)),
)

var MockedRoute53ApiSet = wire.NewSet(
	NewMockedRoute53Api,
	wire.Bind(new(Route53Api), new(*MockedRoute53Api)),
)

var AwsRoute53ApiSet = wire.NewSet(
	NewAwsRoute53Api,
	wire.Bind(new(Route53Api), new(*AwsRoute53Api)),
)
