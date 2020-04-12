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

	logFields["request_uri"] = r.URL.String()
	logFields["request_query"] = r.URL.Query().Encode()

	// data, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	entry.Logger = entry.Logger.WithFields(logFields)
	// 	return entry
	// }

	// var body GraphQLBody
	// err = json.Unmarshal(data, &body)
	// if err != nil {
	// 	entry.Logger = entry.Logger.WithFields(logFields)
	// 	return entry
	// }

	// slQuery := strings.Fields(body.Query)
	// sQuery := strings.Join(slQuery, " ")
	// slVariables := strings.Fields(string(body.Variables))
	// sVariables := strings.Join(slVariables, " ")
	// sVariables = strings.Replace(sVariables, "\"", "", -1)

	// logFields["query"] = sQuery
	// logFields["variables"] = sVariables

	// r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

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
