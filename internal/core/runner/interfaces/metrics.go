package interfaces

type MetricStore interface {
	RegisterMetric(name string, metricType MetricType)
	SetMetric(name string, value float64)
	GetMetric(name string) (float64, bool)
	Aggregate(name string, method AggregationMethod) (float64, error)
	AllMetrics() map[string]float64
}

type MetricType string

const (
	GaugeMetric     MetricType = "gauge"
	CounterMetric   MetricType = "counter"
	HistogramMetric MetricType = "histogram"
)

type AggregationMethod string

const (
	SumAggregation AggregationMethod = "sum"
	AvgAggregation AggregationMethod = "avg"
	MinAggregation AggregationMethod = "min"
	MaxAggregation AggregationMethod = "max"
)
