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
	Err     string  `json:",omitempty"`
	Msg     string  `json:"msg"`
	Service Service `json:"service"`
}

// DebugMsg logs a simple debug message with function name and version
func (s Service) DebugMsg(msg string) {
	s.Logger.Logger(s.FunctionName).Log(logging.Entry{
		Severity: logging.Debug,
		Payload: LogMessage{
			Msg:     msg,
			Service: s,
		},
	})
}

// HandleFatalError produces an error report, cloud log message and standard log fatal
func (s Service) HandleFatalError(msg string, err error) {
	s.ErrorReporter.Report(errorreporting.Entry{Error: err})
	s.Logger.Logger(s.FunctionName).Log(logging.Entry{
		Severity: logging.Error,
		Payload: LogMessage{
			Msg:     msg,
			Err:     err.Error(),
			Service: s,
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
		DbCollection: os.Getenv("DB_COLLECTION"),
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
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		return Service{}, fmt.Errorf("error setting up Error Reporting: %s", err)
	}
	defer errorClient.Close()
	s.ErrorReporter = errorClient

	// Cloud Logging
	logClient, err := logging.NewClient(ctx, s.ProjectID)
	if err != nil {
		return Service{}, fmt.Errorf("error setting up Google Cloud logger: %s", err)
	}
	defer logClient.Close()
	logClient.OnError = func(e error) {
		fmt.Fprintf(os.Stderr, "logClient error: %v", e)
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
