package tesla

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TariffTOUPeriod struct {
	FromDayOfWeek int `json:"fromDayOfWeek"`
	ToDayOfWeek   int `json:"toDayOfWeek"`
	FromHour      int `json:"fromHour"`
	FromMinute    int `json:"fromMinute"`
	ToHour        int `json:"toHour"`
	ToMinute      int `json:"toMinute"`
}

type TariffSeason struct {
	FromDay    int                          `json:"fromDay"`
	ToDay      int                          `json:"toDay"`
	FromMonth  int                          `json:"fromMonth"`
	ToMonth    int                          `json:"toMonth"`
	TOUPeriods map[string][]TariffTOUPeriod `json:"tou_periods"`
}

type TariffSellTariff struct {
	DemandCharges map[string]map[string]float64 `json:"demand_charges"`
	EnergyCharges map[string]map[string]float64 `json:"energy_charges"`
	Seasons       map[string]TariffSeason       `json:"seasons"`
}

type TariffContent struct {
	Code          string                        `json:"code"`
	Name          string                        `json:"name"`
	Utility       string                        `json:"utility"`
	Currency      string                        `json:"currency"`
	DemandCharges map[string]map[string]float64 `json:"demand_charges"`
	EnergyCharges map[string]map[string]float64 `json:"energy_charges"`
	Seasons       map[string]TariffSeason       `json:"seasons"`
	SellTariff    *TariffSellTariff             `json:"sell_tariff"`
}

// TariffRates wraps a map of period/rate-name to rate value; v2 uses this
// as the value type for each season/period key in demand_charges and energy_charges.
type TariffRates struct {
	Rates map[string]float64 `json:"rates"`
}

// TariffSeasonV2TOUPeriods wraps the list of TOU periods for a named period in v2.
type TariffSeasonV2TOUPeriods struct {
	Periods []TariffTOUPeriod `json:"periods"`
}

type TariffSeasonV2 struct {
	FromDay    int                                 `json:"fromDay"`
	ToDay      int                                 `json:"toDay"`
	FromMonth  int                                 `json:"fromMonth"`
	ToMonth    int                                 `json:"toMonth"`
	TOUPeriods map[string]TariffSeasonV2TOUPeriods `json:"tou_periods"`
}

type TariffSellTariffV2 struct {
	DemandCharges map[string]TariffRates    `json:"demand_charges"`
	EnergyCharges map[string]TariffRates    `json:"energy_charges"`
	Seasons       map[string]TariffSeasonV2 `json:"seasons"`
}

type TariffContentV2 struct {
	Code          string                    `json:"code"`
	Name          string                    `json:"name"`
	Utility       string                    `json:"utility"`
	Currency      string                    `json:"currency"`
	Version       int                       `json:"version"`
	DemandCharges map[string]TariffRates    `json:"demand_charges"`
	EnergyCharges map[string]TariffRates    `json:"energy_charges"`
	Seasons       map[string]TariffSeasonV2 `json:"seasons"`
	SellTariff    *TariffSellTariffV2       `json:"sell_tariff"`
}

// this represents site_info endpoint
type EnergySite struct {
	ID                   string           `json:"id"`
	SiteName             string           `json:"site_name"`
	BackupReservePercent int64            `json:"backup_reserve_percent,omitempty"`
	DefaultRealMode      string           `json:"default_real_mode,omitempty"`
	TariffID             string           `json:"tariff_id,omitempty"`
	TariffContent        *TariffContent   `json:"tariff_content,omitempty"`
	TariffContentV2      *TariffContentV2 `json:"tariff_content_v2,omitempty"`

	productId int64
	c         *Client
}

type EnergySiteStatus struct {
	ResourceType      string  `json:"resource_type"`
	SiteName          string  `json:"site_name"`
	GatewayId         string  `json:"gateway_id"`
	PercentageCharged float64 `json:"percentage_charged"`
	BatteryType       string  `json:"battery_type"`
	BackupCapable     bool    `json:"backup_capable"`
	BatteryPower      int64   `json:"battery_power"`

	c *Client

	// These are no longer returned.
	// https://github.com/teslamotors/vehicle-command/issues/215
	// EnergyLeft        float64 `json:"energy_left"`
	// TotalPackEnergy   uint64  `json:"total_pack_energy"`
}

type EnergySiteHistory struct {
	SerialNumber string                        `json:"serial_number"`
	Period       string                        `json:"period"`
	TimeSeries   []EnergySiteHistoryTimeSeries `json:"time_series"`

	c *Client
}

