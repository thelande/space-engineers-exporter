package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	vrage_client "github.com/thelande/space-engineers-exporter/pkg/vrage_client"
)

const namespace = "space_engineers"

var serverInfoLabels = []string{"server_name", "world_name", "version", "server_id"}

func getSEDesc(name string, unit string, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, name, unit), help, nil, nil,
	)
}

var (
	serverInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "info"),
		"Information about the server.",
		serverInfoLabels,
		nil,
	)

	upDesc                = getSEDesc("", "up", "Did the remote API respond to the queries.")
	readyDesc             = getSEDesc("", "ready", "Is the server ready?")
	upTimeDesc            = getSEDesc("up", "seconds", "The number of seconds that the server has been up.")
	playerCountDesc       = getSEDesc("player", "count", "The number of players currently connected to the server.")
	simulationCpuLoadDesc = getSEDesc("simulation_cpu_load", "percent", "The simulation thread CPU load.")
	simulationSpeedDesc   = getSEDesc("", "simulation_speed", "The simulation speed factor.")
	pcuUsedDesc           = getSEDesc("pcu_used", "total", "The total number of PCU used.")
	piratePcuUsedDesc     = getSEDesc("pirate_pcu_used", "total", "The total number of PCU used by pirate factions.")
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

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	ping, err := c.client.Ping()
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to ping remote API", "err", err)
		return
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
}
