// Copyright 2020 The Prometheus Authors
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

package exporter

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/grobie/gomemcache/memcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func TestParseStatsSettings(t *testing.T) {
	addr, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		var statsSettings = map[net.Addr]map[string]string{
			addr: {
				"maxconns":              "10",
				"lru_crawler":           "yes",
				"lru_crawler_sleep":     "100",
				"lru_crawler_tocrawl":   "0",
				"lru_maintainer_thread": "no",
				"hot_lru_pct":           "20",
				"warm_lru_pct":          "40",
				"hot_max_factor":        "0.20",
				"warm_max_factor":       "2.00",
				"accepting_conns":       "1",
			},
		}
		ch := make(chan prometheus.Metric, 100)
		e := New("", 100*time.Millisecond, promslog.NewNopLogger(), nil, true)
		if err := e.parseStatsSettings(ch, statsSettings); err != nil {
			t.Errorf("expect return error, error: %v", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Parallel()
		var statsSettings = map[net.Addr]map[string]string{
			addr: {
				"maxconns":              "10",
				"lru_crawler":           "yes",
				"lru_crawler_sleep":     "100",
				"lru_crawler_tocrawl":   "0",
				"lru_maintainer_thread": "fail",
				"hot_lru_pct":           "20",
				"warm_lru_pct":          "40",
				"hot_max_factor":        "0.20",
				"warm_max_factor":       "2.00",
				"accepting_conns":       "fail",
			},
		}
		ch := make(chan prometheus.Metric, 100)
		e := New("", 100*time.Millisecond, promslog.NewNopLogger(), nil, true)
		if err := e.parseStatsSettings(ch, statsSettings); err == nil {
			t.Error("expect return error but not")
		}
	})
}

func TestParseTimeval(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		_, err := parseTimeval(map[string]string{"rusage_system": "3.5"}, "rusage_system", promslog.NewNopLogger())
		if err != nil {
			t.Errorf("expect return error, error: %v", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Parallel()
		_, err := parseTimeval(map[string]string{"rusage_system": "35"}, "rusage_system", promslog.NewNopLogger())
		if err == nil {
			t.Error("expect return error but not")
		}
	})
}

func TestParseStatsSlabToggle(t *testing.T) {
	addr, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	stats := map[net.Addr]memcache.Stats{
		addr: {
			Stats: map[string]string{
				"version":    "1.6.0",
				"cmd_set":    "0",
				"cas_misses": "0",
				"cas_hits":   "0",
				"cas_badval": "0",
			},
			Slabs: map[int]map[string]string{
				1: {
					"chunk_size":      "96",
					"chunks_per_page": "10922",
					"total_pages":     "1",
					"total_chunks":    "10922",
					"used_chunks":     "1",
					"free_chunks":     "10921",
					"free_chunks_end": "0",
					"mem_requested":   "68",
					"cmd_set":         "2",
					"cas_hits":        "0",
					"cas_badval":      "0",
				},
			},
		},
	}

	countSlabMetrics := func(enableSlab bool) int {
		e := New("", 100*time.Millisecond, promslog.NewNopLogger(), nil, enableSlab)
		ch := make(chan prometheus.Metric, 100)
		if err := e.parseStats(ch, stats); err != nil {
			t.Fatalf("parseStats returned an error: %v", err)
		}
		close(ch)

		var slabMetrics int
		for m := range ch {
			if strings.Contains(m.Desc().String(), "memcached_slab_") {
				slabMetrics++
			}
		}
		return slabMetrics
	}

	if got := countSlabMetrics(true); got == 0 {
		t.Error("expected slab metrics to be exported when the slab collector is enabled, got none")
	}
	if got := countSlabMetrics(false); got != 0 {
		t.Errorf("expected no slab metrics when the slab collector is disabled, got %d", got)
	}
}

