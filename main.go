package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Snapbug/gomemcache/memcache"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "memcache"
)

var (
	cacheOperations  = []string{"get", "delete", "incr", "decr", "cas", "touch"}
	cacheStatuses    = []string{"hits", "misses"}
	usageTimes       = []string{"current", "total"}
	usageResources   = []string{"items", "connections"}
	bytesDirections  = []string{"read", "written"}
	removalsStatuses = []string{"expired", "evicted"}
)

// Exporter collects metrics from a set of memcache servers.
type Exporter struct {
	mutex    sync.RWMutex
	mc       *memcache.Client
	up       *prometheus.GaugeVec
	uptime   *prometheus.CounterVec
	cache    *prometheus.CounterVec
	usage    *prometheus.GaugeVec
	bytes    *prometheus.CounterVec
	removals *prometheus.CounterVec
}

// NewExporter returns an initialized exporter
func NewExporter(mc *memcache.Client) *Exporter {
	return &Exporter{
		mc: mc,
		up: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "up",
				Namespace: namespace,
				Help:      "Are the servers up.",
			},
			[]string{"server"},
		),
		uptime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "uptime",
				Namespace: namespace,
				Help:      "The uptime of the server.",
			},
			[]string{"server"},
		),
		cache: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "cache",
				Namespace: namespace,
				Help:      "The cache hits/misses broken down by command (get, set, etc.).",
			},
			[]string{"server", "command", "status"},
		),
		usage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "usage",
				Namespace: namespace,
				Help:      "Details the resource usage (items/connections) of the server, by time (current/total).",
			},
			[]string{"server", "time", "resource"},
		),
		bytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "bytes",
				Namespace: namespace,
				Help:      "The bytes sent/received by the server.",
			},
			[]string{"server", "direction"},
		),
		removals: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "removal",
				Namespace: namespace,
				Help:      "Number of items that have been evicted/expired (status), and if the were fetched ever or not.",
			},
			[]string{"server", "status", "fetched"},
		),
	}
}

// Describe describes all the metrics exported by the memcache exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.cache.Describe(ch)
	e.usage.Describe(ch)
	e.bytes.Describe(ch)
	e.removals.Describe(ch)
}

// Collect fetches the statistics from the configured memcache servers, and
// delivers them as prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// prevent concurrent metric collections
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.up.Reset()
	e.uptime.Reset()
	e.cache.Reset()
	e.usage.Reset()
	e.bytes.Reset()
	e.removals.Reset()

	stats, err := e.mc.Stats()

	if err != nil {
		glog.Infof("Failed to collect stats from memcache: %s", err)
		return
	}

	for server, _ := range stats {
		e.up.WithLabelValues(server.String()).Set(1)

		m, err := strconv.ParseUint(stats[server]["uptime"], 10, 64)
		if err != nil {
			e.removals.WithLabelValues(server.String()).Set(0)
		} else {
			e.removals.WithLabelValues(server.String()).Set(float64(m))
		}

		for _, op := range cacheOperations {
			for _, st := range cacheStatuses {
				m, err := strconv.ParseUint(stats[server][fmt.Sprintf("%s_%s", op, st)], 10, 64)
				if err != nil {
					e.cache.WithLabelValues(server.String(), op, st).Set(0)
				} else {
					e.cache.WithLabelValues(server.String(), op, st).Set(float64(m))
				}
			}
		}

		for _, t := range usageTimes {
			for _, r := range usageResources {
				m, err := strconv.ParseUint(stats[server][fmt.Sprintf("%s_%s", t, r)], 10, 64)
				if err != nil {
					e.usage.WithLabelValues(server.String(), t, r).Set(0)
				} else {
					e.usage.WithLabelValues(server.String(), t, r).Set(float64(m))
				}
			}
		}

		for _, dir := range bytesDirections {
			m, err := strconv.ParseUint(stats[server][fmt.Sprintf("bytes_%s", dir)], 10, 64)
			if err != nil {
				e.bytes.WithLabelValues(server.String(), dir).Set(0)
			} else {
				e.bytes.WithLabelValues(server.String(), dir).Set(float64(m))
			}
		}

		for _, st := range removalsStatuses {
			m, err := strconv.ParseUint(stats[server][fmt.Sprintf("%s_unfetched", st)], 10, 64)
			if err != nil {
				e.removals.WithLabelValues(server.String(), st, "unfetched").Set(0)
			} else {
				e.removals.WithLabelValues(server.String(), st, "unfetched").Set(float64(m))
			}
		}
		m, err = strconv.ParseUint(stats[server]["evictions"], 10, 64)
		if err != nil {
			e.removals.WithLabelValues(server.String(), "evictions", "fetched").Set(0)
		} else {
			e.removals.WithLabelValues(server.String(), "evictions", "fetched").Set(float64(m))
		}
	}

	e.up.Collect(ch)
	e.uptime.Collect(ch)
	e.cache.Collect(ch)
	e.usage.Collect(ch)
	e.bytes.Collect(ch)
	e.removals.Collect(ch)
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9106", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	env := os.Getenv("memcache_servers")
	if env == "" {
		glog.Fatalf("No servers specified")
	}

	mc := memcache.New(strings.Split(env, ",")...)
	exporter := NewExporter(mc)

	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, prometheus.Handler())
	http.ListenAndServe(*listenAddress, nil)
}
