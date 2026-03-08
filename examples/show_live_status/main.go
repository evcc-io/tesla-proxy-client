package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"

	tesla "github.com/evcc-io/tesla-proxy-client"
)

var tokenPath = flag.String("token", "", "path to token file")

func main() {
	flag.Parse()

	if *tokenPath == "" {
		log.Fatal("--token must be specified")
	}

	ctx := context.Background()

	b, err := os.ReadFile(*tokenPath)
	if err != nil {
		log.Fatal("Couldn't read token:", err)
	}
	var tok *oauth2.Token
	if err := json.Unmarshal(b, &tok); err != nil {
		log.Fatal("Invalid token file:", err)
	}

	client, err := tesla.NewClient(ctx, tesla.WithTokenSource(oauth2.StaticTokenSource(tok)))
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	products, err := client.Products()
	if err != nil {
		log.Fatal("Failed to get products:", err)
	}

	var energySiteID int64
	for _, product := range products {
		if product.ResourceType == "battery" || product.ResourceType == "solar" {
			energySiteID = product.EnergySiteId
			break
		}
	}

	if energySiteID == 0 {
		log.Fatal("No energy sites found")
	}

	energySite, err := client.EnergySite(energySiteID)
	if err != nil {
		log.Fatal("Failed to get energy site:", err)
	}

	fmt.Printf("Energy Site: %s\n", energySite.SiteName)
	fmt.Println("================")

	liveStatus, err := energySite.EnergySiteLiveStatus()
	if err != nil {
		log.Fatal("Failed to get live status:", err)
	}

	fmt.Printf("Solar Power:         %.1f W\n", liveStatus.SolarPower)
	fmt.Printf("Battery Power:       %.1f W\n", liveStatus.BatteryPower)
	fmt.Printf("Load Power:          %.1f W\n", liveStatus.LoadPower)
	fmt.Printf("Grid Power:          %.1f W\n", liveStatus.GridPower)
	fmt.Printf("Percentage Charged:  %.1f%%\n", liveStatus.PercentageCharged)
	fmt.Printf("Grid Status:         %s\n", liveStatus.GridStatus)
	fmt.Printf("Island Status:       %s\n", liveStatus.IslandStatus)
	fmt.Printf("Storm Mode Active:   %t\n", liveStatus.StormModeActive)
	fmt.Printf("Grid Services Active: %t\n", liveStatus.GridServicesActive)
	fmt.Printf("Timestamp:           %s\n", liveStatus.Timestamp)
}
