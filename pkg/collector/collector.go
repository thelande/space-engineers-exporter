package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	vrage_client "github.com/thelande/space-engineers-exporter/pkg/vrage_client"
)

const namespace = "space_engineers"

func getSEDesc(name string, unit string, help string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, name, unit), help, labels, nil,
	)
}

var (
	serverInfoLabels = []string{"server_name", "world_name", "version", "server_id"}

	upDesc                = getSEDesc("", "up", "Did the remote API respond to the queries.", nil)
	serverInfoDesc        = getSEDesc("", "info", "Information about the server", serverInfoLabels)
	readyDesc             = getSEDesc("", "ready", "Is the server ready?", nil)
	upTimeDesc            = getSEDesc("up", "seconds", "The number of seconds that the server has been up.", nil)
	playerCountDesc       = getSEDesc("player", "count", "The number of players currently connected to the server.", nil)
	simulationCpuLoadDesc = getSEDesc("simulation_cpu_load", "percent", "The simulation thread CPU load.", nil)
	simulationSpeedDesc   = getSEDesc("", "simulation_speed", "The simulation speed factor.", nil)
	pcuUsedDesc           = getSEDesc("pcu_used", "total", "The total number of PCU used.", nil)
	piratePcuUsedDesc     = getSEDesc("pirate_pcu_used", "total", "The total number of PCU used by pirate factions.", nil)

	planetLabels = []string{"display_name", "entity_id", "x", "y", "z"}
	planetDesc   = getSEDesc("planet", "info", "Information about the planets.", planetLabels)
	asteroidDesc = getSEDesc("asteroid", "info", "Information about the fixed asteroids.", planetLabels)

	gridLabels    = []string{"powered", "grid_size"}
	pcuLabels     = []string{"powered", "grid_size", "owner"}
	gridCountDesc = getSEDesc("grid", "count", "The number of grids on the server.", gridLabels)
	pcuCountDesc  = getSEDesc("pcu", "count", "The number of PCUs used on the server.", pcuLabels)

	bannedPlayersDesc = getSEDesc("banned_player", "count", "The number of banned players.", nil)
	kickedPlayersDesc = getSEDesc("kicked_player", "count", "The number of kicked players.", nil)
	cheatersDesc      = getSEDesc("cheaters", "count", "The number of players marked as cheaters.", nil)
)

type Collector struct {
	client *vrage_client.VRageClient
	logger log.Logger
}

func NewCollector(client *vrage_client.VRageClient, logger log.Logger) Collector {
	return Collector{client: client, logger: logger}
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	metrics := []*prometheus.Desc{
		upDesc,
		serverInfoDesc,
		readyDesc,
		upTimeDesc,
		playerCountDesc,
		simulationCpuLoadDesc,
		simulationSpeedDesc,
		pcuUsedDesc,
		piratePcuUsedDesc,
		planetDesc,
		asteroidDesc,
		gridCountDesc,
		pcuCountDesc,
		bannedPlayersDesc,
		kickedPlayersDesc,
		cheatersDesc,
	}
	for i := range metrics {
		ch <- metrics[i]
	}
}

func (c Collector) SetUp(ch chan<- prometheus.Metric, up bool) {
	upVal := 0.0
	if up {
		upVal = 1
	}
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, upVal)
}

func (c Collector) CollectServerInfo(ch chan<- prometheus.Metric) error {
	serverInfo, err := c.client.GetServerDetails()
	if err != nil {
		return err
	}

	// space_engineers_info
	ch <- prometheus.MustNewConstMetric(
		serverInfoDesc,
		prometheus.GaugeValue,
		1,
		serverInfo.Data.ServerName,
		serverInfo.Data.WorldName,
		serverInfo.Data.Version,
		fmt.Sprintf("%v", serverInfo.Data.ServerId),
	)

	// space_engineers_ready
	var readyVal float64
	if serverInfo.Data.IsReady {
		readyVal = 1
	}
	ch <- prometheus.MustNewConstMetric(
		readyDesc,
		prometheus.GaugeValue,
		readyVal,
	)

	// space_engineers_up_seconds
	ch <- prometheus.MustNewConstMetric(
		upTimeDesc,
		prometheus.CounterValue,
		float64(serverInfo.Data.TotalTime),
	)

	// space_engineers_player_count
	ch <- prometheus.MustNewConstMetric(
		playerCountDesc,
		prometheus.GaugeValue,
		float64(serverInfo.Data.Players),
	)

	// space_engineers_simulation_cpu_load_percent
	ch <- prometheus.MustNewConstMetric(
		simulationCpuLoadDesc,
		prometheus.GaugeValue,
		serverInfo.Data.SimulationCpuLoad,
	)

	// space_engineers_simulation_speed
	ch <- prometheus.MustNewConstMetric(
		simulationSpeedDesc,
		prometheus.GaugeValue,
		serverInfo.Data.SimSpeed,
	)

	// space_engineers_pcu_used_total
	ch <- prometheus.MustNewConstMetric(
		pcuUsedDesc,
		prometheus.GaugeValue,
		float64(serverInfo.Data.UsedPCU),
	)

	// space_engineers_pirate_pcu_used_total
	ch <- prometheus.MustNewConstMetric(
		piratePcuUsedDesc,
		prometheus.GaugeValue,
		float64(serverInfo.Data.PirateUsedPCU),
	)

	return nil
}

