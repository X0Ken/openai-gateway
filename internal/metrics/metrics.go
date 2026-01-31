package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// RequestCounter counts total requests
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// RequestDuration measures request latency
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// ErrorCounter counts errors
	ErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "channel"},
	)

	// ChannelLatency measures channel response time
	ChannelLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_channel_latency_seconds",
			Help:    "Channel response latency in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"channel", "model"},
	)

	// ChannelErrorRate tracks channel error rate
	ChannelErrorRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_channel_error_rate",
			Help: "Channel error rate (0-1)",
		},
		[]string{"channel"},
	)
)

func init() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ErrorCounter)
	prometheus.MustRegister(ChannelLatency)
	prometheus.MustRegister(ChannelErrorRate)
}

// Middleware returns a Gin middleware that collects metrics
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := http.StatusText(c.Writer.Status())

		RequestCounter.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		RequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)

		if c.Writer.Status() >= 400 {
			ErrorCounter.WithLabelValues("http", "").Inc()
		}
	}
}

// Handler returns the Prometheus metrics handler
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RecordChannelLatency records channel response latency
func RecordChannelLatency(channel, model string, duration time.Duration) {
	ChannelLatency.WithLabelValues(channel, model).Observe(duration.Seconds())
}

// RecordChannelError records a channel error
func RecordChannelError(channel string) {
	ErrorCounter.WithLabelValues("channel", channel).Inc()
}

// SetChannelErrorRate sets the error rate for a channel
func SetChannelErrorRate(channel string, rate float64) {
	ChannelErrorRate.WithLabelValues(channel).Set(rate)
}
