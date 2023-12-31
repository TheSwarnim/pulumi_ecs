package main

import (
	ecs2 "ecs/ecs"
	"ecs/utils"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/alb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
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

		// create a new cluster
		cluster, err := ecs2.NewCluster(ctx)
		if err != nil {
			return err
		}

		// Get the latest Amazon ECS-optimized AMI ID using SSM Parameter
		amiID, err := utils.GetLatestEcsOptimizedAmiId(ctx)

		/*
			first create the launch template
			then create the autoscaling group
			then create the capacity provider
		*/
		clusterSecurityGroupIds := pulumi.StringArray{
			pulumi.String("sg-0d4bd211820e90b03"),
		}

		// create a new launch template
		launchTemplate, err := ecs2.NewLaunchTemplate(ctx, &ecs2.EcsLaunchTemplate{
			Name:                          "pulumi-ecs-launch-template",
			Cluster:                       cluster,
			AmiID:                         amiID,
			VolumeSize:                    50,
			VolumeType:                    "gp3",
			NetworkInterfaceSecurityGroup: clusterSecurityGroupIds,
			NetworkInterfaceSubnetID:      pulumi.StringPtr("subnet-0bad1990bdb6919ec"),
			InstanceProfileArn:            pulumi.StringPtr("arn:aws:iam::369737379577:instance-profile/ecsInstanceRole"),
			KeyName:                       pulumi.StringPtr("swarnim-dev"),
		})
		if err != nil {
			return err
		}

		//create a new autoscaling group
		clusterSubnetIds := pulumi.StringArray{
			pulumi.String("subnet-027691384e95e1c10"),
			pulumi.String("subnet-0bad1990bdb6919ec"),
			pulumi.String("subnet-04fcf156adfca726e"),
		}

		autoscalingGroup, err := ecs2.NewAutoScalingGroup(ctx, &ecs2.EcsAutoScalingGroup{
			Name:               "pulumi-ecs-autoscaling-group",
			Cluster:            cluster,
			VpcZoneIdentifiers: clusterSubnetIds,
			DesiredCapacity:    pulumi.Int(1),
			MaxSize:            pulumi.Int(5),
			MinSize:            pulumi.Int(1),
			LaunchTemplate:     launchTemplate,
		})
		if err != nil {
			return err
		}

		// create a new capacity provider
		capacityProvider, err := ecs2.NewCapacityProvider(ctx, &ecs2.EcsCapacityProvider{
			Name:             "pulumi-ecs-capacity-provider",
			Cluster:          cluster,
			AutoScalingGroup: autoscalingGroup,
		})

		clusterVpcId := pulumi.StringPtr("vpc-01c036ec9248fed3a")

		// Create an Application Load Balancer.
		loadBalancer, err := alb.NewLoadBalancer(ctx, "redash-service-alb", &alb.LoadBalancerArgs{
			Subnets:        clusterSubnetIds,
			SecurityGroups: clusterSecurityGroupIds,
		})
		if err != nil {
			return err
		}

		// Create a target group.
		targetGroup, err := alb.NewTargetGroup(ctx, "redash-service-tg", &alb.TargetGroupArgs{
			Port:       pulumi.Int(80),
			Protocol:   pulumi.String("HTTP"),
			VpcId:      clusterVpcId,
			TargetType: pulumi.String("instance"),
			HealthCheck: &alb.TargetGroupHealthCheckArgs{
				HealthyThreshold:   pulumi.Int(3),
				Interval:           pulumi.Int(30),
				Timeout:            pulumi.Int(25),
				UnhealthyThreshold: pulumi.Int(2),
				Matcher:            pulumi.StringPtr("200-399"),
			},
		})
		if err != nil {
			return err
		}

		// Create a listener for the load balancer.
		_, err = alb.NewListener(ctx, "redash-server-app-listener", &alb.ListenerArgs{
			LoadBalancerArn: loadBalancer.Arn,
			Port:            pulumi.Int(80),
			DefaultActions: alb.ListenerDefaultActionArray{
				&alb.ListenerDefaultActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: targetGroup.Arn,
				},
			},
		})
		if err != nil {
			return err
		}

		// output the capacity provider name
		ctx.Export("capacityProviderName", capacityProvider.Name)

		// Set up the ECS Service using the Capacity Provider
		service, err := ecs.NewService(ctx, "pulumi-redash-server", &ecs.ServiceArgs{
			Cluster:        cluster.Arn,
			TaskDefinition: pulumi.StringPtr("arn:aws:ecs:ap-south-1:369737379577:task-definition/redash-server:5"),
			DesiredCount:   pulumi.Int(2),
			CapacityProviderStrategies: ecs.ServiceCapacityProviderStrategyArray{
				&ecs.ServiceCapacityProviderStrategyArgs{
					CapacityProvider: capacityProvider.Name,
					Weight:           pulumi.Int(100),
					// Optional: specify weight or base as appropriate
				},
			},
			//LaunchType: pulumi.String("EC2"), // either define launch type or capacity provider strategies
			LoadBalancers: ecs.ServiceLoadBalancerArray{
				&ecs.ServiceLoadBalancerArgs{
					TargetGroupArn: targetGroup.Arn,
					ContainerName:  pulumi.String("redash-server"), // Replace with your actual container name
					ContainerPort:  pulumi.Int(5000),               // Replace with your actual container port
				},
			},
		})
		if err != nil {
			return err
		}

		// print ecs service name
		ctx.Export("ecsServiceName", service.Name)

		return nil
	})
}
