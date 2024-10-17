// Copyright 2022 The Prometheus Authors
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

package scraper

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/memcached_exporter/pkg/exporter"
)

type Scraper struct {
	logger    *slog.Logger
	timeout   time.Duration
	tlsConfig *tls.Config

	scrapeCount  prometheus.Counter
	scrapeErrors prometheus.Counter
}

func New(timeout time.Duration, logger *slog.Logger, tlsConfig *tls.Config) *Scraper {
	logger.Debug("Started scrapper")
	return &Scraper{
		logger:    logger,
		timeout:   timeout,
		tlsConfig: tlsConfig,
		scrapeCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "memcached_exporter_scrapes_total",
			Help: "Count of memcached exporter scapes.",
		}),
		scrapeErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "memcached_exporter_scrape_errors_total",
			Help: "Count of memcached exporter scape errors.",
		}),
	}
}

func (s *Scraper) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		s.logger.Debug("scrapping memcached", "target", target)
		s.scrapeCount.Inc()

		if target == "" {
			errorStr := "'target' parameter must be specified"
			s.logger.Warn(errorStr)
			http.Error(w, errorStr, http.StatusBadRequest)
			s.scrapeErrors.Inc()
			return
		}

		e := exporter.New(target, s.timeout, s.logger, s.tlsConfig)
		registry := prometheus.NewRegistry()
		registry.Register(e)

		promhttp.HandlerFor(
			registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
		).ServeHTTP(w, r)
	}
}
