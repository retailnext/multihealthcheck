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

package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/retailnext/multihealthcheck/checker"
)

func HealthCheck(mc *checker.MultiChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
		defer cancel()
		result := mc.Check(ctx)
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

		if result.Ok() {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("OK\n"))
		} else {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("DOWN\n"))
		}
	}
}
