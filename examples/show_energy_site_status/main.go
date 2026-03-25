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
var reservePercentage = flag.Int64("backupPercentage", -1, "backup percentage to set")

// example that demos fetching of site information and optionally setting the battery reserve percentage for the site
func main() {
	flag.Parse()

	if *tokenPath == "" {
		fmt.Println("--token must be specified")
		os.Exit(1)
	}

	if err := run(context.Background(), *tokenPath, *reservePercentage); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, tokenPath string, reservePercentage int64) error {
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

	prods, err := c.Products()
	if err != nil {
		return err
	}

	for i, p := range prods {
		if i > 0 {
			fmt.Println("----")
		}
		fmt.Printf("ID: %s\n", p.ID)
		fmt.Printf("ResourceType %s\n", p.ResourceType)
		if p.EnergySiteId != 0 {
			fmt.Printf("EnergySiteId: %d\n", p.EnergySiteId)

			es, err := c.EnergySite(p.EnergySiteId)
			if err != nil {
				fmt.Printf("error fetching site info: %+v\n", err)
				os.Exit(1)
			}
			fmt.Printf("EnergySite: %+v\n", *es)

			esi, err := es.EnergySiteStatus()
			if err != nil {
				fmt.Printf("error fetching site status: %+v\n", err)
				os.Exit(1)
			}
			fmt.Printf("EnergySiteInfo: %+v\n", *esi)

			if reservePercentage != -1 {
				if err := es.SetBatteryReserve(uint64(reservePercentage)); err != nil {
					fmt.Printf("error setting battery reserve: %+v\n", err)
					os.Exit(1)
				}
			}
		}
	}
	return nil
}
