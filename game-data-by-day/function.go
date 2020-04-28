package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	"contrib.go.opencensus.io/exporter/stackdriver"
	firebase "firebase.google.com/go"
	"go.opencensus.io/trace"

	"github.com/unrealities/warning-track-backend/mlbStats"
)

// gameDataByDay stores necessary information for the cloud function
type gameDataByDay struct {
	dateFmt        string
	dbCollection   string
	duration       time.Duration
	errorReporter  *errorreporting.Client
	firebaseDomain string
	functionName   string
	logger         *logging.Logger
	projectID      string
	version        string
}

// logMessage is a simple struct to ensure JSON formatting in logs
type logMessage struct {
	err string
	msg string
}

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
// ex. POST request:
// https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {"date":"03-01-2020"}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	gameDataByDay := gameDataByDay{
		dateFmt:        "01-02-2006",
		dbCollection:   "game-data-by-day",
		duration:       60 * time.Second,
		firebaseDomain: "firebaseio.com",
		projectID:      "warning-track-backend",
		functionName:   "GetGameDataByDay",
		version:        "v0.0.41",
	}
	log.Printf("running version: %s", gameDataByDay.version)

	var err error
	ctx := context.Background()

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: gameDataByDay.projectID})
	if err != nil {
		log.Fatalf("error setting up OpenCensus Stackdriver Trace exporter: %s", err)
	}
	trace.RegisterExporter(exporter)

	// Error Reporting
	errorClient, err := errorreporting.NewClient(ctx, gameDataByDay.projectID, errorreporting.Config{
		ServiceName:    gameDataByDay.functionName,
		ServiceVersion: gameDataByDay.version,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("error setting up Error Reporting: %s", err)
	}
	defer errorClient.Close()
	gameDataByDay.errorReporter = errorClient

	// Cloud Logging
	logClient, err := logging.NewClient(ctx, gameDataByDay.projectID)
	if err := logClient.Close(); err != nil {
		log.Fatalf("error setting up Google Cloud logger: %s", err)
	}
	defer logClient.Close()
	gameDataByDay.logger = logClient.Logger(gameDataByDay.functionName)

	collection, err := fireStoreCollection(ctx, gameDataByDay.dbCollection, gameDataByDay.firebaseDomain, gameDataByDay.projectID)
	if err != nil {
		gameDataByDay.handleFatalError("error setting up connection to FireStore", err)
	}

	date, err := parseDate(r.Body, gameDataByDay.dateFmt)
	if err != nil {
		gameDataByDay.handleFatalError("error parsing date requested", err)
	}

	daySchedule, err := mlbStats.GetSchedule(date, gameDataByDay.logger)
	if err != nil {
		gameDataByDay.handleFatalError("error getting the daily StatsAPI schedule", err)
	}

	_, err = collection.Doc(date.Format(gameDataByDay.dateFmt)).Set(ctx, daySchedule)
	if err != nil {
		gameDataByDay.handleFatalError("error persisting data to Firebase", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule)
}

// FireStoreCollection sets up a connetion to Firebase and fetches a connection to the desired FireStore collection
func fireStoreCollection(ctx context.Context, databaseCollection, firebaseDomain, projectID string) (*firestore.CollectionRef, error) {
	conf := &firebase.Config{DatabaseURL: fmt.Sprintf("https://%s.%s", projectID, firebaseDomain)}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return nil, err
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}
	return fsClient.Collection(databaseCollection), nil
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

func (g gameDataByDay) handleFatalError(msg string, err error) {
	g.errorReporter.Report(errorreporting.Entry{})
	g.logger.Log(logging.Entry{Severity: logging.Error, Payload: logMessage{msg: msg, err: err.Error()}})
	log.Fatalf("%s: %s", msg, err)
}
