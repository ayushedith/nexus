package load

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nexusapi/nexus/pkg/collection"
)

type Config struct {
	VirtualUsers int
	Duration     time.Duration
	RampUp       time.Duration
	RampDown     time.Duration
	Iterations   int
}

type Engine struct {
	config  *Config
	runner  *collection.Runner
	metrics *Metrics
}

type Metrics struct {
	totalRequests   atomic.Int64
	successRequests atomic.Int64
	failedRequests  atomic.Int64
	totalLatency    atomic.Int64
	minLatency      atomic.Int64
	maxLatency      atomic.Int64
	mu              sync.RWMutex
	latencies       []time.Duration
}

func NewEngine(cfg *Config, runner *collection.Runner) *Engine {
	return &Engine{
		config:  cfg,
		runner:  runner,
		metrics: &Metrics{latencies: make([]time.Duration, 0, 1000)},
	}
}

func (e *Engine) Run(ctx context.Context, req collection.Request) (*LoadTestResult, error) {
	startTime := time.Now()

	workerCount := e.config.VirtualUsers
	reqChan := make(chan struct{}, workerCount*10)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			e.worker(ctx, req, reqChan)
		}(i)
	}

	go e.distributeLoad(ctx, reqChan)

	if e.config.Duration > 0 {
		timer := time.NewTimer(e.config.Duration)
		<-timer.C
		close(reqChan)
	} else if e.config.Iterations > 0 {
		for i := 0; i < e.config.Iterations; i++ {
			select {
			case <-ctx.Done():
				close(reqChan)
				goto wait
			default:
				reqChan <- struct{}{}
			}
		}
		close(reqChan)
	}

wait:
	wg.Wait()

	return &LoadTestResult{
		TotalRequests:   e.metrics.totalRequests.Load(),
		SuccessRequests: e.metrics.successRequests.Load(),
		FailedRequests:  e.metrics.failedRequests.Load(),
		Duration:        time.Since(startTime),
		RPS:             float64(e.metrics.totalRequests.Load()) / time.Since(startTime).Seconds(),
		AvgLatency:      e.calculateAvgLatency(),
		MinLatency:      time.Duration(e.metrics.minLatency.Load()),
		MaxLatency:      time.Duration(e.metrics.maxLatency.Load()),
		P50Latency:      e.calculatePercentile(0.50),
		P95Latency:      e.calculatePercentile(0.95),
		P99Latency:      e.calculatePercentile(0.99),
	}, nil
}

func (e *Engine) distributeLoad(ctx context.Context, reqChan chan struct{}) {
	if e.config.RampUp > 0 {
		interval := e.config.RampUp / time.Duration(e.config.VirtualUsers)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for i := 0; i < e.config.VirtualUsers; i++ {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}
}

func (e *Engine) worker(ctx context.Context, req collection.Request, reqChan <-chan struct{}) {
	for range reqChan {
		select {
		case <-ctx.Done():
			return
		default:
			result := e.runner.ExecuteRequest(req)
			e.recordMetrics(result)
		}
	}
}

func (e *Engine) recordMetrics(result collection.ExecutionResult) {
	e.metrics.totalRequests.Add(1)

	if result.Error == nil && result.Response.StatusCode < 400 {
		e.metrics.successRequests.Add(1)
	} else {
		e.metrics.failedRequests.Add(1)
	}

	latency := result.Response.Time
	e.metrics.totalLatency.Add(int64(latency))

	for {
		min := e.metrics.minLatency.Load()
		if min == 0 || int64(latency) < min {
			if e.metrics.minLatency.CompareAndSwap(min, int64(latency)) {
				break
			}
		} else {
			break
		}
	}

	for {
		max := e.metrics.maxLatency.Load()
		if int64(latency) > max {
			if e.metrics.maxLatency.CompareAndSwap(max, int64(latency)) {
				break
			}
		} else {
			break
		}
	}

	e.metrics.mu.Lock()
	e.metrics.latencies = append(e.metrics.latencies, latency)
	e.metrics.mu.Unlock()
}

func (e *Engine) calculateAvgLatency() time.Duration {
	total := e.metrics.totalRequests.Load()
	if total == 0 {
		return 0
	}
	return time.Duration(e.metrics.totalLatency.Load() / total)
}

func (e *Engine) calculatePercentile(p float64) time.Duration {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()

	if len(e.metrics.latencies) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(e.metrics.latencies))
	copy(sorted, e.metrics.latencies)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	idx := int(float64(len(sorted)) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}

	return sorted[idx]
}

type LoadTestResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	Duration        time.Duration
	RPS             float64
	AvgLatency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
}

func (r *LoadTestResult) String() string {
	return fmt.Sprintf(`Load Test Results:
  Total Requests: %d
  Success: %d (%.2f%%)
  Failed: %d (%.2f%%)
  Duration: %v
  RPS: %.2f
  Avg Latency: %v
  Min Latency: %v
  Max Latency: %v
  P50 Latency: %v
  P95 Latency: %v
  P99 Latency: %v`,
		r.TotalRequests,
		r.SuccessRequests, float64(r.SuccessRequests)/float64(r.TotalRequests)*100,
		r.FailedRequests, float64(r.FailedRequests)/float64(r.TotalRequests)*100,
		r.Duration,
		r.RPS,
		r.AvgLatency,
		r.MinLatency,
		r.MaxLatency,
		r.P50Latency,
		r.P95Latency,
		r.P99Latency,
	)
}
