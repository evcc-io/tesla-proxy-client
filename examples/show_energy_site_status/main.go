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
var operationMode = flag.String("operationMode", "", "operating mode: self_consumption, autonomous, backup")
var gridExport = flag.String("gridExport", "", "grid export rule: battery_ok, pv_only, never")
var gridCharging = flag.String("gridCharging", "", "grid charging enabled: true, false")

// example that demos fetching of site information and optionally setting the battery reserve percentage for the site
func main() {
	flag.Parse()

	if *tokenPath == "" {
		fmt.Println("--token must be specified")
		os.Exit(1)
	}

	if err := run(context.Background(), *tokenPath, *reservePercentage, *operationMode, *gridExport, *gridCharging); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, tokenPath string, reservePercentage int64, operationMode, gridExport, gridCharging string) error {
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
			if operationMode != "" {
				if err := es.SetOperatingMode(operationMode); err != nil {
					fmt.Printf("error setting operating mode: %+v\n", err)
					os.Exit(1)
				}
			}
			if gridExport != "" {
				if err := es.SetGridExport(gridExport); err != nil {
					fmt.Printf("error setting grid export: %+v\n", err)
					os.Exit(1)
				}
			}
			if gridCharging != "" {
				enabled := gridCharging == "true"
				if err := es.SetGridCharging(enabled); err != nil {
					fmt.Printf("error setting grid charging: %+v\n", err)
					os.Exit(1)
				}
			}
		}
	}
	return nil
}
