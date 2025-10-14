// Package semconv ...
package semconv

import (
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

// PrometheusLabel converts a key to a label for Prometheus metrics.
func PrometheusLabel(key attribute.Key) string {
	return strings.Replace(string(key), ".", "_", -1)
}

// PrometheusLabels converts a slice of attribute keys to a slice of labels suitable for Prometheus metrics.
func PrometheusLabels(key ...attribute.Key) []string {
	ret := make([]string, len(key))
	for i, k := range key {
		ret[i] = PrometheusLabel(k)
	}
	return ret
}

// PrometheusValue converts a key and value to a Prometheus metric label values.
func PrometheusValue(kv attribute.KeyValue) (string, string) {
	return PrometheusLabel(kv.Key), kv.Value.Emit()
}
