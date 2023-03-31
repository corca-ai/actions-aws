package server

import (
	"fmt"
	"time"
)

type DeployRunnerOptions struct {
	URL string
}

func (s *ActionsEC2Server) DeployRunner(o DeployRunnerOptions) error {
	err := s.StartRunner(o)

	if err != nil {
		return fmt.Errorf("failed to deploy ec2 runner: %s", err)
	}

	s.lastDeployAt = time.Now()

	return nil
}

func (s *ActionsEC2Server) CreateRunner(o DeployRunnerOptions) error {
	userdata, err := s.ec2.GetUserData(o.URL, s.token)
	if err != nil {
		return fmt.Errorf("could not get task definition: %s", err)
	}

	instanceId, err := s.ec2.CreateInstance(userdata)
	if err != nil {
		return fmt.Errorf("failed to deploy ec2 runner: %s", err)
	}

	if instanceId == nil {
		return fmt.Errorf("failed to deploy ec2 runner: instance id is nil")
	}

	s.instanceId = *instanceId

	return nil
}

func (s *ActionsEC2Server) StartRunner(o DeployRunnerOptions) error {
	err := s.ec2.StartInstance(s.instanceId)
	if err != nil {
		return fmt.Errorf("failed to start ec2 runner: %s", err)
	}

	return nil
}

func (s *ActionsEC2Server) StopRunner() error {
	err := s.ec2.StopInstance(s.instanceId)
	if err != nil {
		return fmt.Errorf("failed to stop ec2 runner: %s", err)
	}

	return nil
}
