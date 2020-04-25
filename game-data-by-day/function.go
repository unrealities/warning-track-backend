package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

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
	firebaseDomain string
	logger         *logging.Logger
	projectID      string
	version        string
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
		version:        "v0.0.38",
	}
	log.Printf("running version: %s", gameDataByDay.version)

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: gameDataByDay.projectID})
	if err != nil {
		log.Fatalf("error setting up OpenCensus Stackdriver Trace exporter")
	}
	trace.RegisterExporter(exporter)

	ctx, cancel := context.WithTimeout(r.Context(), gameDataByDay.duration)
	defer cancel()

	lg, err := CloudLogger(ctx, gameDataByDay.projectID, fmt.Sprintf("get-%s", gameDataByDay.dbCollection))
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger")
	}
	gameDataByDay.logger = lg

	collection, err := FireStoreCollection(ctx, gameDataByDay.dbCollection, gameDataByDay.firebaseDomain, gameDataByDay.projectID, gameDataByDay.logger)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error setting up connection to FireStore: %s", err)}})
		return
	}

	date, err := parseDate(r.Body, gameDataByDay.dateFmt)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error parsing date requested: %s", err)}})
		return
	}

	daySchedule, err := mlbStats.GetSchedule(date, gameDataByDay.logger)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error getting the daily StatsAPI schedule: %s", err)}})
		return
	}

	log.Printf("collection: %+v", collection)
	doc := collection.Doc(date.Format(gameDataByDay.dateFmt))
	log.Printf("doc: %+v", doc)

	// TODO: Problem here
	_, err = doc.Set(ctx, daySchedule)
	log.Printf("received response from setting the collection")
	if err != nil {
		log.Fatalf("error persisting data to firestore: %s", err)
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error trying to set value in Firebase: %s", err)}})
		return
	}

	log.Println("successfully persisted data")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule.TotalGames)
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

// taken from gCloud

// LogMessage is a simple struct to ensure JSON formatting in logs
type LogMessage struct {
	Message string
}

// CloudLogger sets up a connection to Google Cloud Logging for the funciton
func CloudLogger(ctx context.Context, projectID, logName string) (*logging.Logger, error) {
	client, err := logging.NewClient(ctx, projectID)
	defer client.Close()
	return client.Logger(logName), err
}

// FireStoreCollection sets up a connetion to Firebase and fetches a connection to the desired FireStore collection
func FireStoreCollection(ctx context.Context, databaseCollection, firebaseDomain, projectID string, lg *logging.Logger) (*firestore.CollectionRef, error) {
	conf := &firebase.Config{DatabaseURL: fmt.Sprintf("https://%s.%s", projectID, firebaseDomain)}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error initializing Firebase app: %s", err)}})
		return nil, err
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error initializing FireStore client: %s", err)}})
		return nil, err
	}
	return fsClient.Collection(databaseCollection), nil
}
