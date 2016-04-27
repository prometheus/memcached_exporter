package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Snapbug/gomemcache/memcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

const (
	namespace = "memcached"
)

var (
	Version = "0.0.0"
)

// Exporter collects metrics from a memcached server.
type Exporter struct {
	mc *memcache.Client

	up               *prometheus.Desc
	uptime           *prometheus.Desc
	version          *prometheus.Desc
	bytesRead        *prometheus.Desc
	bytesWritten     *prometheus.Desc
	connections      *prometheus.Desc
	connectionsTotal *prometheus.Desc
	currentBytes     *prometheus.Desc
	limitBytes       *prometheus.Desc
	commands         *prometheus.Desc
	items            *prometheus.Desc
	itemsTotal       *prometheus.Desc
	evictions        *prometheus.Desc
	reclaimed        *prometheus.Desc
}

// NewExporter returns an initialized exporter.
func NewExporter(server string, timeout time.Duration) *Exporter {
	c := memcache.New(server)
	c.Timeout = timeout

	return &Exporter{
		mc: c,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the memcached server be reached.",
			nil,
			nil,
		),
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime"),
			"The uptime of the server.",
			nil,
			nil,
		),
		version: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "version"),
			"The version of this memcached server.",
			[]string{"version"},
			nil,
		),
		bytesRead: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "read_bytes_total"),
			"Total number of bytes read by this server from network.",
			nil,
			nil,
		),
		bytesWritten: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "written_bytes_total"),
			"Total number of bytes sent by this server to network.",
			nil,
			nil,
		),
		connections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_connections"),
			"Current number of open connections.",
			nil,
			nil,
		),
		connectionsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connections_total"),
			"Total number of connections opened since the server started running.",
			nil,
			nil,
		),
		currentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_bytes"),
			"Current number of bytes used to store items.",
			nil,
			nil,
		),
		limitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "limit_bytes"),
			"Number of bytes this server is allowed to use for storage.",
			nil,
			nil,
		),
		commands: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "commands_total"),
			"The cache hits/misses asdf broken down by command (get, set, etc.).",
			[]string{"command", "status"},
			nil,
		),
		items: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_items"),
			"Current number of items stored by this instance.",
			nil,
			nil,
		),
		itemsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_total"),
			"Total number of items stored during the life of this instance.",
			nil,
			nil,
		),
		evictions: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_evicted_total"),
			"Number of valid items removed from cache to free memory for new items.",
			nil,
			nil,
		),
		reclaimed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_reclaimed_total"),
			"Number of times an entry was stored using memory from an expired entry.",
			nil,
			nil,
		),
	}
}

// Describe describes all the metrics exported by the memcached exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.uptime
	ch <- e.version
	ch <- e.bytesRead
	ch <- e.bytesWritten
	ch <- e.connections
	ch <- e.connectionsTotal
	ch <- e.currentBytes
	ch <- e.limitBytes
	ch <- e.commands
	ch <- e.items
	ch <- e.itemsTotal
	ch <- e.evictions
	ch <- e.reclaimed
}

// Collect fetches the statistics from the configured memcached server, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	stats, err := e.mc.Stats()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		log.Errorf("Failed to collect stats from memcached: %s", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	for _, s := range stats {
		ch <- prometheus.MustNewConstMetric(e.uptime, prometheus.CounterValue, parse(s, "uptime"))
		ch <- prometheus.MustNewConstMetric(e.version, prometheus.GaugeValue, 1, s["version"])

		for _, op := range []string{"get", "delete", "incr", "decr", "cas", "touch"} {
			ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, parse(s, op+"_hits"), op, "hit")
			ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, parse(s, op+"_misses"), op, "miss")
		}
		ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, parse(s, "cas_badval"), "cas", "badval")
		ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, parse(s, "cmd_flush"), "flush", "hit")

		// memcached includes cas operations again in cmd_set.
		set := 0.
		if setCmd, err := strconv.ParseFloat(s["cmd_set"], 64); err == nil {
			if cas, casErr := sum(s, "cas_misses", "cas_hits", "cas_badval"); casErr == nil {
				set = setCmd - cas
			} else {
				log.Errorf("Failed to parse cas: %s", casErr)
			}
		} else {
			log.Errorf("Failed to parse set %q: %s", s["cmd_set"], err)
		}
		ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, set, "set", "hit")

		ch <- prometheus.MustNewConstMetric(e.currentBytes, prometheus.GaugeValue, parse(s, "bytes"))
		ch <- prometheus.MustNewConstMetric(e.limitBytes, prometheus.GaugeValue, parse(s, "limit_maxbytes"))
		ch <- prometheus.MustNewConstMetric(e.items, prometheus.GaugeValue, parse(s, "curr_items"))
		ch <- prometheus.MustNewConstMetric(e.itemsTotal, prometheus.CounterValue, parse(s, "total_items"))

		ch <- prometheus.MustNewConstMetric(e.bytesRead, prometheus.CounterValue, parse(s, "bytes_read"))
		ch <- prometheus.MustNewConstMetric(e.bytesWritten, prometheus.CounterValue, parse(s, "bytes_written"))

		ch <- prometheus.MustNewConstMetric(e.connections, prometheus.GaugeValue, parse(s, "curr_connections"))
		ch <- prometheus.MustNewConstMetric(e.connectionsTotal, prometheus.CounterValue, parse(s, "total_connections"))

		ch <- prometheus.MustNewConstMetric(e.evictions, prometheus.CounterValue, parse(s, "evictions"))
		ch <- prometheus.MustNewConstMetric(e.reclaimed, prometheus.CounterValue, parse(s, "reclaimed"))
	}

}

func parse(stats map[string]string, key string) float64 {
	v, err := strconv.ParseFloat(stats[key], 64)
	if err != nil {
		log.Errorf("Failed to parse %s %q: %s", key, stats[key], err)
		v = 0
	}
	return v
}

func sum(stats map[string]string, keys ...string) (float64, error) {
	s := 0.
	for _, key := range keys {
		v, err := strconv.ParseFloat(stats[key], 64)
		if err != nil {
			return 0, err
		}
		s += v
	}
	return s, nil
}

func main() {
	var (
		address       = flag.String("memcached.address", "localhost:11211", "Memcached server address.")
		timeout       = flag.Duration("memcached.timeout", time.Second, "memcached connect timeout.")
		pidFile       = flag.String("memcached.pid-file", "", "Optional path to a file containing the memcached PID for additional metrics.")
		listenAddress = flag.String("web.listen-address", ":9106", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	prometheus.MustRegister(NewExporter(*address, *timeout))

	if *pidFile != "" {
		procExporter := prometheus.NewProcessCollectorPIDFn(
			func() (int, error) {
				content, err := ioutil.ReadFile(*pidFile)
				if err != nil {
					return 0, fmt.Errorf("Can't read pid file %q: %s", *pidFile, err)
				}
				value, err := strconv.Atoi(strings.TrimSpace(string(content)))
				if err != nil {
					return 0, fmt.Errorf("Can't parse pid file %q: %s", *pidFile, err)
				}
				return value, nil
			}, namespace)
		prometheus.MustRegister(procExporter)
	}

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Memcached Exporter</title></head>
             <body>
             <h1>Memcached Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Infof("Starting memcached_exporter v%s at %s", Version, *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
