package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"golang.org/x/oauth2"

	tesla "github.com/evcc-io/tesla-proxy-client"
)

var tokenPath = flag.String("token", "", "path to token file")

func main() {
	flag.Parse()

	if *tokenPath == "" {
		fmt.Println("--token must be specified")
		os.Exit(1)
	}

	if err := run(context.Background(), *tokenPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, tokenPath string) error {
	b, err := os.ReadFile(tokenPath)
	if err != nil {
		return err
	}
	var tok *oauth2.Token
	if err := json.Unmarshal(b, &tok); err != nil {
		return err
	}

	c, err := tesla.NewClient(ctx, tesla.WithTokenSource(oauth2.StaticTokenSource(tok)))
	if err != nil {
		return err
	}

	v, err := c.Vehicles()
	if err != nil {
		return err
	}

	for i, v := range v {
		if i > 0 {
			fmt.Println("----")
		}
		fmt.Printf("VIN: %s\n", v.Vin)
		fmt.Printf("Name: %s\n", v.DisplayName)
	}
	return nil
}
