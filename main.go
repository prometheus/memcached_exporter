package main

import (
	"flag"
	"strings"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"os"

	"github.com/Snapbug/gomemcache/memcache"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "memcache"
)

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

func NewExporter(mc *memcache.Client) *Exporter {
	return &Exporter{
		mc: mc,
		up: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "up",
				Namespace: namespace,
				Help:      "If the servers were up.",
			},
			[]string{"server"},
		),
		uptime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "uptime",
				Namespace: namespace,
				Help:      "The time the server has been up.",
			},
			[]string{"server"},
		),
		cache: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "cache",
				Namespace: namespace,
				Help:      "The cache operations broken down by command and result (hit or miss).",
			},
			[]string{"server", "command", "status"},
		),
		usage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "usage",
				Namespace: namespace,
				Help:      "Details the usage of the server, by time (current/total) and resource (items/connections).",
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
				Help:      "Removal statuses from the cache either expired/evicted and if they were touched.",
			},
			[]string{"server", "status", "fetched"},
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.cache.Describe(ch)
	e.usage.Describe(ch)
	e.bytes.Describe(ch)
	e.removals.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
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
		glog.Fatalf("Failed to collect stats from memcache: %s", err)
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

		for _, c := range []string{"get", "delete", "incr", "decr", "cas", "touch"} {
			for _, s := range []string{"hits", "misses"} {
				m, err := strconv.ParseUint(stats[server][fmt.Sprintf("%s_%s", c, s)], 10, 64)
				if err != nil {
					e.cache.WithLabelValues(server.String(), c, s).Set(0)
				} else {
					e.cache.WithLabelValues(server.String(), c, s).Set(float64(m))
				}
			}
		}

		for _, c := range []string{"current", "total"} {
			for _, s := range []string{"items", "connections"} {
				m, err := strconv.ParseUint(stats[server][fmt.Sprintf("%s_%s", c, s)], 10, 64)
				if err != nil {
					e.usage.WithLabelValues(server.String(), c, s).Set(0)
				} else {
					e.usage.WithLabelValues(server.String(), c, s).Set(float64(m))
				}
			}
		}

		for _, c := range []string{"read", "written"} {
			m, err := strconv.ParseUint(stats[server][fmt.Sprintf("bytes_%s", c)], 10, 64)
			if err != nil {
				e.bytes.WithLabelValues(server.String(), c).Set(0)
			} else {
				e.bytes.WithLabelValues(server.String(), c).Set(float64(m))
			}
		}

		for _, c := range []string{"expired", "evicted"} {
			m, err := strconv.ParseUint(stats[server][fmt.Sprintf("%s_unfetched", c)], 10, 64)
			if err != nil {
				e.removals.WithLabelValues(server.String(), c, "unfetched").Set(0)
			} else {
				e.removals.WithLabelValues(server.String(), c, "unfetched").Set(float64(m))
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
