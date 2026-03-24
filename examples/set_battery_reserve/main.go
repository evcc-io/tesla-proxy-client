package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/bogosj/tesla"
)

var tokenPath = flag.String("token", "", "path to token file")

// example that sets the battery reserve percentage to 20% for all energy sites
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
	c, err := tesla.NewClient(ctx, tesla.WithTokenFile(tokenPath))
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
		fmt.Printf("ResourceType: %s\n", p.ResourceType)

		if p.EnergySiteId != 0 {
			fmt.Printf("EnergySiteId: %d\n", p.EnergySiteId)

			es, err := c.EnergySite(p.EnergySiteId)
			if err != nil {
				fmt.Printf("error fetching site info: %+v\n", err)
				continue
			}
			fmt.Printf("Current BackupReservePercent: %d%%\n", es.BackupReservePercent)

			// Set battery reserve to 20%
			fmt.Println("Setting battery reserve to 20%...")
			if err := es.SetBatteryReserve(20); err != nil {
				fmt.Printf("error setting battery reserve: %+v\n", err)
				continue
			}
			fmt.Println("Battery reserve successfully set to 20%")
		}
	}
	return nil
}