func (c Collector) CollectPlanets(ch chan<- prometheus.Metric) error {
	resp, err := c.client.GetPlanets()
	if err != nil {
		return err
	}

	for i := range resp.Data.Planets {
		planet := &resp.Data.Planets[i]
		ch <- prometheus.MustNewConstMetric(
			planetDesc,
			prometheus.GaugeValue,
			1,
			planet.DisplayName,
			fmt.Sprintf("%v", planet.EntityId),
			fmt.Sprintf("%v", planet.Position.X),
			fmt.Sprintf("%v", planet.Position.Y),
			fmt.Sprintf("%v", planet.Position.Z),
		)
	}

	return nil
}

func (c Collector) CollectAsteroids(ch chan<- prometheus.Metric) error {
	resp, err := c.client.GetAsteroids()
	if err != nil {
		return err
	}

	for i := range resp.Data.Asteroids {
		asteroid := &resp.Data.Asteroids[i]
		ch <- prometheus.MustNewConstMetric(
			asteroidDesc,
			prometheus.GaugeValue,
			1,
			asteroid.DisplayName,
			fmt.Sprintf("%v", asteroid.EntityId),
			fmt.Sprintf("%v", asteroid.Position.X),
			fmt.Sprintf("%v", asteroid.Position.Y),
			fmt.Sprintf("%v", asteroid.Position.Z),
		)
	}

	return nil
}

func (c Collector) CollectGrids(ch chan<- prometheus.Metric) error {
	resp, err := c.client.GetGrids()
	if err != nil {
		return err
	}

	owners := make(map[string]bool)
	for i := range resp.Data.Grids {
		grid := &resp.Data.Grids[i]
		if ok := owners[grid.OwnerDisplayName]; !ok {
			owners[grid.OwnerDisplayName] = true
		}
	}

	for _, powered := range []bool{true, false} {
		for _, size := range []string{"Large", "Small"} {
			count := 0
			for i := range resp.Data.Grids {
				grid := &resp.Data.Grids[i]
				if grid.IsPowered == powered && grid.GridSize == size {
					count++
				}
			}
			ch <- prometheus.MustNewConstMetric(
				gridCountDesc,
				prometheus.GaugeValue,
				float64(count),
				fmt.Sprintf("%v", powered),
				size,
			)

			for owner := range owners {
				count = 0
				for i := range resp.Data.Grids {
					grid := &resp.Data.Grids[i]
					if grid.IsPowered == powered && grid.GridSize == size && grid.OwnerDisplayName == owner {
						count += int(grid.PCU)
					}
				}
				ch <- prometheus.MustNewConstMetric(
					pcuCountDesc,
					prometheus.GaugeValue,
					float64(count),
					fmt.Sprintf("%v", powered),
					size,
					owner,
				)
			}
		}
	}

	return nil
}

func (c Collector) CollectPlayers(ch chan<- prometheus.Metric) error {
	banned, err := c.client.GetBannedPlayers()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		bannedPlayersDesc,
		prometheus.GaugeValue,
		float64(len(banned.Data.BannedPlayers)),
	)

	kicked, err := c.client.GetKickedPlayers()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		kickedPlayersDesc,
		prometheus.GaugeValue,
		float64(len(kicked.Data.KickedPlayers)),
	)

	cheaters, err := c.client.GetCheaters()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		cheatersDesc,
		prometheus.GaugeValue,
		float64(len(cheaters.Data.Cheaters)),
	)

	return nil
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	ping, err := c.client.Ping()
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to ping remote API", "err", err)
		ping = false
	}

	c.SetUp(ch, ping)

	// Bail out now if the API is not up.
	if !ping {
		return
	}

	if err = c.CollectServerInfo(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect server info", "err", err)
		return
	}

	if err = c.CollectPlanets(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect planet info", "err", err)
		return
	}

	if err = c.CollectAsteroids(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect asteroid info", "err", err)
		return
	}

	if err = c.CollectGrids(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect grid info", "err", err)
		return
	}

	if err = c.CollectPlayers(ch); err != nil {
		level.Error(c.logger).Log("msg", "Failed to collect banned player count", "err", err)
		return
	}
}
