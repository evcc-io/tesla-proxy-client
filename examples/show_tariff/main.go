package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"

	tesla "github.com/evcc-io/tesla-proxy-client"
)

var tokenPath = flag.String("token", "", "path to token file")
var buyRate = flag.Float64("buy", -1, "buy rate; set custom tariff when both --buy and --sell are specified")
var sellRate = flag.Float64("sell", -1, "sell rate; set custom tariff when both --buy and --sell are specified")
var namedRate = flag.String("rate", "", "named utility tariff plan to set")

func main() {
	flag.Parse()

	if *tokenPath == "" {
		fmt.Println("--token must be specified")
		os.Exit(1)
	}

	if err := run(context.Background(), *tokenPath, *buyRate, *sellRate, *namedRate); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, tokenPath string, buy, sell float64, rate string) error {
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
		if p.EnergySiteId == 0 {
			continue
		}
		if i > 0 {
			fmt.Println("----")
		}

		es, err := c.EnergySite(p.EnergySiteId)
		if err != nil {
			return fmt.Errorf("error fetching site info: %w", err)
		}

		fmt.Printf("TariffID: %s\n", es.TariffID)
		printCurrentRates(es)

		if rate != "" {
			if err := es.SetNamedTariff(rate); err != nil {
				return fmt.Errorf("error setting named tariff: %w", err)
			}
			fmt.Printf("Named tariff set to: %s\n", rate)
		} else if buy >= 0 && sell >= 0 {
			if err := es.SetCustomTariff(buy, sell); err != nil {
				return fmt.Errorf("error setting custom tariff: %w", err)
			}
			fmt.Printf("Custom tariff set: buy=%.4f sell=%.4f\n", buy, sell)
		}
	}
	return nil
}

func printCurrentRates(es *tesla.EnergySite) {
	now := time.Now()
	month := int(now.Month())
	day := now.Day()
	weekday := int(now.Weekday())
	hour := now.Hour()
	minute := now.Minute()

	isWeekend := weekday == 0 || weekday == 6 // Sunday=0, Saturday=6

	if es.TariffContentV2 != nil {
		tc := es.TariffContentV2
		for seasonName, season := range tc.Seasons {
			if !inRange(month, season.FromMonth, season.ToMonth) || !inRange(day, season.FromDay, season.ToDay) {
				continue
			}
			for periodName, touPeriods := range season.TOUPeriods {
				for _, p := range touPeriods.Periods {
					if !inRange(weekday, p.FromDayOfWeek, p.ToDayOfWeek) {
						continue
					}
					if !inTimeRange(hour, minute, p.FromHour, p.FromMinute, p.ToHour, p.ToMinute) {
						continue
					}
					// buy rate: EnergyCharges[seasonName].Rates[periodName], falling back to "ALL" keys
					// for flat-rate tariffs where charges are keyed by "ALL" rather than season name
					var buyVal float64
					if ec, ok := tc.EnergyCharges[seasonName]; ok {
						if buyVal = ec.Rates[periodName]; buyVal == 0 {
							buyVal = ec.Rates["ALL"]
						}
					} else if ec, ok := tc.EnergyCharges["ALL"]; ok {
						if buyVal = ec.Rates[periodName]; buyVal == 0 {
							buyVal = ec.Rates["ALL"]
						}
					}
					// sell rate: SellTariff.EnergyCharges[seasonName|monthName|"ALL"].Rates[periodName|hourKey|"ALL"]
					var sellVal float64
					if tc.SellTariff != nil {
						monthName := now.Month().String()
						dayType := "weekday"
						if isWeekend {
							dayType = "weekend"
						}
						hourKey := fmt.Sprintf("hour_%d_%s", hour, dayType)
						for _, key := range []string{seasonName, monthName, "ALL"} {
							ec, ok := tc.SellTariff.EnergyCharges[key]
							if !ok {
								continue
							}
							if v := ec.Rates[periodName]; v != 0 {
								sellVal = v
							} else if v := ec.Rates[hourKey]; v != 0 {
								sellVal = v
							} else if v := ec.Rates["ALL"]; v != 0 {
								sellVal = v
							}
							if sellVal != 0 {
								break
							}
						}
					}
					fmt.Printf("Period: %s  buy=%.4f  sell=%.4f\n", periodName, buyVal, sellVal)
					return
				}
			}
		}
		fmt.Println("Period: (no matching period found)")
		return
	}

	if es.TariffContent != nil {
		tc := es.TariffContent
		for seasonName, season := range tc.Seasons {
			if !inRange(month, season.FromMonth, season.ToMonth) || !inRange(day, season.FromDay, season.ToDay) {
				continue
			}
			for periodName, periods := range season.TOUPeriods {
				for _, p := range periods {
					if !inRange(weekday, p.FromDayOfWeek, p.ToDayOfWeek) {
						continue
					}
					if !inTimeRange(hour, minute, p.FromHour, p.FromMinute, p.ToHour, p.ToMinute) {
						continue
					}
					// buy rate: EnergyCharges[seasonName][periodName], falling back to "ALL" keys
					var buyVal float64
					if ec, ok := tc.EnergyCharges[seasonName]; ok {
						if buyVal = ec[periodName]; buyVal == 0 {
							buyVal = ec["ALL"]
						}
					} else if ec, ok := tc.EnergyCharges["ALL"]; ok {
						if buyVal = ec[periodName]; buyVal == 0 {
							buyVal = ec["ALL"]
						}
					}
					// sell rate: SellTariff.EnergyCharges[seasonName|monthName|"ALL"][periodName|hourKey|"ALL"]
					var sellVal float64
					if tc.SellTariff != nil {
						monthName := now.Month().String()
						dayType := "weekday"
						if isWeekend {
							dayType = "weekend"
						}
						hourKey := fmt.Sprintf("hour_%d_%s", hour, dayType)
						for _, key := range []string{seasonName, monthName, "ALL"} {
							ec, ok := tc.SellTariff.EnergyCharges[key]
							if !ok {
								continue
							}
							if v := ec[periodName]; v != 0 {
								sellVal = v
							} else if v := ec[hourKey]; v != 0 {
								sellVal = v
							} else if v := ec["ALL"]; v != 0 {
								sellVal = v
							}
							if sellVal != 0 {
								break
							}
						}
					}
					fmt.Printf("Period: %s  buy=%.4f  sell=%.4f\n", periodName, buyVal, sellVal)
					return
				}
			}
		}
		fmt.Println("Period: (no matching period found)")
		return
	}

	fmt.Println("Period: (no tariff content available)")
}

// inRange returns true if v is within [lo, hi] inclusive, handling wrap-around when lo > hi.
func inRange(v, lo, hi int) bool {
	if lo <= hi {
		return v >= lo && v <= hi
	}
	// wrap-around (e.g. month 10–5 meaning Oct through May)
	return v >= lo || v <= hi
}

// inTimeRange returns true if hour:minute is in [fromHour:fromMinute, toHour:toMinute).
// toHour=0, toMinute=0 is treated as midnight (24:00 = end of day).
func inTimeRange(hour, minute, fromHour, fromMinute, toHour, toMinute int) bool {
	t := hour*60 + minute
	from := fromHour*60 + fromMinute
	to := toHour*60 + toMinute
	if to == 0 {
		to = 24 * 60
	}
	return t >= from && t < to
}
