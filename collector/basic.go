package collector

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "cloudwartch"
)

type basicMetric struct {
	Type prometheus.ValueType
	Desc *prometheus.Desc
}
