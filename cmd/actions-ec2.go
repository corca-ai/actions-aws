package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/corca-ai/actions-ecs/pkg/aws"
	"github.com/corca-ai/actions-ecs/pkg/server"
	"github.com/joho/godotenv"
)

func Getenv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	err := godotenv.Load()
	if err == nil {
		log.Println("Loaded .env file")
	}

	maxRunnerIdleTime, err := time.ParseDuration(Getenv("MAX_RUNNER_IDLE_TIME", "30m"))
	if err != nil {
		panic(fmt.Errorf("failed to parse MAX_RUNNER_IDLE_TIME: %v", err))
	}

	runnerWaitTimeout, err := time.ParseDuration(Getenv("RUNNER_WAIT_TIMEOUT", "3m"))
	if err != nil {
		panic(fmt.Errorf("failed to parse RUNNER_WAIT_TIMEOUT: %v", err))
	}

	port, err := strconv.ParseInt(Getenv("PORT", "8000"), 10, 32)
	if err != nil {
		panic(fmt.Errorf("failed to parse PORT: %v", err))
	}

	s := server.NewActionsEC2Server(server.ActionsEC2ServerOptions{
		EC2: aws.EC2Options{
			Region:          Getenv("AWS_REGION", "ap-northeast-2"),
			AccessKeyId:     Getenv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: Getenv("AWS_SECRET_ACCESS_KEY", ""),
		},
		Secret:            Getenv("GITHUB_SECRET", ""),
		InstanceId:        Getenv("AWS_EC2_INSTANCE_ID", ""),
		URL:               Getenv("GITHUB_URL", ""),
		Token:             Getenv("GITHUB_TOKEN", ""),
		Labels:            append([]string{"self-hosted", "Linux", "X64"}, strings.Split(Getenv("LABELS", ""), ",")...),
		MaxRunnerIdleTime: maxRunnerIdleTime,
		RunnerWaitTimeout: runnerWaitTimeout,
	})

	if err := s.Initialize(); err != nil {
		panic(fmt.Errorf("failed to initialize server: %v", err))
	}

	log.Printf("Listening on :%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), s); err != nil {
		log.Fatal(err)
	}
}
