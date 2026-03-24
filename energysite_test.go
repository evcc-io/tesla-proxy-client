package tesla

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	EnergySiteInfoJSON = `{
		"response": {
			"id": "STE20240101-00001",
			"site_name": "My Energy Site",
			"backup_reserve_percent": 20,
			"default_real_mode": "self_consumption",
			"tariff_id": "PGE-E-ELEC-NEM3-2026",
			"tariff_content": {
				"code": "PGE-E-ELEC-2026-CA-NEM3-flag",
				"name": "Residential TOU",
				"utility": "Pacific Gas & Electric Co",
				"currency": "USD",
				"demand_charges": {"ALL": {"ALL": 0}},
				"energy_charges": {
					"Summer": {"ON_PEAK": 0.57908, "OFF_PEAK": 0.36052}
				},
				"seasons": {
					"Summer": {
						"fromDay": 1, "toDay": 30, "fromMonth": 6, "toMonth": 9,
						"tou_periods": {
							"ON_PEAK": [{"fromDayOfWeek": 0, "toDayOfWeek": 6, "fromHour": 16, "toHour": 21}]
						}
					}
				}
			},
			"tariff_content_v2": {
				"code": "PGE-E-ELEC-2026-CA-NEM3-flag",
				"name": "Residential TOU",
				"utility": "Pacific Gas & Electric Co",
				"currency": "USD",
				"version": 1,
				"demand_charges": {"ALL": {"rates": {"ALL": 0}}},
				"energy_charges": {
					"Summer": {"rates": {"ON_PEAK": 0.57908, "OFF_PEAK": 0.36052}}
				},
				"seasons": {
					"Summer": {
						"fromDay": 1, "toDay": 30, "fromMonth": 6, "toMonth": 9,
						"tou_periods": {
							"ON_PEAK": {"periods": [{"toDayOfWeek": 6, "fromHour": 16, "toHour": 21}]}
						}
					}
				}
			}
		}
	}`

	EnergySiteLiveStatusJSON = `{
		"response": {
			"solar_power": 4250.5,
			"percentage_charged": 92.6,
			"battery_power": -1250.0,
			"load_power": 3000.5,
			"grid_power": 0.0,
			"grid_services_active": false,
			"grid_status": "SystemGridConnected",
			"island_status": "SystemIslandStatusOnGrid",
			"storm_mode_active": false,
			"timestamp": "2024-01-01T12:00:00.000000Z"
		}
	}`
)

func TestEnergySiteSpec(t *testing.T) {
	ts := serveHTTP(t)
	defer ts.Close()

	client := NewTestClient(ts)

	testMux.HandleFunc("/api/1/energy_sites/12345/site_info", serveJSON(EnergySiteInfoJSON))
	testMux.HandleFunc("/api/1/energy_sites/12345/live_status", serveJSON(EnergySiteLiveStatusJSON))

	Convey("Should get energy site", t, func() {
		energySite, err := client.EnergySite(12345)
		So(err, ShouldBeNil)
		So(energySite.ID, ShouldEqual, "STE20240101-00001")
		So(energySite.SiteName, ShouldEqual, "My Energy Site")
		So(energySite.BackupReservePercent, ShouldEqual, 20)
		So(energySite.TariffID, ShouldEqual, "PGE-E-ELEC-NEM3-2026")
		So(energySite.TariffContent, ShouldNotBeNil)
		So(energySite.TariffContent.EnergyCharges["Summer"]["ON_PEAK"], ShouldEqual, 0.57908)
		So(energySite.TariffContent.Seasons["Summer"].TOUPeriods["ON_PEAK"][0].FromHour, ShouldEqual, 16)
		So(energySite.TariffContentV2, ShouldNotBeNil)
		So(energySite.TariffContentV2.Version, ShouldEqual, 1)
		So(energySite.TariffContentV2.EnergyCharges["Summer"].Rates["ON_PEAK"], ShouldEqual, 0.57908)
		So(energySite.TariffContentV2.Seasons["Summer"].TOUPeriods["ON_PEAK"].Periods[0].FromHour, ShouldEqual, 16)
	})

	Convey("Should get energy site live status", t, func() {
		energySite, err := client.EnergySite(12345)
		So(err, ShouldBeNil)

		liveStatus, err := energySite.EnergySiteLiveStatus()
		So(err, ShouldBeNil)
		So(liveStatus.SolarPower, ShouldEqual, 4250.5)
		So(liveStatus.PercentageCharged, ShouldEqual, 92.6)
		So(liveStatus.BatteryPower, ShouldEqual, -1250.0)
		So(liveStatus.LoadPower, ShouldEqual, 3000.5)
		So(liveStatus.GridPower, ShouldEqual, 0.0)
		So(liveStatus.GridServicesActive, ShouldBeFalse)
		So(liveStatus.GridStatus, ShouldEqual, "SystemGridConnected")
		So(liveStatus.IslandStatus, ShouldEqual, "SystemIslandStatusOnGrid")
		So(liveStatus.StormModeActive, ShouldBeFalse)
		So(liveStatus.Timestamp, ShouldEqual, "2024-01-01T12:00:00.000000Z")
	})
}

