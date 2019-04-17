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
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/grobie/gomemcache/memcache"
	"github.com/prometheus/client_golang/prometheus"
)

func TestAcceptance(t *testing.T) {
	errc := make(chan error)

	addr := "localhost:11211"
	// MEMCACHED_PORT might be set by a linked memcached docker container.
	if env := os.Getenv("MEMCACHED_PORT"); env != "" {
		addr = strings.TrimPrefix(env, "tcp://")
	}

	ctx, cancel := context.WithCancel(context.Background())
	exporter := exec.CommandContext(ctx, "./memcached_exporter", "--memcached.address", addr)
	go func() {
		defer close(errc)

		if err := exporter.Run(); err != nil && errc != nil {
			errc <- err
		}
	}()

	defer cancel()

	// Wait for the exporter to be up and running.
OUTER:
	for {
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-timer.C:
			resp, err := http.Get("http://localhost:9150/")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					break OUTER
				}
			}
		case err := <-errc:
			t.Fatal("error running the exporter:", err)
		}
	}

	client, err := memcache.New(addr)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.StatsReset(); err != nil {
		t.Fatal(err)
	}

	item := &memcache.Item{Key: "foo", Value: []byte("bar")}
	if err := client.Set(item); err != nil {
		t.Fatal(err)
	}
	if err := client.Set(item); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get("foo"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get("qux"); err != memcache.ErrCacheMiss {
		t.Fatal(err)
	}
	last, err := client.Get("foo")
	if err != nil {
		t.Fatal(err)
	}
	last.Value = []byte("banana")
	if err := client.CompareAndSwap(last); err != nil {
		t.Fatal(err)
	}
	large := &memcache.Item{Key: "large", Value: bytes.Repeat([]byte("."), 130)}
	if err := client.Set(large); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get("http://localhost:9150/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		// memcached_current_connections varies depending on memcached versions
		// so it isn't practical to check for an exact value.
		`memcached_current_connections `,
		`memcached_up 1`,
		`memcached_commands_total{command="get",status="hit"} 2`,
		`memcached_commands_total{command="get",status="miss"} 1`,
		`memcached_commands_total{command="set",status="hit"} 3`,
		`memcached_commands_total{command="cas",status="hit"} 1`,
		`memcached_current_bytes 262`,
		`memcached_max_connections 1024`,
		`memcached_current_items 2`,
		`memcached_items_total 4`,
		`memcached_slab_current_items{slab="1"} 1`,
		`memcached_slab_current_items{slab="5"} 1`,
		`memcached_slab_commands_total{command="set",slab="1",status="hit"} 2`,
		`memcached_slab_commands_total{command="cas",slab="1",status="hit"} 1`,
		`memcached_slab_commands_total{command="set",slab="5",status="hit"} 1`,
		`memcached_slab_commands_total{command="cas",slab="5",status="hit"} 0`,
		`memcached_slab_current_chunks{slab="1"} 10922`,
		`memcached_slab_current_chunks{slab="5"} 4369`,
		`memcached_slab_mem_requested_bytes{slab="1"} 68`,
		`memcached_slab_mem_requested_bytes{slab="5"} 194`,
	}
	for _, test := range tests {
		if !bytes.Contains(body, []byte(test)) {
			t.Errorf("want metrics to include %q, have:\n%s", test, body)
		}
	}

	cancel()

	<-errc
}

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
			},
		}
		ch := make(chan prometheus.Metric, 100)
		e := NewExporter("", 100*time.Millisecond)
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
			},
		}
		ch := make(chan prometheus.Metric, 100)
		e := NewExporter("", 100*time.Millisecond)
		if err := e.parseStatsSettings(ch, statsSettings); err == nil {
			t.Error("expect return error but not")
		}
	})
}
