package utils

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GetLatestEcsOptimizedAmiId(ctx *pulumi.Context) (string, error) {
	// Get the latest Amazon ECS-optimized AMI ID using SSM Parameter
	parameter, err := ssm.LookupParameter(ctx, &ssm.LookupParameterArgs{
		Name: "/aws/service/ecs/optimized-ami/amazon-linux-2023/recommended/image_id", // Replace with your parameter name
	})
	if err != nil {
		return "", err
	}
	return parameter.Value, nil
}
