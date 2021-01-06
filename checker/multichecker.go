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
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MultiChecker struct {
	lock     sync.RWMutex
	checkers checkers
	metrics  metrics
}

func (m *MultiChecker) AddGetCheck(name, url string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metrics.setupOnce.Do(m.metrics.setup)

	rc := registeredChecker{
		Name:     name,
		Checker:  newGetChecker(url),
		Duration: m.metrics.checkDuration.WithLabelValues(name),
		Errors:   m.metrics.checkErrors.WithLabelValues(name),
	}
	m.checkers = append(m.checkers, rc)
	sort.Sort(m.checkers)
}

func (m *MultiChecker) Check(ctx context.Context) MultiResult {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.checkers.run(ctx)
}

func (m *MultiChecker) MustRegisterWith(registerer prometheus.Registerer) {
	m.metrics.setupOnce.Do(m.metrics.setup)
	registerer.MustRegister(m.metrics.checkDuration, m.metrics.checkErrors)
}

type checkers []registeredChecker

func (cc checkers) Len() int           { return len(cc) }
func (cc checkers) Swap(i, j int)      { cc[i], cc[j] = cc[j], cc[i] }
func (cc checkers) Less(i, j int) bool { return cc[i].Name < cc[j].Name }

func (cc checkers) run(ctx context.Context) MultiResult {
	var wg sync.WaitGroup
	mr := make(MultiResult, len(cc))

	for i, rc := range cc {
		wg.Add(1)
		go func(i int, rc registeredChecker) {
			defer wg.Done()
			t0 := time.Now()
			err := rc.Checker.doCheck(ctx)
			rc.Duration.Observe(time.Since(t0).Seconds())
			if err != nil {
				rc.Errors.Inc()
			}
			mr[i].Name = rc.Name
			mr[i].Error = err
		}(i, rc)
	}

	wg.Wait()
	return mr
}

type registeredChecker struct {
	Name     string
	Checker  checker
	Duration prometheus.Observer
	Errors   prometheus.Counter
}

type CheckResult struct {
	Name  string
	Error error
}

type MultiResult []CheckResult

func (r MultiResult) Ok() bool {
	for _, v := range r {
		if v.Error != nil {
			return false
		}
	}
	return true
}
