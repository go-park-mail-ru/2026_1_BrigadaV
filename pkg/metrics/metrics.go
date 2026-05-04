package metrics

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	GRPCRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests.",
	}, []string{"service", "method", "status"})

	GRPCRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_request_duration_seconds",
		Help:    "gRPC request latency.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method"})
)

func init() {
	prometheus.MustRegister(HTTPRequestsTotal, HTTPRequestDuration)
	prometheus.MustRegister(GRPCRequestsTotal, GRPCRequestDuration)
}

func StartMetricsServer(port string) {
	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":"+port, r); err != nil {
			panic(err)
		}
	}()
}

func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return promhttp.InstrumentHandlerCounter(HTTPRequestsTotal,
		promhttp.InstrumentHandlerDuration(HTTPRequestDuration, next),
	)
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		timer := prometheus.NewTimer(GRPCRequestDuration.WithLabelValues(info.FullMethod, ""))
		defer timer.ObserveDuration()
		resp, err := handler(ctx, req)
		status := "OK"
		if err != nil {
			status = "ERROR"
		}
		GRPCRequestsTotal.WithLabelValues(info.FullMethod, "", status).Inc()
		return resp, err
	}
}
