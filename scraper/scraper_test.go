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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/promslog"
)

func TestHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		s := New(1*time.Second, promslog.NewNopLogger(), nil)

		req, err := http.NewRequest("GET", "/?target=127.0.0.1:11211", nil)

		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.Handler())

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %d, want: %d. body: %s",
				status, http.StatusOK, rr.Body.String())
		}

		memcachedUpMetric := "memcached_up 1"

		if body := rr.Body.String(); !strings.Contains(body, memcachedUpMetric) {
			t.Errorf("handler could not inspect metrics. body: %s", body)
		}
	})

	t.Run("No target", func(t *testing.T) {
		t.Parallel()

		s := New(1*time.Second, promslog.NewNopLogger(), nil)

		req, err := http.NewRequest("GET", "/", nil)

		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.Handler())

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %d, want: %d", rr.Code, http.StatusBadRequest)
		}
	})
}
