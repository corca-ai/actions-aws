package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var USER_DATA = "./resource/user-data.sh"

type EC2 struct {
	client *ec2.Client
}

type EC2Options struct {
	Region          string
	AccessKeyId     string
	SecretAccessKey string
}

func NewEC2(o EC2Options) *EC2 {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(o.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(o.AccessKeyId, o.SecretAccessKey, ""),
		),
	)
	if err != nil {
		panic(err)
	}
	return &EC2{
		client: ec2.NewFromConfig(cfg),
	}
}

func (e *EC2) GetUserData(url string, token string) (string, error) {
	f, err := os.Open(USER_DATA)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	s := string(b)
	s = strings.ReplaceAll(s, "{{ ACTIONS_RUNNER_VERSION }}", "2.303.0")
	s = strings.ReplaceAll(s, "{{ GITHUB_URL }}", url)
	s = strings.ReplaceAll(s, "{{ GITHUB_TOKEN }}", token)
	return s, nil
}

func (e *EC2) CreateInstance(userdata string) (*string, error) {
	ri, err := e.client.RunInstances(context.Background(), &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-04cebc8d6c4f297a3"), // x86 Ubuntu Server 22.04 LTS (HVM), SSD Volume Type
		InstanceType: types.InstanceTypeC6iXlarge,
		MaxCount:     aws.Int32(1),
		MinCount:     aws.Int32(1),
		UserData:     aws.String(base64.StdEncoding.EncodeToString([]byte(userdata))),
		KeyName:      aws.String("github-actions-runner"),
	})

	if err != nil {
		return nil, fmt.Errorf("could not create instance: %s", err)
	}

	instance := ri.Instances[0]

	log.Printf("Created instance: %s", *instance.InstanceId)

	return instance.InstanceId, nil
}

func (e *EC2) DescribeInstance(id string) (types.InstanceStateName, error) {
	status, err := e.client.DescribeInstanceStatus(context.Background(), &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{id},
	})
	if err != nil {
		return "", fmt.Errorf("could not describe instance: %s", err)
	}

	if len(status.InstanceStatuses) == 0 {
		return "", fmt.Errorf("could not find instance: %s", id)
	}

	return status.InstanceStatuses[0].InstanceState.Name, nil
}

func (e *EC2) StartInstance(id string) error {
	s, err := e.DescribeInstance(id)
	if err != nil {
		return err
	}

	if s == types.InstanceStateNameRunning {
		return nil
	}

	if s == types.InstanceStateNameStopping {
		return fmt.Errorf("instance is stopping")
	}

	_, err = e.client.StartInstances(context.Background(), &ec2.StartInstancesInput{
		InstanceIds: []string{id},
	})
	if err != nil {
		return fmt.Errorf("could not start instance: %s", err)
	}

	return nil
}

func (e *EC2) StopInstance(id string) error {
	s, err := e.DescribeInstance(id)
	if err != nil {
		return err
	}

	if s == types.InstanceStateNameStopped {
		return nil
	}

	if s == types.InstanceStateNameStopping {
		return nil
	}

	_, err = e.client.StopInstances(context.Background(), &ec2.StopInstancesInput{
		InstanceIds: []string{id},
	})
	if err != nil {
		return fmt.Errorf("could not stop instance: %s", err)
	}

	return nil
}
