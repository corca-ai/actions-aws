package main

import (
	"log"
	"net/http"
	"os"

	"github.com/corca-ai/actions-ecs/pkg/aws"
	"github.com/corca-ai/actions-ecs/pkg/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err == nil {
		log.Println("Loaded .env file")
	}

	s := server.NewActionsEC2Server(server.ActionsEC2ServerOptions{
		EC2: aws.EC2Options{
			Region:          os.Getenv("AWS_REGION"),
			AccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			InstanceId:      os.Getenv("AWS_EC2_INSTANCE_ID"),
		},
		Token: os.Getenv("GITHUB_TOKEN"),
	})

	if err := s.Initialize(); err != nil {
		panic(err)
	}

	log.Println("Listening on :8000")
	if err := http.ListenAndServe(":8000", s); err != nil {
		log.Fatal(err)
	}
}
