package server

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type DeployRunnerOptions struct {
	URL string
}

func (s *ActionsEC2Server) DeployRunner(o DeployRunnerOptions) error {
	if err := s.WaitForRunnerStateChangable(s.runnerWaitTimeout); err != nil {
		return err
	}

	if err := s.StartRunner(o); err != nil {
		return err
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

func (s *ActionsEC2Server) RunnerIsRunning() (bool, error) {
	state, err := s.ec2.DescribeInstance(s.instanceId)
	if err != nil {
		return false, fmt.Errorf("failed to describe instance: %s", err)
	}

	return state == types.InstanceStateNameRunning, nil
}

func (s *ActionsEC2Server) RunnerIsStopping() (bool, error) {
	state, err := s.ec2.DescribeInstance(s.instanceId)
	if err != nil {
		return false, fmt.Errorf("failed to describe instance: %s", err)
	}

	return state == types.InstanceStateNameStopping, nil
}

func (s *ActionsEC2Server) WaitForRunnerStateChangable(timeout time.Duration) error {
	result := make(chan error, 1)
	go func() {
		for {
			stopping, err := s.RunnerIsStopping()
			if err != nil {
				result <- err
			}
			if !stopping {
				result <- nil
			}
		}
	}()

	if timeout > 0 {
		timer := time.NewTimer(timeout)
		select {
		case err := <-result:
			if !timer.Stop() {
				<-timer.C
			}
			return err
		case <-timer.C:
			return fmt.Errorf("timeout exceeded while waiting for instance to be changable (duration: %v)", timeout)
		}
	} else {
		err := <-result
		return err
	}
}
