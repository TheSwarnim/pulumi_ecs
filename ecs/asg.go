package ecs

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EcsAutoScalingGroup struct {
	Name               string
	Cluster            *ecs.Cluster
	VpcZoneIdentifiers pulumi.StringArray
	DesiredCapacity    pulumi.Int
	MaxSize            pulumi.Int
	MinSize            pulumi.Int
	LaunchTemplate     *ec2.LaunchTemplate
}

func NewAutoScalingGroup(ctx *pulumi.Context, ecsAutoScalingGroup *EcsAutoScalingGroup) (*autoscaling.Group, error) {
	autoscalingGroup, err := autoscaling.NewGroup(ctx, ecsAutoScalingGroup.Name, &autoscaling.GroupArgs{
		VpcZoneIdentifiers: ecsAutoScalingGroup.VpcZoneIdentifiers,
		DesiredCapacity:    ecsAutoScalingGroup.DesiredCapacity,
		MaxSize:            ecsAutoScalingGroup.MaxSize,
		MinSize:            ecsAutoScalingGroup.MinSize,
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
					LaunchTemplateId: ecsAutoScalingGroup.LaunchTemplate.ID(),
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

	return autoscalingGroup, err
}
