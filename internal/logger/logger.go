package logger

import (
    "net/http"
    "os"
    "time"
    
    "github.com/sirupsen/logrus"
)

var Log = logrus.New()

func Init() {
    Log.SetOutput(os.Stdout)
    Log.SetFormatter(&logrus.TextFormatter{
        FullTimestamp:   true,
        TimestampFormat: "2006-01-02 15:04:05",
    })
    Log.SetLevel(logrus.DebugLevel)
    Log.Info("Logger initialized")
}

func AccessLogMiddleware(next http.Handler) http.Handler {
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        
        start := time.Now()
        next.ServeHTTP(w, r)
        
        Log.WithFields(logrus.Fields{
            "method": r.Method,
            "path":   r.URL.Path,
            "remote": r.RemoteAddr,
            "time":   time.Since(start),
        }).Info("request")
    })
}