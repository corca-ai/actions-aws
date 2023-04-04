package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
	err := godotenv.Load()
	if err == nil {
		log.Println("Loaded .env file")
	}

	maxRunnerIdleTime, err := time.ParseDuration(Getenv("MAX_RUNNER_IDLE_TIME", "30m"))
	if err != nil {
		panic(fmt.Errorf("failed to parse MAX_RUNNER_IDLE_TIME: %v", err))
	}

	s := server.NewActionsEC2Server(server.ActionsEC2ServerOptions{
		EC2: aws.EC2Options{
			Region:          Getenv("AWS_REGION", "ap-northeast-2"),
			AccessKeyId:     Getenv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: Getenv("AWS_SECRET_ACCESS_KEY", ""),
		},
		InstanceId:        Getenv("AWS_EC2_INSTANCE_ID", ""),
		URL:               Getenv("GITHUB_URL", ""),
		Token:             Getenv("GITHUB_TOKEN", ""),
		MaxRunnerIdleTime: maxRunnerIdleTime,
	})

	if err := s.Initialize(); err != nil {
		panic(fmt.Errorf("failed to initialize server: %v", err))
	}

	log.Println("Listening on :8000")
	if err := http.ListenAndServe(":8000", s); err != nil {
		log.Fatal(err)
	}
}
