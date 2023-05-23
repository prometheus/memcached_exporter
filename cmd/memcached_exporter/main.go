// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/prometheus/memcached_exporter/pkg/exporter"
)

func main() {
	var (
		address            = kingpin.Flag("memcached.address", "Memcached server address.").Default("localhost:11211").String()
		timeout            = kingpin.Flag("memcached.timeout", "memcached connect timeout.").Default("1s").Duration()
		pidFile            = kingpin.Flag("memcached.pid-file", "Optional path to a file containing the memcached PID for additional metrics.").Default("").String()
		enableTLS          = kingpin.Flag("memcached.tls.enable", "Enable TLS connections to memcached").Bool()
		certFile           = kingpin.Flag("memcached.tls.cert-file", "Client certificate file.").Default("").String()
		keyFile            = kingpin.Flag("memcached.tls.key-file", "Client private key file.").Default("").String()
		caFile             = kingpin.Flag("memcached.tls.ca-file", "Client root CA file.").Default("").String()
		insecureSkipVerify = kingpin.Flag("memcached.tls.insecure-skip-verify", "Skip server certificate verification").Bool()
		serverName         = kingpin.Flag("memcached.tls.server-name", "Memcached TLS certificate servername").Default("").String()
		webConfig          = webflag.AddFlags(kingpin.CommandLine, ":9150")
		metricsPath        = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	)
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Print("memcached_exporter"))
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting memcached_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	var (
		tlsConfig *tls.Config
		err       error
	)
	if *enableTLS {
		if *serverName == "" {
			*serverName, _, err = net.SplitHostPort(*address)
			if err != nil {
				if strings.Contains(*address, "/") {
					level.Error(logger).Log("msg",
						"If --memcached.tls.enable is set and --memcached.address is a unix socket, "+
							"you must also specify --memcached.tls.server-name")
				} else {
					level.Error(logger).Log("msg", "Error parsing memcached address", "err", err)
				}
				os.Exit(1)
			}
		}
		tlsConfig, err = promconfig.NewTLSConfig(&promconfig.TLSConfig{
			CertFile:           *certFile,
			KeyFile:            *keyFile,
			CAFile:             *caFile,
			ServerName:         *serverName,
			InsecureSkipVerify: *insecureSkipVerify,
		})
		if err != nil {
			level.Error(logger).Log("msg", "Failed to create TLS config", "err", err)
			os.Exit(1)
		}
	}

	prometheus.MustRegister(version.NewCollector("memcached_exporter"))
	prometheus.MustRegister(exporter.New(*address, *timeout, logger, tlsConfig))

	if *pidFile != "" {
		procExporter := collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			PidFn:     prometheus.NewPidFileFn(*pidFile),
			Namespace: exporter.Namespace,
		})
		prometheus.MustRegister(procExporter)
	}

	http.Handle(*metricsPath, promhttp.Handler())
	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "memcached_exporter",
			Description: "Prometheus Exporter for Memcached servers",
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
		http.Handle("/", landingPage)
	}

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error running HTTP server", "err", err)
		os.Exit(1)
	}
}
