package function

import (
	"fmt"
	"log"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	"go.opencensus.io/trace"
)

// Service stores necessary information for the cloud function
type Service struct {
	DateFmt         string
	DbCollection    string
	ErrorReporter   *errorreporting.Client
	FirestoreClient *firestore.Client
	FunctionName    string
	Logger          *logging.Client
	ProjectID       string
	TraceSpan       *trace.Span
	Version         string
}

// LogMessage is a simple struct to ensure JSON formatting in logs
type LogMessage struct {
	Err      string `json:",omitempty"`
	Function string `json:"fn"`
	Msg      string `json:"msg"`
	Version  string `json:"version"`
}

// HandleFatalError produces an error report, cloud log message and standard log fatal
func (g Service) HandleFatalError(msg string, err error) {
	g.ErrorReporter.Report(errorreporting.Entry{Error: err})
	g.Logger.Logger(g.FunctionName).Log(logging.Entry{
		Severity: logging.Error,
		Payload: LogMessage{
			Msg:      msg,
			Err:      err.Error(),
			Function: g.FunctionName,
			Version:  g.Version,
		},
		Trace:  fmt.Sprintf("projects/%s/trace/%s", g.ProjectID, g.TraceSpan.SpanContext().TraceID.String()),
		SpanID: g.TraceSpan.SpanContext().SpanID.String(),
	})
	log.Fatalf("%s: %s", msg, err)
}

// DebugMsg logs a simple debug message with function name and version
func (g Service) DebugMsg(msg string) {
	g.Logger.Logger(g.FunctionName).Log(logging.Entry{
		Severity: logging.Debug,
		Payload: LogMessage{
			Msg:      msg,
			Function: g.FunctionName,
			Version:  g.Version,
		},
	})
}
