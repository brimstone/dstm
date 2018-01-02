package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	dstm "github.com/brimstone/dstm/types"
	"github.com/brimstone/jwt/jwt"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/handlers"
	"github.com/spf13/cobra"
)

var client *docker.Client

func writeError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	fmt.Fprintf(w, err.Error())
}

// Serve Serve tokens to clients
func Serve(cmd *cobra.Command, args []string) {
	key := os.Getenv("KEY")
	if key == "" {
		fmt.Fprintf(os.Stderr, "Environment variable KEY must be specified.\n")
		os.Exit(1)
	}
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

	http.Handle("/v2/token", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Report our source location
		w.Header().Add("X-Source", "https://github.com/brimstone/dstm")
		// Report our LICENSE
		w.Header().Add("X-License", "AGPLv3 http://www.gnu.org/licenses/agpl-3.0.txt")

		bearer := strings.Split(r.Header.Get("Authorization"), " ")
		if len(bearer) != 2 || bearer[0] != "Bearer" {
			writeError(w, 403, errors.New("Invalid Auth"))
			return
		}

		payload := make(map[string]interface{})
		err := jwt.Verify(key, bearer[1], &payload)

		swarm, err := client.InspectSwarm(context.Background())
		if err != nil {
			writeError(w, 503, errors.New("Unable to connect to swarm"))
			return
		}

		// Create struct to send to client
		clientToken := dstm.Token{
			Addresses: []string{},
		}
		nodes, err := client.ListNodes(docker.ListNodesOptions{})
		for _, node := range nodes {
			// TODO if addr is 127.0.0.1, probably should replace with how the client got here
			clientToken.Addresses = append(clientToken.Addresses, node.Status.Addr+":2377")
		}

		// Add token to client
		if payload["manager"].(bool) {
			clientToken.Token = swarm.JoinTokens.Manager
		} else {
			clientToken.Token = swarm.JoinTokens.Worker
		}

		// Encode as JSON
		clientJSON, _ := json.Marshal(clientToken)
		// Send JSON to client
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s\n", clientJSON)

		// TODO Set timer to rotate tokens
	})))

	fmt.Println("Ready to serve on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
