package function

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	"contrib.go.opencensus.io/exporter/stackdriver"
	firebase "firebase.google.com/go"
	"go.opencensus.io/trace"
)

// Service stores necessary information for the cloud function
type Service struct {
	Date            time.Time
	DateFmt         string
	DBCollection    string
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
	Date         string `json:"date"`
	DBCollection string `json:"dbCollection"`
	Err          string `json:",omitempty"`
	FunctionName string `json:"functionName"`
	Msg          string `json:"msg"`
	ProjectID    string `json:"projectID"`
	Version      string `json:"version"`
}

// DebugMsg logs a simple debug message with function name and version
func (s Service) DebugMsg(msg string) {
	s.Logger.Logger(s.FunctionName).Log(logging.Entry{
		Severity: logging.Debug,
		Payload: LogMessage{
			Date:         s.Date.Format(s.DateFmt),
			DBCollection: s.DBCollection,
			FunctionName: s.FunctionName,
			Msg:          msg,
			ProjectID:    s.ProjectID,
			Version:      s.Version,
		},
	})
}

// HandleFatalError produces an error report, cloud log message and standard log fatal
func (s Service) HandleFatalError(msg string, err error) {
	s.ErrorReporter.Report(errorreporting.Entry{Error: err})
	s.Logger.Logger(s.FunctionName).Log(logging.Entry{
		Severity: logging.Error,
		Payload: LogMessage{
			Date:         s.Date.Format(s.DateFmt),
			DBCollection: s.DBCollection,
			Err:          err.Error(),
			FunctionName: s.FunctionName,
			Msg:          msg,
			ProjectID:    s.ProjectID,
			Version:      s.Version,
		},
		Trace:  fmt.Sprintf("projects/%s/trace/%s", s.ProjectID, s.TraceSpan.SpanContext().TraceID.String()),
		SpanID: s.TraceSpan.SpanContext().SpanID.String(),
	})
	log.Fatalf("%s: %s", msg, err)
}

// InitService initializes the function service with default
func InitService(ctx context.Context) (Service, error) {
	s := Service{
		DateFmt:      os.Getenv("DATE_FMT"),
		DBCollection: os.Getenv("DB_COLLECTION"),
		ProjectID:    os.Getenv("PROJECT_ID"),
		FunctionName: os.Getenv("FN_NAME"),
		Version:      os.Getenv("VERSION"),
	}

	// Tracing
	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: s.ProjectID})
	if err != nil {
		return Service{}, fmt.Errorf("error setting up OpenCensus Stackdriver Trace exporter: %s", err)
	}
	trace.RegisterExporter(exporter)
	ctx, span := trace.StartSpan(ctx, s.FunctionName)
	s.TraceSpan = span

	// Error Reporting
	errorClient, err := errorreporting.NewClient(ctx, s.ProjectID, errorreporting.Config{
		ServiceName:    s.FunctionName,
		ServiceVersion: s.Version,
	})
	if err != nil {
		return Service{}, fmt.Errorf("error setting up Error Reporting: %s", err)
	}
	s.ErrorReporter = errorClient

	// Cloud Logging
	logClient, err := logging.NewClient(ctx, s.ProjectID)
	if err != nil {
		return Service{}, fmt.Errorf("error setting up Google Cloud logger: %s", err)
	}
	s.Logger = logClient

	// Firestore
	conf := &firebase.Config{ProjectID: s.ProjectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return Service{}, fmt.Errorf("error setting up Firebase app: %s", err)
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		return Service{}, fmt.Errorf("error setting up Firestore client: %s", err)
	}
	s.FirestoreClient = fsClient

	return s, nil
}
