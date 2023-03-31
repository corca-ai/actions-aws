package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/corca-ai/actions-ecs/pkg/aws"
	"github.com/google/go-github/v50/github"
)

type ActionsEC2Server struct {
	ecs               *aws.ECS
	ec2               *aws.EC2
	token             string
	lastDeployAt      time.Time
	maxRunnerIdleTime time.Duration
}

type ActionsEC2ServerOptions struct {
	ECS   aws.ECSOptions
	EC2   aws.EC2Options
	Token string
}

func NewActionsEC2Server(o ActionsEC2ServerOptions) *ActionsEC2Server {
	return &ActionsEC2Server{
		ecs:               aws.NewECS(o.ECS),
		ec2:               aws.NewEC2(o.EC2),
		token:             o.Token,
		lastDeployAt:      time.Now(),
		maxRunnerIdleTime: 30 * time.Minute,
	}
}

func (s *ActionsEC2Server) Initialize() error {
	log.Println("Initializing server...")

	go func() {
		for {
			err := s.Purge()
			if err != nil {
				log.Printf("[ERR] Error while purging: %s\n", err)
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	return nil
}

func (s *ActionsEC2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	log.Printf("[INFO] Received request: %s %s\n", r.Method, r.URL.Path)

	var payload github.WorkflowJobEvent
	if body, err := io.ReadAll(r.Body); err != nil {
		log.Printf("[ERR] Could not read requested body: %s\n", err)
		return
	} else {
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("[ERR] Could not parse requested body: %s\n", err)
			return
		}
	}

	if err := s.Handle(payload); err != nil {
		log.Printf("[ERR] Error while handling request: %s\n", err)
		return
	}

	io.WriteString(w, "OK\n")
}

func (s *ActionsEC2Server) Handle(payload github.WorkflowJobEvent) error {
	if payload.GetAction() == "queued" {
		return s.DeployRunner(DeployRunnerOptions{
			URL: payload.Repo.GetHTMLURL(),
		})
	}
	return nil
}

func (s *ActionsEC2Server) Purge() error {
	if s.lastDeployAt.Add(s.maxRunnerIdleTime).Before(time.Now()) {
		log.Printf("[INFO] Runner has been idle for %s, stopping it\n", s.maxRunnerIdleTime)
		return s.StopRunner()
	}
	return nil
}
