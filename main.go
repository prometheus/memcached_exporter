package main

import (
	"flag"
	"net/http"
	"strconv"
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

// Exporter collects metrics from a set of memcached servers.
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

// NewExporter returns an initialized exporter
func NewExporter(server string, timeout time.Duration) *Exporter {
	c := memcache.New(server)
	c.Timeout = timeout

	return &Exporter{
		mc: c,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the memcached server be reached.",
			nil,
			prometheus.Labels{"server": server},
		),
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime"),
			"The uptime of the server.",
			nil,
			prometheus.Labels{"server": server},
		),
		version: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "version"),
			"The version of this memcached server.",
			[]string{"version"},
			prometheus.Labels{"server": server},
		),
		bytesRead: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "read_bytes_total"),
			"Total number of bytes read by this server from network.",
			nil,
			prometheus.Labels{"server": server},
		),
		bytesWritten: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "written_bytes_total"),
			"Total number of bytes sent by this server to network.",
			nil,
			prometheus.Labels{"server": server},
		),
		connections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_connections"),
			"Current number of open connections.",
			nil,
			prometheus.Labels{"server": server},
		),
		connectionsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connections_total"),
			"Total number of connections opened since the server started running.",
			nil,
			prometheus.Labels{"server": server},
		),
		currentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_bytes"),
			"Current number of bytes used to store items.",
			nil,
			prometheus.Labels{"server": server},
		),
		limitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "limit_bytes"),
			"Number of bytes this server is allowed to use for storage.",
			nil,
			prometheus.Labels{"server": server},
		),
		commands: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "commands_total"),
			"The cache hits/misses asdf broken down by command (get, set, etc.).",
			[]string{"command", "status"},
			prometheus.Labels{"server": server},
		),
		items: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_items"),
			"Current number of items stored by this instance.",
			nil,
			prometheus.Labels{"server": server},
		),
		itemsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_total"),
			"Total number of items stored during the life of this instance.",
			nil,
			prometheus.Labels{"server": server},
		),
		evictions: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_evicted_total"),
			"Number of valid items removed from cache to free memory for new items.",
			nil,
			prometheus.Labels{"server": server},
		),
		reclaimed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_reclaimed_total"),
			"Number of times an entry was stored using memory from an expired entry.",
			nil,
			prometheus.Labels{"server": server},
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

// Collect fetches the statistics from the configured memcached servers, and
// delivers them as prometheus metrics. It implements prometheus.Collector.
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
		timeout       = flag.Duration("memcached.timeout", time.Second, "memcached connect timeout.")
		listenAddress = flag.String("web.listen-address", ":9106", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	servers := flag.Args()
	if len(servers) == 0 {
		servers = []string{"localhost:11211"}
	}
	for _, s := range servers {
		prometheus.MustRegister(NewExporter(s, *timeout))
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
