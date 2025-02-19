// Copyright 2022 Metrika Inc.
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

package buf

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	buckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

	bufferInsertDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "agent_buffer_insert_duration_seconds",
		Help:    "Histogram of buffer insert() duration in seconds",
		Buckets: buckets,
	})

	bufferGetDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "agent_buffer_get_duration_seconds",
		Help:    "Histogram of buffer get() duration in seconds",
		Buckets: buckets,
	})

	bufferDrainDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "agent_buffer_drain_duration_seconds",
		Help:    "Histogram of buffer drain() duration in seconds",
		Buckets: buckets,
	})

	// MetricsDropCnt tracked metrics dropped by agent
	MetricsDropCnt = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agent_metrics_drop_total_count", Help: "The total number of metrics dropped",
	}, []string{"reason"})
)
