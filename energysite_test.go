package tesla

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	EnergySiteInfoJSON = `{
		"response": {
			"id": "STE20240101-00001",
			"site_name": "My Energy Site",
			"backup_reserve_percent": 20,
			"default_real_mode": "self_consumption"
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
