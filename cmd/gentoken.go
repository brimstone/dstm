package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/brimstone/jwt/jwt"
	"github.com/spf13/cobra"
)

func init() {
	GenTokenCmd.Flags().StringP("key", "k", "", "The secret key used to sign the token.")
	GenTokenCmd.Flags().BoolP("manager", "m", false, "Token allows nodes to join as managers.")
	rootCmd.AddCommand(GenTokenCmd)
}

var GenTokenCmd = &cobra.Command{
	Use:   "gentoken",
	Short: "Generates a token, signed by a key.",
	Long: `Generates a token signed by either an HMAC or RSA key.
Use genrsa to generate these keys.
`,
	Run: GenToken,
}

func GenToken(cmd *cobra.Command, args []string) {
	key, _ := cmd.Flags().GetString("key")
	if key == "" {
		fmt.Fprintf(os.Stderr, "Must specify a key with -k\n")
		os.Exit(1)
	}

	payload := make(map[string]interface{})
	payload["manager"] = false

	if ok, _ := cmd.Flags().GetBool("manager"); ok {
		payload["manager"] = true
	}
	marshalled, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	token, err := jwt.GenToken(key, marshalled)
	if err != nil {
		panic(err)
	}
	fmt.Println(token)
}
