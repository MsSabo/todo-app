package ads

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var requestMetrics = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Namespace:  "ads",
	Subsystem:  "http",
	Name:       "request",
	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
}, []string{"status"})

var httpRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Response time of HTTP request",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"status"},
)

var usersActionsCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "user_actions_total",
		Help: "Total number of actions performed by users",
	},
	[]string{"user_id"},
)

func IncrementUserAction(id int) {
	usersActionsCounter.WithLabelValues(strconv.Itoa(id)).Inc()
}

func ObserveRequest(d time.Duration, status int) {
	httpRequestDuration.WithLabelValues(strconv.Itoa(status)).Observe(d.Seconds())
}
