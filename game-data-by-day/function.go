package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"

	"github.com/unrealities/warning-track-backend/gCloud"
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
	log.Printf("received a request: %+v", r)

	gameDataByDay := gameDataByDay{
		dateFmt:        "01-02-2006",
		dbCollection:   "game-data-by-day",
		duration:       60 * time.Second,
		firebaseDomain: "firebaseio.com",
		projectID:      "warning-track-backend",
		version:        "v0.0.31",
	}

	log.Printf("gameDataByDay: %+v", gameDataByDay)

	log.Printf("running version: %s", gameDataByDay.version)

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: gameDataByDay.projectID})
	if err != nil {
		log.Fatalf("error setting up OpenCensus Stackdriver Trace exporter")
	}
	trace.RegisterExporter(exporter)

	log.Printf("successfully registered trace exporter: %+v", exporter)

	ctx, cancel := context.WithTimeout(r.Context(), gameDataByDay.duration)
	defer cancel()

	log.Printf("successfully set context: %+v", ctx)

	lg, err := gCloud.CloudLogger(ctx, gameDataByDay.projectID, fmt.Sprintf("get-%s", gameDataByDay.dbCollection))
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger")
	}

	log.Printf("succesffuly connected to Google Cloud Logger: %+v", lg)

	gameDataByDay.logger = lg

	log.Printf("successfully registered logger in gameDataByDay: %+v", gameDataByDay)

	collection, err := gCloud.FireStoreCollection(ctx, gameDataByDay.dbCollection, gameDataByDay.firebaseDomain, gameDataByDay.projectID, gameDataByDay.logger)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gCloud.LogMessage{Message: fmt.Sprintf("error setting up connection to FireStore: %s", err)}})
		return
	}

	log.Printf("successfully fetched collection: %+v", collection)

	date, err := parseDate(r.Body, gameDataByDay.dateFmt)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gCloud.LogMessage{Message: fmt.Sprintf("error parsing date requested: %s", err)}})
		return
	}

	log.Printf("successfully parsed date: %+v", date)

	daySchedule, err := mlbStats.GetSchedule(date, gameDataByDay.logger)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gCloud.LogMessage{Message: fmt.Sprintf("error getting the daily StatsAPI schedule: %s", err)}})
		return
	}

	log.Println("successfully fetched daySchedule")

	_, err = collection.Doc(date.Format(gameDataByDay.dateFmt)).Set(ctx, daySchedule)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: gCloud.LogMessage{Message: fmt.Sprintf("error trying to set value in Firebase: %s", err)}})
		return
	}

	log.Println("successfully persisted data")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daySchedule)
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
