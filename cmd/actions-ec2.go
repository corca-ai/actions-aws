package main

import (
	"log"
	"net/http"

	"github.com/corca-ai/actions-ecs/pkg/aws"
	"github.com/corca-ai/actions-ecs/pkg/server"
)

func main() {
	s := server.NewActionsEC2Server(server.ActionsEC2ServerOptions{
		EC2: aws.EC2Options{
			ClientId:        "",
			Region:          "",
			AccessKeyId:     "",
			SecretAccessKey: "",
		},
	})

	if err := s.Initialize(); err != nil {
		panic(err)
	}

	log.Println("Listening on :8000")
	if err := http.ListenAndServe(":8000", s); err != nil {
		log.Fatal(err)
	}
}
