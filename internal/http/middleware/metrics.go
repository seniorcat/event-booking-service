package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "event_booking_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_booking_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route"},
	)
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "event_booking_http_errors_total",
			Help: "Total http server errors (5xx)",
		},
		[]string{"method", "route", "status"},
	)

	//дополнительная метрика
	BookingsCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "booking_bookings_total",
			Help: "Total number of bookings",
		},
		[]string{"event_id", "status"},
	)
)

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &metricsResponseWriter{ResponseWriter: w}
		defer func() {
			if err := recover(); err != nil {
				if !rw.wroteHeader {
					rw.WriteHeader(http.StatusInternalServerError)
				}
				panic(err)
			}

		}()

		next.ServeHTTP(rw, r)
		if !rw.wroteHeader {
			rw.statusCode = http.StatusOK
		}
		duration := time.Since(start).Seconds()
		method := r.Method
		route := r.URL.Path
		status := strconv.Itoa(rw.statusCode)
		requestsTotal.WithLabelValues(method, route, status).Inc()

		if rw.statusCode >= 500 && rw.statusCode < 600 {
			errorsTotal.WithLabelValues(method, route, status).Inc()
		}

		requestDuration.WithLabelValues(method, route).Observe(duration)

	})
}
