package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Count to get single numerical value as int from Counter Metric.
func Count(counter prometheus.Counter) (int, error) {
	m := &dto.Metric{}
	e := counter.Write(m)
	return int(m.Counter.GetValue()), e
}

// GaugeValue to get single numerical value as int from Gauge Metric.
func GaugeValue(gauge prometheus.Gauge) (int, error) {
	m := &dto.Metric{}
	e := gauge.Write(m)
	return int(m.Gauge.GetValue()), e
}
