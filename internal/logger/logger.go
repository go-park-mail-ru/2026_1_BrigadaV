package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

var Log *logrus.Logger

func Init(level string) {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	Log.SetLevel(lvl)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		r = r.WithContext(ctx)

		if Log != nil {
			Log.WithFields(logrus.Fields{
				"request_id": reqID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"remote":     r.RemoteAddr,
			}).Info("request started")
		}

		next.ServeHTTP(w, r)
	})
}

func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func Info(ctx context.Context, msg string, fields logrus.Fields) {
	if Log == nil {
		return
	}
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["request_id"] = GetRequestID(ctx)
	Log.WithFields(fields).Info(msg)
}

func Error(ctx context.Context, msg string, fields logrus.Fields) {
	if Log == nil {
		return
	}
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["request_id"] = GetRequestID(ctx)
	Log.WithFields(fields).Error(msg)
}

func Debug(ctx context.Context, msg string, fields logrus.Fields) {
	if Log == nil {
		return
	}
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["request_id"] = GetRequestID(ctx)
	Log.WithFields(fields).Debug(msg)
}

func Warn(ctx context.Context, msg string, fields logrus.Fields) {
	if Log == nil {
		return
	}
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["request_id"] = GetRequestID(ctx)
	Log.WithFields(fields).Warn(msg)
}
