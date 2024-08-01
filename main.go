/*
Copyright 2023 Thomas Helander

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/thelande/space-engineers-exporter/pkg/collector"
	vrage_client "github.com/thelande/space-engineers-exporter/pkg/vrage_client"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

const (
	exporterName  = "space_engineers_exporter"
	exporterTitle = "Space Engineers Dedicated Server Exporter"
)

var (
	api = kingpin.Flag(
		"remote-api.url",
		"URL of the remote API",
	).Default("http://127.0.0.1:8080").String()

	key = kingpin.Flag(
		"remote-api.key",
		"The secret key used to communicate with the remote API.",
	).Envar("SE_REMOTE_API_KEY").String()

	keyFile = kingpin.Flag(
		"remote-api.key-file",
		"Path of the file containing the remote API secret key.",
	).Envar("SE_REMOTE_API_KEY_FILE").String()

	sslVerify = kingpin.Flag(
		"remote-api.ssl-verify",
		"Verify the remote API SSL certificate, when true.",
	).Default("true").Bool()

	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	webConfig = webflag.AddFlags(kingpin.CommandLine, ":9815")
	logger    log.Logger
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Print(exporterName))
	kingpin.Parse()

	logger = promlog.New(promlogConfig)
	level.Info(logger).Log("msg", fmt.Sprintf("Starting %s", exporterName), "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	client, err := vrage_client.NewVRageClient(*api, *keyFile, *key, *sslVerify, &logger)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			level.Error(logger).Log("msg", "Secret key file not found", "path", *keyFile)
		} else {
			level.Error(logger).Log("msg", "Unknown error occurred while creating VRage client", "err", err)
		}
		os.Exit(1)
	}

	collector := collector.NewCollector(client, logger)

	// Uncomment the following two lines and comment out prometheus.MustRegister(collector)
	// to exclude the go metrics. Make sure to swap line 88 and 89 as well.
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)
	// prometheus.MustRegister(collector)

	landingConfig := web.LandingConfig{
		Name:        exporterTitle,
		Description: "Prometheus Space Engineers Dedicated Server Exporter",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{
				Address: *metricsPath,
				Text:    "Metrics",
			},
		},
	}
	landingPage, err := web.NewLandingPage(landingConfig)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	// http.Handle(*metricsPath, promhttp.Handler())
	http.Handle("/", landingPage)

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "HTTP listener stopped", "error", err)
		os.Exit(1)
	}
}
