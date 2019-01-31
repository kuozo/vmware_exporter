package pkg

import (
	"fmt"
	"strings"
)

// CalMetricPercent return percent
func CalMetricPercent(member, total float64) float64 {
	if total <= 0 {
		return 0
	}
	percent := 100 * member / total
	return percent
}

// GenerateMetricName return metric name
func GenerateMetricName(namespace, metricName string) string {
	var b strings.Builder
	fmt.Fprint(&b, namespace, "_", metricName)
	return b.String()
}
