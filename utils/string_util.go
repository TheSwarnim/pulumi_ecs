package utils

import (
	"encoding/base64"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Base64EncodedUserData(cluster *ecs.Cluster) pulumi.StringOutput {
	encodedUserData := cluster.Name.ApplyT(func(name string) string {
		userData := "#!/bin/bash\necho ECS_CLUSTER=" + name + " >> /etc/ecs/ecs.config"
		return base64.StdEncoding.EncodeToString([]byte(userData))
	}).(pulumi.StringOutput)
	return encodedUserData
}
