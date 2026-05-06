package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"guidely-app/internal/logger"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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
			logger.Log.WithFields(logrus.Fields{
				"port": port,
				"err":  err,
			}).Error("metrics server stopped")
		}
	}()
}

// responseWriter перехватывает статус-код ответа.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// routeTemplate возвращает шаблон маршрута gorilla/mux (например /api/places/{id})
// вместо реального пути, чтобы не раздувать кардинальность метрик.
func routeTemplate(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if tmpl, err := route.GetPathTemplate(); err == nil {
			return tmpl
		}
	}
	return r.URL.Path
}

// HTTPMetricsMiddleware собирает hits, тайминги и статусы по шаблону маршрута,
// а также логирует каждый запрос через общий логгер проекта.
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		tmpl := routeTemplate(r)
		duration := time.Since(start)
		status := strconv.Itoa(wrapped.statusCode)

		HTTPRequestsTotal.WithLabelValues(r.Method, tmpl, status).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, tmpl).Observe(duration.Seconds())

		logger.Log.WithFields(logrus.Fields{
			"request_id": logger.GetRequestID(r.Context()),
			"method":     r.Method,
			"route":      tmpl,
			"status":     wrapped.statusCode,
			"duration":   fmt.Sprintf("%.3fms", float64(duration.Nanoseconds())/1e6),
		}).Info("http request metrics")
	})
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		service := info.FullMethod
		method := info.FullMethod
		if idx := len(info.FullMethod); idx > 0 {
			parts := splitFullMethod(info.FullMethod)
			service = parts[0]
			method = parts[1]
		}

		timer := prometheus.NewTimer(GRPCRequestDuration.WithLabelValues(service, method))
		defer timer.ObserveDuration()

		resp, err := handler(ctx, req)
		status := "OK"
		if err != nil {
			status = "ERROR"
		}
		GRPCRequestsTotal.WithLabelValues(service, method, status).Inc()
		return resp, err
	}
}

func splitFullMethod(fullMethod string) [2]string {
	if len(fullMethod) > 0 && fullMethod[0] == '/' {
		fullMethod = fullMethod[1:]
	}
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			return [2]string{fullMethod[:i], fullMethod[i+1:]}
		}
	}
	return [2]string{fullMethod, ""}
}
