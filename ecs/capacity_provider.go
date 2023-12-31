package ecs

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EcsCapacityProvider struct {
	Name             string
	Cluster          *ecs.Cluster
	AutoScalingGroup *autoscaling.Group
}

func NewCapacityProvider(ctx *pulumi.Context, ecsCapacityProvider *EcsCapacityProvider) (*ecs.CapacityProvider, error) {
	// create a new capacity provider
	capacityProvider, err := ecs.NewCapacityProvider(ctx, ecsCapacityProvider.Name, &ecs.CapacityProviderArgs{
		AutoScalingGroupProvider: &ecs.CapacityProviderAutoScalingGroupProviderArgs{
			AutoScalingGroupArn: ecsCapacityProvider.AutoScalingGroup.Arn,
			ManagedScaling: &ecs.CapacityProviderAutoScalingGroupProviderManagedScalingArgs{
				// Define your managed scaling settings here.
				MaximumScalingStepSize: pulumi.Int(1000),
				MinimumScalingStepSize: pulumi.Int(1),
				Status:                 pulumi.String("ENABLED"),
				TargetCapacity:         pulumi.Int(75), // Target capacity is specified as a percentage
			},
			ManagedTerminationProtection: pulumi.String("DISABLED"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Attach the Capacity Provider to the ECS Cluster
	_, err = ecs.NewClusterCapacityProviders(ctx, "pulumi-ecs-capacity-providers", &ecs.ClusterCapacityProvidersArgs{
		ClusterName:       ecsCapacityProvider.Cluster.Name,
		CapacityProviders: pulumi.StringArray{capacityProvider.Name},
		DefaultCapacityProviderStrategies: ecs.ClusterCapacityProvidersDefaultCapacityProviderStrategyArray{
			&ecs.ClusterCapacityProvidersDefaultCapacityProviderStrategyArgs{
				CapacityProvider: capacityProvider.Name,
				Weight:           pulumi.Int(1),
				Base:             pulumi.Int(0),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return capacityProvider, nil
}
