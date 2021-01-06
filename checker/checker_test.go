// Copyright 2021 RetailNext, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package checker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func staticHandler(statusCode int, body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write(body)
	}
}

func makeTestServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/500", staticHandler(500, []byte("500\n")))
	mux.Handle("/200", staticHandler(200, []byte("OK\n")))
	return mux
}

type checkerTestCase struct {
	URL            string
	Timeout        time.Duration
	ExpectError    error
	ExpectAnyError bool
}

func (tc checkerTestCase) ctx() (context.Context, context.CancelFunc) {
	if tc.Timeout == 0 {
		return context.WithTimeout(context.Background(), 1*time.Second)
	}
	return context.WithTimeout(context.Background(), tc.Timeout)
}

func (tc checkerTestCase) checkError(t *testing.T, actual error) {
	t.Helper()
	if actual == nil {
		if tc.ExpectError != nil {
			t.Fatal("expected error but got nil")
		}
		return
	}
	if tc.ExpectAnyError {
		return
	}
	if errors.Is(actual, tc.ExpectError) {
		return
	}
	t.Fatalf("unexpected error: %+v", actual)
}

func TestChecker(t *testing.T) {
	sm := makeTestServeMux()
	plainServer := httptest.NewServer(sm)
	tlsServer := httptest.NewTLSServer(sm)

	t.Cleanup(func() {
		plainServer.Close()
		tlsServer.Close()
	})

	cases := map[string]checkerTestCase{
		"ConnectionError": {
			URL:            "https://127.255.255.255:255/../",
			Timeout:        1 * time.Millisecond,
			ExpectAnyError: true,
		},
		"Plain200": {
			URL: plainServer.URL + "/200",
		},
		"Plain500": {
			URL:         plainServer.URL + "/500",
			ExpectError: non2xxStatusCode(500),
		},
		"TLS200": {
			URL: tlsServer.URL + "/200",
		},
		"TLS500": {
			URL:         tlsServer.URL + "/500",
			ExpectError: non2xxStatusCode(500),
		},
	}

	for name := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := cases[name].ctx()
			defer cancel()
			chk := newGetChecker(cases[name].URL)
			result := chk.doCheck(ctx)
			cases[name].checkError(t, result)
		})
	}
}
