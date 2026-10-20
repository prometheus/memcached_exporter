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
