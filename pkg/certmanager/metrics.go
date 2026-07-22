// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package certmanager

import (
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type (
	provisionPhase  string
	provisionResult string
)

const (
	provisionPhaseCreateOrder provisionPhase = "create_order"
	provisionPhasePollOrder   provisionPhase = "poll_order"
	provisionPhaseIssueCert   provisionPhase = "issue_cert"
	provisionPhaseDNSCheck    provisionPhase = "dns_check"

	provisionResultOK          provisionResult = "ok"
	provisionResultNotReady    provisionResult = "not_ready"
	provisionResultError       provisionResult = "error"
	provisionResultRateLimited provisionResult = "rate_limited"
	provisionResultDNSError    provisionResult = "dns_error"
)

type metrics struct {
	provisionSteps *prometheus.CounterVec
	acmeErrors     *prometheus.CounterVec
	acmeCooldown   prometheus.Gauge
	stepDuration   *prometheus.HistogramVec
}

func newMetrics(registerer prometheus.Registerer) *metrics {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	return &metrics{
		provisionSteps: registerCollector(
			registerer,
			prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Subsystem: "certmanager",
					Name:      "certificate_provision_steps_total",
					Help:      "Certificate provisioning steps by phase and result.",
				},
				[]string{"phase", "result"},
			),
		),
		acmeErrors: registerCollector(
			registerer,
			prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Subsystem: "certmanager",
					Name:      "certificate_acme_errors_total",
					Help:      "ACME errors by problem type.",
				},
				[]string{"problem_type"},
			),
		),
		acmeCooldown: registerCollector(
			registerer,
			prometheus.NewGauge(
				prometheus.GaugeOpts{
					Subsystem: "certmanager",
					Name:      "certificate_acme_cooldown",
					Help:      "1 while the ACME client is in a global rate-limit cooldown.",
				},
			),
		),
		stepDuration: registerCollector(
			registerer,
			prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Subsystem: "certmanager",
					Name:      "certificate_provision_step_duration_seconds",
					Help:      "Duration of certificate provisioning steps in seconds.",
				},
				[]string{"phase"},
			),
		),
	}
}

func registerCollector[T prometheus.Collector](
	registerer prometheus.Registerer,
	collector T,
) T {
	if err := registerer.Register(collector); err != nil {
		if already, ok := errors.AsType[prometheus.AlreadyRegisteredError](err); ok {
			if existing, ok := already.ExistingCollector.(T); ok {
				return existing
			}
		}

		panic(err)
	}

	return collector
}

func (m *metrics) observeStep(phase provisionPhase, result provisionResult, started time.Time) {
	m.provisionSteps.WithLabelValues(string(phase), string(result)).Inc()
	m.stepDuration.WithLabelValues(string(phase)).Observe(time.Since(started).Seconds())
}

func (m *metrics) recordACMEError(problemType string) {
	if problemType == "" {
		problemType = "unknown"
	}

	m.acmeErrors.WithLabelValues(problemType).Inc()
}

func (m *metrics) setCooldown(active bool) {
	if active {
		m.acmeCooldown.Set(1)
		return
	}

	m.acmeCooldown.Set(0)
}
