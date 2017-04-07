package main

import (
	"encoding/json"
	"sync/atomic"
)

// Centrally keeps track of request metrics
// Notice we use signed integers to prevent extra casting when avoiding underflow in running mean/variance calculations
type MetricsEngine struct {
	requests    int64 //number of requests so far.
	requestMean int64 //mean elapsed request time in nanoseconds
}

// Snapshot of metrics truncated to milliseconds
//
// This is what we return from GET /stats
type MSMetrics struct {
	Total   uint64 `json:"total"` //number of requests so far
	Average uint64 `json:"average"`
}

// Uses numerically stable recurrence relations to calculate online (running) sample mean/variance:
// M_1 = x_1 ,  M_k = M_{k-1} + (x_k - M_{k-1}) / k
//
// where x_k is the k-th observation, M_{k-1} is the sample mean at observation k-1,
//
// For reference: The Art of Computer Programming Vol. 2 (Seminumerical Algorithms). Knuth, Donald. p232.
// We're technically using fixed precision integer arithmetic using extra precision via the use of nanoseconds for intermediate calculations
//
// Individual metrics are updated atomically and in a lock-free manner
// otherwise locking central metrics store with a mutex quickly becomes a bottleneck (even if we use a channel)
func (m *MetricsEngine) AddDuration(duration int64) {
	d := duration - atomic.LoadInt64(&m.requestMean)                   //x_k - M_{k-1}
	atomic.AddInt64(&m.requestMean, d/atomic.AddInt64(&m.requests, 1)) //M_k = M_{k-1} + (x_k - M_{k-1})/k
}

//Atomically loads snapshot of metrics truncated to milliseconds
func (m *MetricsEngine) MSSnapshot() *MSMetrics {
	s := MSMetrics{
		Total:   uint64(atomic.LoadInt64(&m.requests)),
		Average: uint64(atomic.LoadInt64(&m.requestMean) / 1000000),
	}
	return &s
}

//Encode Millisecond Metrics Snapshot to JSON byte slice
func (m *MSMetrics) MarshalJson() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, nil
}
