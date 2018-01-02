package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	dstm "github.com/brimstone/dstm/types"
	"github.com/docker/docker/api/types/swarm"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
)

func init() {
	JoinCmd.Flags().StringP("token", "t", "", "The token to send to the server component")
	JoinCmd.Flags().StringP("server", "s", "", "Server component inside the swarm.")
	rootCmd.AddCommand(JoinCmd)
}

var JoinCmd = &cobra.Command{
	Use:   "join",
	Short: "Joins a node to a swarm",
	Long:  `Joins a node to an existing swarm.`,
	Run:   Join,
}

func Join(cmd *cobra.Command, args []string) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		fmt.Fprintf(os.Stderr, "Must specify a token with -t\n")
		os.Exit(1)
	}
	server, _ := cmd.Flags().GetString("server")
	if server == "" {
		fmt.Fprintf(os.Stderr, "Must specify a server with -s\n")
		os.Exit(1)
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	// Don't connect to swarm if this engine is already in a swarm
	_, err = client.InspectSwarm(context.Background())
	if err == nil {
		fmt.Fprintf(os.Stderr, "Node is in a swarm, exiting.\n")
		os.Exit(0)
	}

	if !strings.HasPrefix(server, "http") {
		server = "https://" + server
	}

	// Make request to server with dstm token.
	req, err := http.NewRequest("GET", server+"/v2/token", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting token: %s\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting token: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading the response body: %s\n", err)
		os.Exit(1)
	}

	// Decode response
	var clientToken dstm.Token
	err = json.Unmarshal(body, &clientToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading the response body: %s\n", err)
		os.Exit(1)
	}

	// TODO Connect to server with swarm token

	err = client.JoinSwarm(docker.JoinSwarmOptions{
		Context: context.Background(),
		JoinRequest: swarm.JoinRequest{
			RemoteAddrs: clientToken.Addresses,
			ListenAddr:  "0.0.0.0:2377",
			JoinToken:   clientToken.Token,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to join swarm: %s\n", err)
		os.Exit(1)
	}
}
