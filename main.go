package main

import (
	"encoding/base64"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
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

		cluster, err := ecs.NewCluster(ctx, "pulumi-ecs-cluster", nil)
		if err != nil {
			return err
		}

		// Output the cluster name
		ctx.Export("clusterName", cluster.Name)

		// Get the latest Amazon ECS-optimized AMI ID using SSM Parameter
		parameter, err := ssm.LookupParameter(ctx, &ssm.LookupParameterArgs{
			Name: "/aws/service/ecs/optimized-ami/amazon-linux-2023/recommended/image_id", // Replace with your parameter name
		})
		if err != nil {
			return err
		}

		// Output the latest AMI ID
		ctx.Export("latestEcsOptimizedAmiId", pulumi.String(parameter.Value))

		/*
			first create the launch template
			then create the autoscaling group
			then create the capacity provider
		*/

		// create a new launch template
		launchTemplate, err := ec2.NewLaunchTemplate(ctx, "pulumi-ecs-launch-template", &ec2.LaunchTemplateArgs{
			ImageId:  pulumi.String(parameter.Value),
			UserData: base64EncodedUserData(cluster),
			SecurityGroupNames: pulumi.StringArray{
				pulumi.String("sg-0d4bd211820e90b03"),
			},
		})

		if err != nil {
			return err
		}

		// Output the launch template name
		ctx.Export("launchTemplateName", launchTemplate.Name)

		return nil
	})
}

func base64EncodedUserData(cluster *ecs.Cluster) pulumi.StringOutput {
	encodedUserData := cluster.Name.ApplyT(func(name string) string {
		userData := "#!/bin/bash\necho ECS_CLUSTER=" + name + " >> /etc/ecs/ecs.config"
		return base64.StdEncoding.EncodeToString([]byte(userData))
	}).(pulumi.StringOutput)
	return encodedUserData
}
