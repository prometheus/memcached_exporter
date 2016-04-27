package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/Snapbug/gomemcache/memcache"
)

func TestAcceptance(t *testing.T) {
	done := false

	// TODO(ts): Select unused port.
	server := exec.Command("memcached")
	go func() {
		if err := server.Run(); err != nil && !done {
			t.Fatal(err)
		}
	}()
	defer func() {
		if server.Process != nil {
			server.Process.Kill()
		}
	}()

	// TODO(ts): Select unused port and set memcached port.
	exporter := exec.Command("./memcached_exporter")
	go func() {
		if err := exporter.Run(); err != nil && !done {
			t.Fatal(err)
		}
	}()
	defer func() {
		if exporter.Process != nil {
			exporter.Process.Kill()
		}
	}()

	defer func() {
		done = true
	}()

	// TODO(ts): Replace sleep with ready check loop.
	<-time.After(100 * time.Millisecond)

	client := memcache.New("localhost:11211")
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

	resp, err := http.Get("http://localhost:9106/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		`memcached_up 1`,
		`memcached_commands_total{command="get",status="hit"} 2`,
		`memcached_commands_total{command="get",status="miss"} 1`,
		`memcached_commands_total{command="set",status="hit"} 2`,
		`memcached_commands_total{command="cas",status="hit"} 1`,
		`memcached_current_bytes 74`,
		`memcached_current_connections 11`,
		`memcached_current_items 1`,
		`memcached_items_total 3`,
	}
	for _, test := range tests {
		if !bytes.Contains(body, []byte(test)) {
			t.Errorf("want metrics to include %q, have:\n%s", test, body)
		}
	}
}
