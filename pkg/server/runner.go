package server

import (
	"fmt"
)

type DeployRunnerOptions struct {
	WorkflowID int64
	URL        string
}

func (s *ActionsEC2Server) DeployRunner(o DeployRunnerOptions) error {
	userdata, err := s.ec2.GetUserData(o.URL, s.token)
	if err != nil {
		return fmt.Errorf("could not get task definition: %s", err)
	}

	err = s.ec2.DeployEC2Runner(userdata)
	if err != nil {
		return fmt.Errorf("failed to deploy ec2 runner: %s", err)
	}
	return nil
}
