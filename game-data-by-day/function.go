package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	firebase "firebase.google.com/go"

	"github.com/unrealities/warning-track-backend/mlbStats"
)

// GameDataByDay stores necessary information for the cloud function
type GameDataByDay struct {
	DateFmt         string
	DbCollection    string
	ErrorReporter   *errorreporting.Client
	FirestoreClient *firestore.Client
	FunctionName    string
	Logger          *logging.Client
	ProjectID       string
	Timeout         time.Duration
	Version         string
}

// LogMessage is a simple struct to ensure JSON formatting in logs
type LogMessage struct {
	Err      string `json:",omitempty"`
	Function string `json:"fn"`
	Msg      string `json:"msg"`
	Version  string `json:"version"`
}

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"date":"03-01-2020"}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	gameDataByDay := GameDataByDay{
		DateFmt:      "01-02-2006",
		DbCollection: "game-data-by-day",
		ProjectID:    "warning-track-backend",
		FunctionName: "GetGameDataByDay",
		Timeout:      60 * time.Second,
		Version:      "v0.0.58",
	}
	log.Printf("running version: %s", gameDataByDay.Version)

	var err error
	ctx := context.Background()

	// // Tracing
	// exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: gameDataByDay.ProjectID})
	// if err != nil {
	// 	log.Fatalf("error setting up OpenCensus Stackdriver Trace exporter: %s", err)
	// }
	// trace.RegisterExporter(exporter)

	// // Error Reporting
	// errorClient, err := errorreporting.NewClient(ctx, gameDataByDay.ProjectID, errorreporting.Config{
	// 	ServiceName:    gameDataByDay.FunctionName,
	// 	ServiceVersion: gameDataByDay.Version,
	// 	OnError: func(err error) {
	// 		log.Printf("Could not log error: %v", err)
	// 	},
	// })
	// if err != nil {
	// 	log.Fatalf("error setting up Error Reporting: %s", err)
	// }
	// defer errorClient.Close()
	// gameDataByDay.ErrorReporter = errorClient

	// Cloud Logging
	logClient, err := logging.NewClient(ctx, gameDataByDay.ProjectID)
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger: %s", err)
	}
	defer logClient.Close()
	logClient.OnError = func(e error) {
		fmt.Fprintf(os.Stderr, "logClient error: %v", e)
	}
	gameDataByDay.Logger = logClient

	// Firestore
	conf := &firebase.Config{ProjectID: gameDataByDay.ProjectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		gameDataByDay.HandleFatalError("error setting up Firebase app", err)
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		gameDataByDay.HandleFatalError("error setting up Firestore client", err)
	}
	collection := fsClient.Collection(gameDataByDay.DbCollection)

	gameDataByDay.DebugMsg("successfully fetched Firestore collection")

	date, err := parseDate(r.Body, gameDataByDay.DateFmt)
	if err != nil {
		gameDataByDay.HandleFatalError("error parsing date requested", err)
	}

	daySchedule, err := mlbStats.GetSchedule(date)
	if err != nil {
		gameDataByDay.HandleFatalError("error getting the daily StatsAPI schedule", err)
	}
	gameDataByDay.DebugMsg("successfully fetched schedule")

	log.Printf("daySchedule.Dates[0].Games[0].GameNumber: %v", daySchedule.Dates[0].Games[0].GameNumber)
	_, err = collection.Doc(date.Format(gameDataByDay.DateFmt)).Set(ctx, daySchedule.Dates[0].Games[0])
	if err != nil {
		gameDataByDay.HandleFatalError("error persisting data to Firebase", err)
	}

	logClient.Logger(gameDataByDay.FunctionName).Log(logging.Entry{
		Severity: logging.Debug,
		Payload: LogMessage{
			Msg:      "stackdriver success",
			Function: gameDataByDay.FunctionName,
			Version:  gameDataByDay.Version,
		},
	})

	gameDataByDay.DebugMsg("stackdriver debug message success")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule.Dates[0].Games[0])
}

// parseDate parses the request body and returns a time.Time value of the requested date
func parseDate(reqBody io.ReadCloser, dateFormat string) (time.Time, error) {
	var d struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(reqBody).Decode(&d); err != nil {
		return time.Time{}, err
	}

	return time.Parse(dateFormat, d.Date)
}

// HandleFatalError produces an error report, cloud log message and standard log fatal
// TODO: include tracing (Trace in logging.Entry)
func (g GameDataByDay) HandleFatalError(msg string, err error) {
	// g.errorReporter.Report(errorreporting.Entry{Error: err})
	// g.logger.Logger(g.FunctionName).Log(logging.Entry{
	// 	Severity: logging.Error,
	// 	Payload: logMessage{
	// 		Msg:      msg,
	// 		Err:      err.Error(),
	// 		Function: g.FunctionName,
	// 		Version:  g.Version,
	// 	},
	// })
	log.Fatalf("%s: %s", msg, err)
}

// DebugMsg logs a simple debug message with function name and version
func (g GameDataByDay) DebugMsg(msg string) {
	g.Logger.Logger(g.FunctionName).Log(logging.Entry{
		Severity: logging.Debug,
		Payload: LogMessage{
			Msg:      msg,
			Function: g.FunctionName,
			Version:  g.Version,
		},
	})
	log.Println(msg)
}
