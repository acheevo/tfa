package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// InMemoryCollector implements MetricsCollector using in-memory storage
type InMemoryCollector struct {
	metrics map[string]*inMemoryMetric
	mu      sync.RWMutex
	logger  *slog.Logger
}

// inMemoryMetric represents a metric stored in memory
type inMemoryMetric struct {
	definition *MetricDefinition
	value      float64
	labels     map[string]string
	timestamp  time.Time
	// For histograms
	buckets map[float64]uint64
	sum     float64
	count   uint64
	// For summaries
	samples []float64
}

// timer implements the Timer interface
type timer struct {
	name      string
	labels    map[string]string
	startTime time.Time
	collector *InMemoryCollector
}

// NewInMemoryCollector creates a new in-memory metrics collector
func NewInMemoryCollector(logger *slog.Logger) *InMemoryCollector {
	return &InMemoryCollector{
		metrics: make(map[string]*inMemoryMetric),
		logger:  logger,
	}
}

// IncrementCounter increments a counter metric by 1
func (c *InMemoryCollector) IncrementCounter(name string, labels map[string]string) error {
	return c.IncrementCounterBy(name, 1, labels)
}

// IncrementCounterBy increments a counter metric by the specified value
func (c *InMemoryCollector) IncrementCounterBy(name string, value float64, labels map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.getMetricKey(name, labels)
	metric, exists := c.metrics[key]

	if !exists {
		metric = &inMemoryMetric{
			value:     0,
			labels:    labels,
			timestamp: time.Now(),
		}
		c.metrics[key] = metric
	}

	metric.value += value
	metric.timestamp = time.Now()

	return nil
}

// SetGauge sets a gauge metric to the specified value
func (c *InMemoryCollector) SetGauge(name string, value float64, labels map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.getMetricKey(name, labels)
	metric, exists := c.metrics[key]

	if !exists {
		metric = &inMemoryMetric{
			labels: labels,
		}
		c.metrics[key] = metric
	}

	metric.value = value
	metric.timestamp = time.Now()

	return nil
}

// IncrementGauge increments a gauge metric by 1
func (c *InMemoryCollector) IncrementGauge(name string, labels map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.getMetricKey(name, labels)
	metric, exists := c.metrics[key]

	if !exists {
		metric = &inMemoryMetric{
			value:  0,
			labels: labels,
		}
		c.metrics[key] = metric
	}

	metric.value++
	metric.timestamp = time.Now()

	return nil
}

// DecrementGauge decrements a gauge metric by 1
func (c *InMemoryCollector) DecrementGauge(name string, labels map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.getMetricKey(name, labels)
	metric, exists := c.metrics[key]

	if !exists {
		metric = &inMemoryMetric{
			value:  0,
			labels: labels,
		}
		c.metrics[key] = metric
	}

	metric.value--
	metric.timestamp = time.Now()

	return nil
}

// ObserveHistogram observes a value for a histogram metric
func (c *InMemoryCollector) ObserveHistogram(name string, value float64, labels map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.getMetricKey(name, labels)
	metric, exists := c.metrics[key]

	if !exists {
		metric = &inMemoryMetric{
			labels:  labels,
			buckets: make(map[float64]uint64),
			sum:     0,
			count:   0,
		}
		c.metrics[key] = metric
	}

	// Update histogram buckets
	if metric.definition != nil && metric.definition.Buckets != nil {
		for _, bucket := range metric.definition.Buckets {
			if value <= bucket {
				metric.buckets[bucket]++
			}
		}
	}

	metric.sum += value
	metric.count++
	metric.timestamp = time.Now()

	return nil
}

// ObserveSummary observes a value for a summary metric
func (c *InMemoryCollector) ObserveSummary(name string, value float64, labels map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.getMetricKey(name, labels)
	metric, exists := c.metrics[key]

	if !exists {
		metric = &inMemoryMetric{
			labels:  labels,
			samples: make([]float64, 0),
		}
		c.metrics[key] = metric
	}

	// Add sample (keep only last 1000 samples for memory efficiency)
	metric.samples = append(metric.samples, value)
	if len(metric.samples) > 1000 {
		metric.samples = metric.samples[1:]
	}

	metric.timestamp = time.Now()

	return nil
}

