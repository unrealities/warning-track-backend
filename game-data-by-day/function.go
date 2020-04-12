package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
	firebase "firebase.google.com/go"
)

const (
	databaseCollection = "game-data-by-day"
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
	ctx := r.Context()

	// setup logger
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger")
	}
	defer client.Close()
	lg := client.Logger(logName)

	// setup Firebase and Firestore
	conf := &firebase.Config{DatabaseURL: fmt.Sprintf("https://%s.firebaseio.com", projectID)}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error initializing Firebase app: %s", err)}})
		return
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error initializing FireStore client: %s", err)}})
		return
	}
	collection := fsClient.Collection(databaseCollection)

	var d struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error attempting to decode json body: %s", err)}})
		return
	}
	lg.Log(logging.Entry{Severity: logging.Debug, Payload: LogMessage{Message: fmt.Sprintf("date requested: %+v", d.Date)}})

	parsedDate, err := time.Parse("01-02-2006", d.Date)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error parsing date requested: %s", err)}})
		return
	}

	URL := statsAPIScheduleURL(parsedDate)
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
	_, err = collection.Doc(d.Date).Set(ctx, statsAPIScheduleResp)
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
