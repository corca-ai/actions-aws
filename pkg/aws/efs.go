package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/efs"
)

type EFS struct {
	client *efs.Client
}

type EFSOptions struct {
	Region          string
	AccessKeyId     string
	SecretAccessKey string
}

func NewEFS(o EFSOptions) *EFS {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(o.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(o.AccessKeyId, o.SecretAccessKey, ""),
		),
	)
	if err != nil {
		panic(err)
	}
	return &EFS{
		client: efs.NewFromConfig(cfg),
	}
}

func (e *EFS) CreateFileSystem() error {
	_, err := e.client.CreateFileSystem(context.Background(), &efs.CreateFileSystemInput{})

	if err != nil {
		return fmt.Errorf("could not create file system: %s", err)
	}

	return nil
}
