package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	firebase "firebase.google.com/go"
)

const (
	databaseCollection = "game-data-by-day"
	dateFormat         = "01-02-2006"
	duration           = 60 * time.Second
	firebaseDomain     = "firebaseio.com"
	logName            = "get-game-data-by-day"
	projectID          = "warning-track-backend"
)

// LogMessage is a simple struct to ensure JSON formatting in logs
type LogMessage struct {
	Message string
}

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
//
// ex.: https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {'"date":"03-01-2020"'}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	defer cancel()

	lg, err := cloudLogger(ctx, projectID, logName)
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger")
	}

	collection, err := fireStoreCollection(ctx, databaseCollection, firebaseDomain, projectID, lg)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error setting up connection to FireStore: %s", err)}})
		return
	}

	date, err := parseDate(r.Body, dateFormat, lg)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error parsing date requested: %s", err)}})
		return
	}

	URL := statsAPIScheduleURL(date)
	lg.Log(logging.Entry{Severity: logging.Debug, Payload: LogMessage{Message: fmt.Sprintf("making Get request: %s", URL)}})
	resp, err := http.Get(URL)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error in Get request: %s", err)}})
		return
	}
	defer resp.Body.Close()

	lg.Log(logging.Entry{Severity: logging.Debug, Payload: LogMessage{Message: "parsing response from Get request"}})
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error reading Get response body: %s", err)}})
		return
	}

	lg.Log(logging.Entry{Severity: logging.Debug, Payload: LogMessage{Message: "successfully received response from Get"}})

	statsAPIScheduleResp := statsAPISchedule{}
	err = json.Unmarshal(body, &statsAPIScheduleResp)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error trying to unmarshal response from statsAPI: %s", err)}})
		return
	}

	// Integrate with FireStore to persist data
	_, err = collection.Doc(date.Format(dateFormat)).Set(ctx, statsAPIScheduleResp)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error trying to set value in Firebase: %s", err)}})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statsAPIScheduleResp)
}

// statsAPIScheduleURL returns the URL for all the game schedule data for the given time
func statsAPIScheduleURL(time time.Time) string {
	host := "http://statsapi.mlb.com"
	path := "/api/v1/schedule"
	query := "?sportId=1&hydrate=game(content(summary,media(epg))),linescore(runners),flags,team&date="
	month := time.Format("01")
	day := time.Format("02")
	year := time.Format("2006")
	return host + path + query + month + "/" + day + "/" + year
}

// cloudLogger sets up a connection to Google Cloud Logging for the funciton
func cloudLogger(ctx context.Context, projectID, logName string) (*logging.Logger, error) {
	client, err := logging.NewClient(ctx, projectID)
	defer client.Close()
	return client.Logger(logName), err
}

// fireStoreCollection sets up a connetion to Firebase and fetches a connection to the desired FireStore collection
func fireStoreCollection(ctx context.Context, databaseCollection, firebaseDomain, projectID string, lg *logging.Logger) (*firestore.CollectionRef, error) {
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

// parseDate parses the request body and returns a time.Time value of the requested date
func parseDate(reqBody io.ReadCloser, dateFormat string, lg *logging.Logger) (time.Time, error) {
	var d struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(reqBody).Decode(&d); err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error attempting to decode json body: %s", err)}})
		return time.Time{}, err
	}
	lg.Log(logging.Entry{Severity: logging.Debug, Payload: LogMessage{Message: fmt.Sprintf("date requested: %+v", d.Date)}})

	return time.Parse(dateFormat, d.Date)
}
