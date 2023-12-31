package ecs

import (
	"ecs/utils"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// EcsLaunchTemplate is a struct that holds all the information required to create a new Launch Template for ECS
type EcsLaunchTemplate struct {
	Name                          string
	Cluster                       *ecs.Cluster
	AmiID                         string
	VolumeSize                    int
	VolumeType                    string
	NetworkInterfaceSecurityGroup pulumi.StringArray
	NetworkInterfaceSubnetID      pulumi.StringPtrInput
	InstanceProfileArn            pulumi.StringPtrInput
	KeyName                       pulumi.StringPtrInput
}

func NewLaunchTemplate(ctx *pulumi.Context, ecsLaunchTemplate *EcsLaunchTemplate) (*ec2.LaunchTemplate, error) {
	launchTemplate, err := ec2.NewLaunchTemplate(ctx, ecsLaunchTemplate.Name, &ec2.LaunchTemplateArgs{
		ImageId:  pulumi.String(ecsLaunchTemplate.AmiID),
		UserData: utils.Base64EncodedUserData(ecsLaunchTemplate.Cluster),
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
				SecurityGroups:           ecsLaunchTemplate.NetworkInterfaceSecurityGroup,
				DeleteOnTermination:      pulumi.StringPtr("true"),
				SubnetId:                 ecsLaunchTemplate.NetworkInterfaceSubnetID,
			},
		},
		IamInstanceProfile: &ec2.LaunchTemplateIamInstanceProfileArgs{
			Arn: ecsLaunchTemplate.InstanceProfileArn,
		},
		KeyName: ecsLaunchTemplate.KeyName,
	})

	return launchTemplate, err
}
