package main

import (
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
		*/
		_, err := ecs.NewCluster(ctx, "pulumi-ecs-cluster", nil)
		if err != nil {
			return err
		}

		return nil
	})
}