type EnergySiteHistoryTimeSeries struct {
	Timestamp                           time.Time `json:"timestamp"`
	SolarEnergyExported                 float64   `json:"solar_energy_exported"`
	GeneratorEnergyExported             float64   `json:"generator_energy_exported"`
	GridEnergyImported                  float64   `json:"grid_energy_imported"`
	GridServicesEnergyImported          float64   `json:"grid_services_energy_imported"`
	GridServicesEnergyExported          float64   `json:"grid_services_energy_exported"`
	GridEnergyExportedFromSolar         float64   `json:"grid_energy_exported_from_solar"`
	GridEnergyExportedFromGenerator     float64   `json:"grid_energy_exported_from_generator"`
	GridEnergyExportedFromBattery       float64   `json:"grid_energy_exported_from_battery"`
	BatteryEnergyExported               float64   `json:"battery_energy_exported"`
	BatteryEnergyImportedFromGrid       float64   `json:"battery_energy_imported_from_grid"`
	BatteryEnergyImportedFromSolar      float64   `json:"battery_energy_imported_from_solar"`
	BatteryEnergyImportedFromGenerator  float64   `json:"battery_energy_imported_from_generator"`
	ConsumerEnergyImportedFromGrid      float64   `json:"consumer_energy_imported_from_grid"`
	ConsumerEnergyImportedFromSolar     float64   `json:"consumer_energy_imported_from_solar"`
	ConsumerEnergyImportedFromBattery   float64   `json:"consumer_energy_imported_from_battery"`
	ConsumerEnergyImportedFromGenerator float64   `json:"consumer_energy_imported_from_generator"`
}

type SiteInfoResponse struct {
	Response *EnergySite `json:"response"`
}

type SiteStatusResponse struct {
	Response *EnergySiteStatus `json:"response"`
}

type SiteHistoryResponse struct {
	Response *EnergySiteHistory `json:"response"`
}

type EnergySiteLiveStatus struct {
	SolarPower         float64 `json:"solar_power"`
	PercentageCharged  float64 `json:"percentage_charged"`
	BatteryPower       float64 `json:"battery_power"`
	LoadPower          float64 `json:"load_power"`
	GridPower          float64 `json:"grid_power"`
	GridServicesActive bool    `json:"grid_services_active"`
	GridStatus         string  `json:"grid_status"`
	IslandStatus       string  `json:"island_status"`
	StormModeActive    bool    `json:"storm_mode_active"`
	Timestamp          string  `json:"timestamp"`

	c *Client

	// These are no longer returned.
	// https://github.com/teslamotors/vehicle-command/issues/215
	// EnergyLeft         float64 `json:"energy_left"`
	// TotalPackEnergy    float64 `json:"total_pack_energy"`
}

type SiteLiveStatusResponse struct {
	Response *EnergySiteLiveStatus `json:"response"`
}

// return fetches the energy site for the given product ID
func (c *Client) EnergySite(productID int64) (*EnergySite, error) {
	siteInfoResponse := &SiteInfoResponse{}
	if err := c.getJSON(c.baseURL+"/energy_sites/"+strconv.FormatInt(productID, 10)+"/site_info", siteInfoResponse); err != nil {
		return nil, err
	}
	siteInfoResponse.Response.c = c
	siteInfoResponse.Response.productId = productID
	return siteInfoResponse.Response, nil
}

func (s *EnergySite) EnergySiteStatus() (*EnergySiteStatus, error) {
	siteStatusResponse := &SiteStatusResponse{}
	if err := s.c.getJSON(s.statusPath(), siteStatusResponse); err != nil {
		return nil, err
	}
	siteStatusResponse.Response.c = s.c
	return siteStatusResponse.Response, nil
}

type HistoryPeriod string

const (
	HistoryPeriodDay   HistoryPeriod = "day"
	HistoryPeriodWeek  HistoryPeriod = "week"
	HistoryPeriodMonth HistoryPeriod = "month"
	HistoryPeriodYear  HistoryPeriod = "year"
)

func (s *EnergySite) EnergySiteHistory(period HistoryPeriod) (*EnergySiteHistory, error) {
	historyResponse := &SiteHistoryResponse{}
	if err := s.c.getJSON(s.historyPath(period), historyResponse); err != nil {
		return nil, err
	}
	historyResponse.Response.c = s.c
	return historyResponse.Response, nil
}

func (s *EnergySite) EnergySiteLiveStatus() (*EnergySiteLiveStatus, error) {
	liveStatusResponse := &SiteLiveStatusResponse{}
	if err := s.c.getJSON(s.liveStatusPath(), liveStatusResponse); err != nil {
		return nil, err
	}
	liveStatusResponse.Response.c = s.c
	return liveStatusResponse.Response, nil
}

func (s *EnergySite) basePath() string {
	return strings.Join([]string{s.c.baseURL, "energy_sites", strconv.FormatInt(s.productId, 10)}, "/")
}

func (s *EnergySite) statusPath() string {
	return strings.Join([]string{s.basePath(), "site_status"}, "/")
}

func (s *EnergySite) historyPath(period HistoryPeriod) string {
	v := url.Values{}
	v.Set("kind", "energy")
	v.Set("period", string(period))

	return strings.Join([]string{s.basePath(), "history"}, "/") + fmt.Sprintf("?%s", v.Encode())
}

