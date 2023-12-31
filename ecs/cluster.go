package ecs

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewCluster(ctx *pulumi.Context) (*ecs.Cluster, error) {
	cluster, err := ecs.NewCluster(ctx, "pulumi-ecs-cluster", nil)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
