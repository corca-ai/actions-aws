package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/corca-ai/actions-ecs/pkg/aws"
	"github.com/google/go-github/v50/github"
)

type ActionsEC2Server struct {
	ecs *aws.ECS
	ec2 *aws.EC2
}

type ActionsEC2ServerOptions struct {
	ECS aws.ECSOptions
	EC2 aws.EC2Options
}

func NewActionsEC2Server(o ActionsEC2ServerOptions) *ActionsEC2Server {
	return &ActionsEC2Server{
		ecs: aws.NewECS(o.ECS),
		ec2: aws.NewEC2(o.EC2),
	}
}

func (s *ActionsEC2Server) Initialize() error {
	log.Println("Initializing server...")

	return nil
}

func (s *ActionsEC2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var payload github.PushEvent
	if body, err := io.ReadAll(r.Body); err != nil {
		log.Printf("[ERR] Could not read requested body: %s\n", err)
	} else {
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("[ERR] Could not parse requested body: %s\n", err)
		}
	}

	if err := s.Handle(payload); err != nil {
		log.Printf("[ERR] Error while handling request: %s\n", err)
	}

	io.WriteString(w, "OK\n")
}

func (s *ActionsEC2Server) Handle(payload github.PushEvent) error {
	if *payload.Action == "queued" {
		return s.DeployRunner()
	}
	return nil
}

func (s *ActionsEC2Server) DeployRunner() error {
	err := s.ec2.InitializeInstance()
	if err != nil {
		return fmt.Errorf("failed to initialize instance: %s", err)
	}
	return nil
}