func (s *EnergySite) liveStatusPath() string {
	return strings.Join([]string{s.basePath(), "live_status"}, "/")
}

func (s *EnergySite) tariffPath() string {
	return strings.Join([]string{s.basePath(), "time_of_use_settings"}, "/")
}

func (s *EnergySite) operationPath() string {
	return strings.Join([]string{s.basePath(), "operation"}, "/")
}

func (s *EnergySite) gridImportExportPath() string {
	return strings.Join([]string{s.basePath(), "grid_import_export"}, "/")
}

func (s *EnergySite) SetBatteryReserve(percent uint64) error {
	url := s.basePath() + "/backup"
	payload := fmt.Sprintf(`{"backup_reserve_percent":%d}`, percent)
	body, err := s.sendCommand(url, []byte(payload))
	if err != nil {
		return err
	}
	return checkCommandResponse(body)
}

// SetOperatingMode sets the site operating mode.
// Valid modes: "self_consumption", "autonomous", "backup"
func (s *EnergySite) SetOperatingMode(mode string) error {
	payload, err := json.Marshal(map[string]string{"default_real_mode": mode})
	if err != nil {
		return err
	}
	body, err := s.c.post(s.operationPath(), payload)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return checkCommandResponse(body)
}

// SetGridCharging enables or disables charging from the grid.
func (s *EnergySite) SetGridCharging(enabled bool) error {
	payload, err := json.Marshal(map[string]bool{"disallow_charge_from_grid_with_solar_installed": !enabled})
	if err != nil {
		return err
	}
	body, err := s.c.post(s.gridImportExportPath(), payload)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return checkCommandResponse(body)
}

// SetGridExport sets the grid export rule.
// Valid modes: "battery_ok", "pv_only", "never"
func (s *EnergySite) SetGridExport(mode string) error {
	payload, err := json.Marshal(map[string]string{"customer_preferred_export_rule": mode})
	if err != nil {
		return err
	}
	body, err := s.c.post(s.gridImportExportPath(), payload)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return checkCommandResponse(body)
}

func checkCommandResponse(body []byte) error {
	var outer struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(body, &outer); err != nil {
		return err
	}
	if len(outer.Response) == 0 {
		return nil
	}
	var inner struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal([]byte(outer.Response), &inner); err != nil {
		return err
	}
	if inner.Code != 200 && inner.Code != 201 {
		return fmt.Errorf("command failed: %s", inner.Message)
	}
	return nil
}

// SetNamedTariff sets the energy site tariff to a named utility rate plan.
func (s *EnergySite) SetNamedTariff(tariffName string) error {
	payload, err := json.Marshal(map[string]any{
		"tariff":       tariffName,
		"tou_settings": map[string]any{},
	})
	if err != nil {
		return err
	}
	body, err := s.c.post(s.tariffPath(), payload)
	if err != nil {
		return err
	}
	return checkCommandResponse(body)
}

// SetCustomTariff sets a flat buy/sell rate on the energy site.
func (s *EnergySite) SetCustomTariff(buyPrice, sellPrice float64) error {
	payload, err := json.Marshal(map[string]interface{}{
		"tariff": "",
		"tou_settings": map[string]interface{}{
			"name":     "Custom",
			"utility":  "",
			"currency": "USD",
			"demand_charges": map[string]interface{}{
				"ALL": map[string]interface{}{"rates": map[string]float64{"ALL": 0}},
			},
			"energy_charges": map[string]interface{}{
				"ALL": map[string]interface{}{"rates": map[string]float64{"ALL": buyPrice}},
			},
			"seasons": map[string]interface{}{
				"All Year": map[string]interface{}{
					"fromDay": 1, "toDay": 31, "fromMonth": 1, "toMonth": 12,
					"tou_periods": map[string]interface{}{
						"ALL": map[string]interface{}{
							"periods": []map[string]int{{"toDayOfWeek": 6}},
						},
					},
				},
			},
			"sell_tariff": map[string]interface{}{
				"demand_charges": map[string]interface{}{
					"ALL": map[string]interface{}{"rates": map[string]float64{"ALL": 0}},
				},
				"energy_charges": map[string]interface{}{
					"ALL": map[string]interface{}{"rates": map[string]float64{"ALL": sellPrice}},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	body, err := s.c.post(s.tariffPath(), payload)
	if err != nil {
		return err
	}
	return checkCommandResponse(body)
}

// Sends a command to the vehicle
func (s *EnergySite) sendCommand(url string, reqBody []byte) ([]byte, error) {
	body, err := s.c.post(url, reqBody)
	if err != nil {
		return nil, err
	}
	if len(body) > 0 {
		response := &CommandResponse{}
		if err := json.Unmarshal(body, response); err != nil {
			return nil, err
		}
		if !response.Response.Result && response.Response.Reason != "" {
			return nil, errors.New(response.Response.Reason)
		}
	}
	return body, nil
}
