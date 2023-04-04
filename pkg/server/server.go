package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	url               string
	secret            string
	token             string
	instanceId        string
	lastDeployAt      time.Time
	maxRunnerIdleTime time.Duration
}

type ActionsEC2ServerOptions struct {
	ECS               aws.ECSOptions
	EC2               aws.EC2Options
	URL               string
	Secret            string
	Token             string
	InstanceId        string
	MaxRunnerIdleTime time.Duration
}

func NewActionsEC2Server(o ActionsEC2ServerOptions) *ActionsEC2Server {
	return &ActionsEC2Server{
		ecs:               aws.NewECS(o.ECS),
		ec2:               aws.NewEC2(o.EC2),
		secret:            o.Secret,
		token:             o.Token,
		url:               o.URL,
		instanceId:        o.InstanceId,
		lastDeployAt:      time.Now(),
		maxRunnerIdleTime: o.MaxRunnerIdleTime,
	}
}

func (s *ActionsEC2Server) Initialize() error {
	log.Println("Initializing server...")
	log.Printf("- Max Runner Idle Time: %s\n", s.maxRunnerIdleTime)

	if s.instanceId == "" {
		log.Println("No instance id specified, creating one...")
		if err := s.CreateRunner(DeployRunnerOptions{URL: s.url}); err != nil {
			return fmt.Errorf("could not create runner: %s", err)
		}
	}

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

func (s *ActionsEC2Server) VerifySignature(payloadBody []byte, signatureHeader string) error {
	if signatureHeader == "" {
		return fmt.Errorf("x-hub-signature-256 header is missing")
	}

	hash := hmac.New(sha256.New, []byte(s.secret))
	hash.Write(payloadBody)
	expectedSignature := "sha256=" + hex.EncodeToString(hash.Sum(nil))

	if !hmac.Equal([]byte(expectedSignature), []byte(signatureHeader)) {
		return fmt.Errorf("request signatures didn't match")
	}

	return nil
}

func (s *ActionsEC2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Invalid method"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not read request body"))
		return
	}

	if s.secret != "" {
		if err := s.VerifySignature(body, r.Header.Get("X-Hub-Signature-256")); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
	}

	var payload github.WorkflowJobEvent
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse request body"))
		return
	}

	if err := s.Handle(payload); err != nil {
		log.Printf("[ERR] Error while handling request: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *ActionsEC2Server) Handle(payload github.WorkflowJobEvent) error {
	if payload.GetAction() == "queued" {
		log.Printf("[INFO] queued: %s - %s\n", payload.GetSender().GetLogin(), payload.GetWorkflowJob().GetWorkflowName())
		return s.DeployRunner(DeployRunnerOptions{
			URL: payload.Repo.GetHTMLURL(),
		})
	}
	return nil
}

func (s *ActionsEC2Server) Purge() error {
	running, err := s.RunnerIsRunning()
	if err != nil {
		return fmt.Errorf("could not check if runner is running: %s", err)
	}
	if !running {
		return nil
	}
	if s.lastDeployAt.Add(s.maxRunnerIdleTime).Before(time.Now()) {
		log.Printf("[INFO] Runner has been idle for %s, stopping it\n", s.maxRunnerIdleTime)
		return s.StopRunner()
	}
	return nil
}
