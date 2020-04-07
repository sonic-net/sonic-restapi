package mseeserver

import (
    "github.com/comail/colog"
    "log"
    "net/http"
    "os"
)

type LoggingResponseWriter struct {
    inner http.ResponseWriter
}

func (w LoggingResponseWriter) Header() http.Header {
    return w.inner.Header()
}

func (w LoggingResponseWriter) Write(b []byte) (int, error) {
    return w.inner.Write(b)
}

func (w LoggingResponseWriter) WriteHeader(statusCode int) {
    log.Printf("info: request: return %d", statusCode)
    w.inner.WriteHeader(statusCode)
}

func NewLoggingResponseWriter(w http.ResponseWriter) LoggingResponseWriter {
    return LoggingResponseWriter{inner: w}
}

func InitLogging() {
    colog.Register()

    level, err := colog.ParseLevel(*LogLevelFlag)
    if err != nil {
        log.Fatalf("error: invalid minimum log level %s", *LogLevelFlag)
    }
    colog.SetMinLevel(level)

    file, err := os.OpenFile(*LogFileFlag, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("error: couldn't open log file %s, %s", *LogFileFlag, err)
    }
    colog.SetOutput(file)
}