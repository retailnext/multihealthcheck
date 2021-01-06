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
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	setupOnce     sync.Once
	checkDuration *prometheus.HistogramVec
	checkErrors   *prometheus.CounterVec
}

func (m *metrics) setup() {
	m.checkDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "health_check_duration_seconds",
		Help: "Health check latency.",
	}, []string{"check"})

	m.checkErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "health_check_errors_total",
		Help: "Health check errors.",
	}, []string{"check"})
}
