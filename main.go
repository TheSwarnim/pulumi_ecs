package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// create a new ecs cluster
		/*
			the cluster will have 1 capacity provider of 100% Spot
			and 1 capacity provider of 100% On-Demand
			with minimum and maximum of 1 and 5 instances

			To add a capacity provider to your ECS cluster with a mix of 100%
			Spot and 100% On-Demand instances, you'll need to define two AWS EC2
			Auto Scaling groups corresponding to your Spot and On-Demand configurations.
			Then, create two ECS Capacity Providers associated with these Auto Scaling
			groups and add them to your ECS cluster.
		*/

		_, err := ecs.NewCluster(ctx, "pulumi-ecs-cluster", nil)
		if err != nil {
			return err
		}

		// Get the latest Amazon ECS-optimized AMI ID using SSM Parameter
		parameter, err := ssm.LookupParameter(ctx, &ssm.LookupParameterArgs{
			Name: "/aws/service/ecs/optimized-ami/amazon-linux-2023/recommended/image_id", // Replace with your parameter name
		})
		if err != nil {
			return err
		}

		// Output the latest AMI ID
		ctx.Export("latestEcsOptimizedAmiId", pulumi.String(parameter.Value))
		return nil
	})
}