// StartTimer starts a timer for measuring duration
func (c *InMemoryCollector) StartTimer(name string, labels map[string]string) Timer {
	return &timer{
		name:      name,
		labels:    labels,
		startTime: time.Now(),
		collector: c,
	}
}

// RecordDuration records a duration measurement
func (c *InMemoryCollector) RecordDuration(name string, duration time.Duration, labels map[string]string) error {
	return c.ObserveHistogram(name, duration.Seconds(), labels)
}

// RegisterMetric registers a metric definition
func (c *InMemoryCollector) RegisterMetric(metric *MetricDefinition) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize metric if needed
	key := metric.Name
	if _, exists := c.metrics[key]; !exists {
		inMemMetric := &inMemoryMetric{
			definition: metric,
			timestamp:  time.Now(),
		}

		// Initialize type-specific fields
		switch metric.Type {
		case MetricTypeHistogram:
			inMemMetric.buckets = make(map[float64]uint64)
			if metric.Buckets != nil {
				for _, bucket := range metric.Buckets {
					inMemMetric.buckets[bucket] = 0
				}
			}
		case MetricTypeSummary:
			inMemMetric.samples = make([]float64, 0)
		}

		c.metrics[key] = inMemMetric
	}

	c.logger.Debug("Metric registered", "name", metric.Name, "type", metric.Type)
	return nil
}

// Collect returns all current metrics
func (c *InMemoryCollector) Collect(ctx context.Context) ([]*Metric, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var metrics []*Metric

	for key, metric := range c.metrics {
		// Extract metric name from key (remove label hash)
		name := c.extractMetricName(key)

		m := &Metric{
			Name:      name,
			Value:     metric.value,
			Labels:    metric.labels,
			Timestamp: metric.timestamp,
		}

		if metric.definition != nil {
			m.Type = metric.definition.Type
			m.Help = metric.definition.Help
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

// Stop stops the timer and returns the elapsed duration
func (t *timer) Stop() time.Duration {
	return time.Since(t.startTime)
}

// StopAndRecord stops the timer and records the duration
func (t *timer) StopAndRecord() error {
	duration := time.Since(t.startTime)
	return t.collector.RecordDuration(t.name, duration, t.labels)
}

// Helper methods

// getMetricKey generates a unique key for a metric with labels
func (c *InMemoryCollector) getMetricKey(name string, labels map[string]string) string {
	key := name
	if len(labels) > 0 {
		for k, v := range labels {
			key += ";" + k + "=" + v
		}
	}
	return key
}

// extractMetricName extracts the metric name from a metric key
func (c *InMemoryCollector) extractMetricName(key string) string {
	// Find the first semicolon which separates name from labels
	for i, char := range key {
		if char == ';' {
			return key[:i]
		}
	}
	return key
}

// GetCurrentValue returns the current value of a metric
func (c *InMemoryCollector) GetCurrentValue(name string, labels map[string]string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.getMetricKey(name, labels)
	metric, exists := c.metrics[key]
	if !exists {
		return 0, false
	}

	return metric.value, true
}

// GetMetricCount returns the number of registered metrics
func (c *InMemoryCollector) GetMetricCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.metrics)
}

// Reset resets all metrics to their initial values
func (c *InMemoryCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, metric := range c.metrics {
		metric.value = 0
		metric.sum = 0
		metric.count = 0
		if metric.buckets != nil {
			for bucket := range metric.buckets {
				metric.buckets[bucket] = 0
			}
		}
		if metric.samples != nil {
			metric.samples = metric.samples[:0]
		}
		metric.timestamp = time.Now()
	}
}

// GetStats returns statistics about the collector
func (c *InMemoryCollector) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"total_metrics": len(c.metrics),
		"memory_usage":  "estimated", // Could calculate actual memory usage
	}

	// Count metrics by type
	typeCount := make(map[MetricType]int)
	for _, metric := range c.metrics {
		if metric.definition != nil {
			typeCount[metric.definition.Type]++
		}
	}

	stats["metrics_by_type"] = typeCount

	return stats
}
