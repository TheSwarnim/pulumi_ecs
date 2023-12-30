package main

import (
	"encoding/base64"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/autoscaling"
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
		clusterSecurityGroupIds := pulumi.StringArray{
			pulumi.String("sg-0d4bd211820e90b03"),
		}

		// create a new launch template
		launchTemplate, err := ec2.NewLaunchTemplate(ctx, "pulumi-ecs-launch-template", &ec2.LaunchTemplateArgs{
			ImageId:  pulumi.String(parameter.Value),
			UserData: base64EncodedUserData(cluster),
			//VpcSecurityGroupIds:  clusterSecurityGroupIds,
			UpdateDefaultVersion: pulumi.Bool(true),
			BlockDeviceMappings: ec2.LaunchTemplateBlockDeviceMappingArray{
				&ec2.LaunchTemplateBlockDeviceMappingArgs{
					DeviceName: pulumi.String("/dev/xvda"),
					Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
						VolumeSize: pulumi.Int(50),
						VolumeType: pulumi.String("gp3"),
					},
				},
			},
			NetworkInterfaces: ec2.LaunchTemplateNetworkInterfaceArray{
				&ec2.LaunchTemplateNetworkInterfaceArgs{
					AssociatePublicIpAddress: pulumi.String("true"),
					SecurityGroups:           clusterSecurityGroupIds,
					DeleteOnTermination:      pulumi.StringPtr("true"),
					SubnetId:                 pulumi.StringPtr("subnet-0bad1990bdb6919ec"),
				},
			},
			IamInstanceProfile: &ec2.LaunchTemplateIamInstanceProfileArgs{
				Arn: pulumi.String("arn:aws:iam::369737379577:instance-profile/ecsInstanceRole"),
			},
			KeyName: pulumi.StringPtr("swarnim-dev"),
		})

		if err != nil {
			return err
		}

		// Output the launch template name
		ctx.Export("launchTemplateName", launchTemplate.Name)

		//create a new autoscaling group
		clusterSubnetIds := pulumi.StringArray{
			pulumi.String("subnet-027691384e95e1c10"),
			pulumi.String("subnet-0bad1990bdb6919ec"),
			pulumi.String("subnet-04fcf156adfca726e"),
		}

		autoscalingGroup, err := autoscaling.NewGroup(ctx, "pulumi-ecs-autoscaling-group", &autoscaling.GroupArgs{
			VpcZoneIdentifiers: clusterSubnetIds,
			DesiredCapacity:    pulumi.Int(1),
			MaxSize:            pulumi.Int(5),
			MinSize:            pulumi.Int(1),
			//LaunchTemplate: &autoscaling.GroupLaunchTemplateArgs{ // either we define this if the launch template has instance type defined, or we define the mixed instances policy
			//	Id: launchTemplate.ID(),
			//},
			DesiredCapacityType: pulumi.StringPtr("units"),
			MixedInstancesPolicy: &autoscaling.GroupMixedInstancesPolicyArgs{
				InstancesDistribution: &autoscaling.GroupMixedInstancesPolicyInstancesDistributionArgs{
					SpotAllocationStrategy: pulumi.StringPtr("price-capacity-optimized"),
				},
				LaunchTemplate: &autoscaling.GroupMixedInstancesPolicyLaunchTemplateArgs{
					LaunchTemplateSpecification: &autoscaling.GroupMixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationArgs{
						LaunchTemplateId: launchTemplate.ID(),
					},
					Overrides: &autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArray{
						&autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
							InstanceType: pulumi.String("c6i.large"),
						},
						&autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
							InstanceType: pulumi.String("c6i.xlarge"),
						},
						&autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
							InstanceType: pulumi.String("c6i.2xlarge"),
						},
						&autoscaling.GroupMixedInstancesPolicyLaunchTemplateOverrideArgs{
							InstanceType: pulumi.String("c6i.4xlarge"),
						},
					},
				},
			},
		})

		// output the autoscaling group name
		ctx.Export("autoscalingGroupName", autoscalingGroup.Name)

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
