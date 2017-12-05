package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/handlers"
)

var client *docker.Client

func writeError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	fmt.Fprintf(w, err.Error())
}

func main() {

	var err error
	client, err = docker.NewClientFromEnv()

	if err != nil {
		panic(err)
	}
	_, err = client.InspectSwarm(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Node is not in swarm mode\n")
		os.Exit(1)
	}

	http.Handle("/v1/token/worker", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("AUTH_WORKER") != r.Header.Get("Authorization") {
			writeError(w, 403, errors.New("Invalid Auth"))
			return
		}
		swarm, err := client.InspectSwarm(context.Background())
		if err != nil {
			writeError(w, 503, errors.New("Unable to connect to swarm"))
			return
		}

		fmt.Fprintf(w, "%s\n", swarm.JoinTokens.Worker)
	})))

	http.Handle("/v1/token/manager", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("AUTH_MANAGER") != r.Header.Get("Authorization") {
			writeError(w, 403, errors.New("Invalid Auth"))
			return
		}
		swarm, err := client.InspectSwarm(context.Background())
		if err != nil {
			writeError(w, 503, errors.New("Unable to connect to swarm"))
			return
		}

		fmt.Fprintf(w, "%s\n", swarm.JoinTokens.Manager)
	})))

	fmt.Println("Ready to serve")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
