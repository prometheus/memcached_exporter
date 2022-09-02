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

package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/memcached_exporter/pkg/exporter"
)

type Scraper struct {
	logger log.Logger

	totalScrapes prometheus.Counter
	scrapeErrors prometheus.Counter
}

func New(logger log.Logger) *Scraper {
	logger.Log("Creating new scraper")
	return &Scraper{
		logger: logger,
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "exporter_scrapes_total",
			Help: "Current total redis scrapes.",
		}),
		scrapeErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "exporter_scrapes_total",
			Help: "Current total redis scrapes.",
		}),
	}
}

func (s *Scraper) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		s.logger.Log("Calling handler with target: %s", target)
		s.totalScrapes.Inc()

		if target == "" {
			http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
			s.scrapeErrors.Inc()
			return
		}

		u, err := url.Parse(target)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid 'target' parameter, parse err: %ck ", err), http.StatusBadRequest)
			s.scrapeErrors.Inc()
			return
		}

		e := exporter.New(u.String(), 30*time.Second, s.logger)
		registry := prometheus.NewRegistry()
		registry.MustRegister(e)

		promhttp.HandlerFor(
			registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
		).ServeHTTP(w, r)
	}
}