func TestParseStatsGlobalCounters(t *testing.T) {
	addr, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	newStats := func(extstore bool) map[net.Addr]memcache.Stats {
		s := map[string]string{
			"version":               "1.6.0",
			"cmd_set":               "0",
			"cas_misses":            "0",
			"cas_hits":              "0",
			"cas_badval":            "0",
			"get_expired":           "1",
			"get_flushed":           "2",
			"evicted_active":        "3",
			"evicted_unfetched":     "4",
			"expired_unfetched":     "5",
			"slabs_moved":           "6",
			"slab_reassign_running": "0",
			"get_extstore":          "7",
			"get_aborted_extstore":  "8",
			"recache_from_extstore": "9",
		}
		if extstore {
			s["extstore_limit_maxbytes"] = "1024"
		}
		return map[net.Addr]memcache.Stats{addr: {Stats: s}}
	}

	collect := func(extstore bool) string {
		e := New("", 100*time.Millisecond, promslog.NewNopLogger(), nil, false)
		ch := make(chan prometheus.Metric, 100)
		if err := e.parseStats(ch, newStats(extstore)); err != nil {
			t.Fatalf("parseStats returned an error: %v", err)
		}
		close(ch)

		var descs strings.Builder
		for m := range ch {
			descs.WriteString(m.Desc().String())
		}
		return descs.String()
	}

	has := func(descs, name string) bool {
		return strings.Contains(descs, `fqName: "`+name+`"`)
	}

	global := []string{
		"memcached_get_expired_total",
		"memcached_get_flushed_total",
		"memcached_items_evicted_active_total",
		"memcached_items_evicted_unfetched_total",
		"memcached_items_expired_unfetched_total",
		"memcached_slabs_moved_total",
		"memcached_slab_reassign_running",
	}
	extstore := []string{
		"memcached_get_extstore_total",
		"memcached_get_aborted_extstore_total",
		"memcached_recache_from_extstore_total",
	}

	without := collect(false)
	for _, name := range global {
		if !has(without, name) {
			t.Errorf("expected %s to be exported", name)
		}
	}
	for _, name := range extstore {
		if has(without, name) {
			t.Errorf("expected %s not to be exported without extstore", name)
		}
	}

	with := collect(true)
	for _, name := range extstore {
		if !has(with, name) {
			t.Errorf("expected %s to be exported with extstore", name)
		}
	}
}