var TariffCommandResponseJSON = `{"response":{"Message":"Updated","Code":201}}`

func TestSetTariff(t *testing.T) {
	ts := serveHTTP(t)
	defer ts.Close()

	client := NewTestClient(ts)

	var lastBody []byte
	testMux.HandleFunc("/api/1/energy_sites/12345/time_of_use_settings", func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastBody = body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(TariffCommandResponseJSON))
	})

	Convey("Should set named tariff", t, func() {
		energySite, err := client.EnergySite(12345)
		So(err, ShouldBeNil)
		err = energySite.SetNamedTariff("PGE-E-ELEC-NEM3-2026")
		So(err, ShouldBeNil)
		var m map[string]interface{}
		So(json.Unmarshal(lastBody, &m), ShouldBeNil)
		So(m["tariff"], ShouldEqual, "PGE-E-ELEC-NEM3-2026")
	})

	Convey("Should set custom tariff", t, func() {
		energySite, err := client.EnergySite(12345)
		So(err, ShouldBeNil)
		err = energySite.SetCustomTariff(0.30, 0.10)
		So(err, ShouldBeNil)
		var m map[string]interface{}
		So(json.Unmarshal(lastBody, &m), ShouldBeNil)
		So(m["tou_settings"], ShouldNotBeNil)
	})
}

var StormModeCommandResponseJSON = `{"response":{"code":201,"message":"Updated"}}`

func TestSetStormMode(t *testing.T) {
	ts := serveHTTP(t)
	defer ts.Close()

	client := NewTestClient(ts)

	var lastBody []byte
	testMux.HandleFunc("/api/1/energy_sites/12345/storm_mode", func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastBody = body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(StormModeCommandResponseJSON))
	})

	Convey("Should enable storm mode", t, func() {
		energySite, err := client.EnergySite(12345)
		So(err, ShouldBeNil)
		err = energySite.SetStormMode(true)
		So(err, ShouldBeNil)
		var m map[string]interface{}
		So(json.Unmarshal(lastBody, &m), ShouldBeNil)
		So(m["enabled"], ShouldEqual, true)
	})

	Convey("Should disable storm mode", t, func() {
		energySite, err := client.EnergySite(12345)
		So(err, ShouldBeNil)
		err = energySite.SetStormMode(false)
		So(err, ShouldBeNil)
		var m map[string]interface{}
		So(json.Unmarshal(lastBody, &m), ShouldBeNil)
		So(m["enabled"], ShouldEqual, false)
	})
}

func TestEnergySitePaths(t *testing.T) {
	ts := serveHTTP(t)
	defer ts.Close()

	client := NewTestClient(ts)
	energySite := &EnergySite{
		ID:        "STE20240101-00001",
		SiteName:  "Test Site",
		productId: 12345,
		c:         client,
	}

	Convey("Should have correct base path", t, func() {
		So(energySite.basePath(), ShouldEqual, ts.URL+"/api/1/energy_sites/12345")
	})

	Convey("Should have correct live status path", t, func() {
		So(energySite.liveStatusPath(), ShouldEqual, ts.URL+"/api/1/energy_sites/12345/live_status")
	})
}
