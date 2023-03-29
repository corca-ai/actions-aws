package aws

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var TASK_DEF_JSON = "./resource/task-def.json"

type ECS struct {
	client      *ecs.Client
	cluster     *types.Cluster
	ClientId    string
	ClusterName string
	Region      string
}

type ECSOptions struct {
	ClientId        string
	ClusterName     string
	Region          string
	AccessKeyId     string
	SecretAccessKey string
}

func NewECS(o ECSOptions) *ECS {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(o.Region), config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(o.AccessKeyId, o.SecretAccessKey, ""),
	))
	if err != nil {
		panic(err)
	}
	return &ECS{
		client:      ecs.NewFromConfig(cfg),
		ClientId:    o.ClientId,
		ClusterName: o.ClusterName,
		Region:      o.Region,
	}
}

func (e *ECS) GetTaskDefinition() (*string, error) {
	f, err := os.Open(TASK_DEF_JSON)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	s := string(b)
	s = strings.ReplaceAll(s, "{{ AWS_CLIENT_ID }}", e.ClientId)
	s = strings.ReplaceAll(s, "{{ AWS_REGION }}", e.Region)
	return &s, nil
}

func (e *ECS) InitializeCluster() error {
	cc, err := e.client.CreateCluster(context.Background(), &ecs.CreateClusterInput{
		ClusterName: &e.ClusterName,
	})
	if err != nil {
		return fmt.Errorf("could not create cluster: %s", err)
	}
	e.cluster = cc.Cluster

	taskdef, err := e.GetTaskDefinition()
	if err != nil {
		return fmt.Errorf("could not get task definition: %s", err)
	}

	s, err := e.client.CreateService(context.Background(), &ecs.CreateServiceInput{
		Cluster:                 e.cluster.ClusterArn,
		ServiceName:             aws.String("actions-runner"),
		DesiredCount:            aws.Int32(1),
		DeploymentConfiguration: &types.DeploymentConfiguration{},
		DeploymentController: &types.DeploymentController{
			Type: types.DeploymentControllerTypeEcs,
		},
		LaunchType: types.LaunchTypeFargate,
		// ClientToken *string
		// EnableECSManagedTags bool
		// EnableExecuteCommand bool
		// HealthCheckGracePeriodSeconds *int32
		// NetworkConfiguration *types.NetworkConfiguration
		// PlacementConstraints []types.PlacementConstraint
		// PlacementStrategy []types.PlacementStrategy
		// PlatformVersion *string
		// Role *string
		// SchedulingStrategy types.SchedulingStrategy

		// PropagateTags types.PropagateTags
		// CapacityProviderStrategy []types.CapacityProviderStrategyItem
		// ServiceConnectConfiguration *types.ServiceConnectConfiguration
		// ServiceRegistries []types.ServiceRegistry
	})
	if err != nil {
		return fmt.Errorf("could not create service: %s", err)
	}

	cts, err := e.client.CreateTaskSet(context.Background(), &ecs.CreateTaskSetInput{
		Cluster:        e.cluster.ClusterArn,
		Service:        s.Service.ServiceArn,
		TaskDefinition: taskdef,
	})

	if err != nil {
		return fmt.Errorf("could not create task set: %s", err)
	}
	log.Println(cts)

	return nil
}