func TestParseStatsRuntimeStats(t *testing.T) {
	addr, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	// Each entry maps a key of the "stats" output to the metric it is
	// exported as.
	runtimeStats := []struct {
		key    string
		metric string
		value  float64
	}{
		{"threads", "memcached_threads", 4},
		{"connection_structures", "memcached_connection_structures", 20},
		{"malloc_fails", "memcached_malloc_fails_total", 1},
		{"auth_cmds", "memcached_auth_commands_total", 2},
		{"auth_errors", "memcached_auth_errors_total", 3},
		{"hash_power_level", "memcached_hash_power_level", 16},
		{"hash_bytes", "memcached_hash_bytes", 524288},
		{"hash_is_expanding", "memcached_hash_expanding", 1},
		{"response_obj_count", "memcached_response_objects", 18},
		{"response_obj_bytes", "memcached_response_objects_bytes", 3456},
		{"response_obj_oom", "memcached_response_objects_oom_total", 4},
		{"read_buf_count", "memcached_read_buffers", 19},
		{"read_buf_bytes", "memcached_read_buffers_bytes", 65536},
		{"read_buf_bytes_free", "memcached_read_buffers_free_bytes", 1024},
		{"read_buf_oom", "memcached_read_buffers_oom_total", 5},
		{"lru_crawler_running", "memcached_lru_crawler_running", 2},
		{"lru_maintainer_juggles", "memcached_lru_maintainer_juggles_total", 6},
		{"lru_bumps_dropped", "memcached_lru_bumps_dropped_total", 7},
		{"slab_global_page_pool", "memcached_slab_global_page_pool", 21},
		{"slab_reassign_rescues", "memcached_slab_reassign_rescues_total", 8},
		{"slab_reassign_chunk_rescues", "memcached_slab_reassign_chunk_rescues_total", 9},
		{"slab_reassign_inline_reclaim", "memcached_slab_reassign_inline_reclaim_total", 10},
		{"slab_reassign_busy_items", "memcached_slab_reassign_busy_items_total", 11},
		{"slab_reassign_busy_deletes", "memcached_slab_reassign_busy_deletes_total", 12},
		{"slab_reassign_busy_nomem", "memcached_slab_reassign_busy_nomem_total", 13},
		{"log_worker_dropped", "memcached_log_worker_dropped_total", 14},
		{"log_worker_written", "memcached_log_worker_written_total", 15},
		{"log_watcher_skipped", "memcached_log_watcher_skipped_total", 16},
		{"log_watcher_sent", "memcached_log_watcher_sent_total", 17},
		{"log_watchers", "memcached_log_watchers", 1},
	}

	base := func() map[string]string {
		return map[string]string{
			"version":    "1.6.0",
			"cmd_set":    "0",
			"cas_misses": "0",
			"cas_hits":   "0",
			"cas_badval": "0",
		}
	}

	type sample struct {
		value   float64
		counter bool
	}

	collect := func(s map[string]string) (map[string]sample, error) {
		e := New("", 100*time.Millisecond, promslog.NewNopLogger(), nil, false)
		ch := make(chan prometheus.Metric, 200)
		err := e.parseStats(ch, map[net.Addr]memcache.Stats{addr: {Stats: s}})
		close(ch)

		reg := prometheus.NewPedanticRegistry()
		reg.MustRegister(metricsCollector(ch))
		families, gerr := reg.Gather()
		if gerr != nil {
			t.Fatalf("gathering metrics failed: %v", gerr)
		}

		got := map[string]sample{}
		for _, mf := range families {
			m := mf.GetMetric()[0]
			if c := m.GetCounter(); c != nil {
				got[mf.GetName()] = sample{c.GetValue(), true}
			} else {
				got[mf.GetName()] = sample{m.GetGauge().GetValue(), false}
			}
		}
		return got, err
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		s := base()
		for _, r := range runtimeStats {
			s[r.key] = strconv.FormatFloat(r.value, 'f', -1, 64)
		}
		got, err := collect(s)
		if err != nil {
			t.Fatalf("parseStats returned an error: %v", err)
		}
		for _, r := range runtimeStats {
			m, ok := got[r.metric]
			if !ok {
				t.Errorf("expected %s to be exported", r.metric)
				continue
			}
			if m.value != r.value {
				t.Errorf("%s: expected %v, got %v", r.metric, r.value, m.value)
			}
			if wantCounter := strings.HasSuffix(r.metric, "_total"); m.counter != wantCounter {
				t.Errorf("%s: expected counter=%v, got counter=%v", r.metric, wantCounter, m.counter)
			}
		}
	})

	t.Run("MissingKeys", func(t *testing.T) {
		t.Parallel()

		got, err := collect(base())
		if err != nil {
			t.Fatalf("parseStats returned an error: %v", err)
		}
		for _, r := range runtimeStats {
			if _, ok := got[r.metric]; ok {
				t.Errorf("did not expect %s to be exported without its stat", r.metric)
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Parallel()

		s := base()
		s["threads"] = "many"
		s["hash_power_level"] = "16"
		got, err := collect(s)
		if err == nil {
			t.Error("expect return error but not")
		}
		if _, ok := got["memcached_threads"]; ok {
			t.Error("did not expect memcached_threads to be exported for an unparsable value")
		}
		if _, ok := got["memcached_hash_power_level"]; !ok {
			t.Error("expected memcached_hash_power_level to still be exported")
		}
	})
}

// metricsCollector replays already collected metrics. It describes nothing, so
// the registry treats it as an unchecked collector.
type metricsCollector chan prometheus.Metric

func (c metricsCollector) Describe(chan<- *prometheus.Desc) {}

func (c metricsCollector) Collect(ch chan<- prometheus.Metric) {
	for m := range c {
		ch <- m
	}
}
