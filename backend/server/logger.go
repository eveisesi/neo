package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {

	entry := logrus.NewEntry(logger)

	return middleware.RequestLogger(&StructuredLogger{entry})
}

// StructuredLogger holds our application's instance of our logger
type StructuredLogger struct {
	Logger *logrus.Entry
}

type GraphQLBody struct {
	Query     string          `json:"query"`
	Variables json.RawMessage `json:"variables"`
}

// NewLogEntry will return a new log entry scoped to the http.Request
func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: l.Logger}
	logFields := logrus.Fields{}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	logFields["user_agent"] = r.UserAgent()

	logFields["remote_addr"] = r.RemoteAddr

	logFields["request_path"] = r.URL.Path

	entry.Logger = entry.Logger.WithFields(logFields)

	return entry
}

// StructuredLoggerEntry holds our FieldLogger entry
type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

// Write will write to logger entry once the http.Request is complete
func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"status": status, "resp_bytes_length": bytes,
		"elapsed_ms": elapsed.Milliseconds(),
	})

	l.Logger.Info("received request")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}

// // Helper methods used by the application to get the request-scoped
// // logger entry and set additional fields between handlers.
// //
// // This is a useful pattern to use to set state on the entry as it
// // passes through the handler chain, which at any point can be logged
// // with a call to .Print(), .Info(), etc.

// // GetLogEntry will get return the logger off of the http request
// func GetLogEntry(r *http.Request) logrus.FieldLogger {
// 	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
// 	return entry.Logger
// }

// // LogEntrySetField will set a new field on a log entry
// func LogEntrySetField(r *http.Request, key string, value interface{}) {
// 	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
// 		entry.Logger = entry.Logger.WithField(key, value)
// 	}
// }

// // LogEntrySetFields will set a map of key/value pairs on a log entry
// func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
// 	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
// 		entry.Logger = entry.Logger.WithFields(fields)
// 	}
// }
