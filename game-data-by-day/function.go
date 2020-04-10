package function

import (
	"bytes"
	"context"
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
	firebaseAccountKeyPath = ""
	logName                = "get-game-data-by-day"
	projectID              = "warning-track-backend"
)

// LogMessage is a simple struct to ensure JSON formatting in logs
type LogMessage struct {
	Message string
}

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
//
// ex.: https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {'"date":"03-01-2020"'}
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	// setup logger
	ctx := context.Background()
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("error setting up Google Cloud logger")
	}
	defer client.Close()
	lg := client.Logger(logName)
	var d struct {
		Date string `json:"date"`
	}

	// setup Firebase
	conf := &firebase.Config{
		DatabaseURL: "https://warning-track-backend.firebaseio.com",
	}

	// TODO: https://github.com/firebase/firebase-admin-go/blob/master/snippets/db.go
	_, err = firebase.NewApp(ctx, conf)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error initializing Firebase app: %s", err)}})
		return
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

	// Integrate with Firebase to persist data
	httpClient := &http.Client{}
	fbPUTUrl := firebaseURL(parsedDate)
	b, err := json.Marshal(statsAPIScheduleResp)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error trying to marshal response from statsAPI: %s", err)}})
		return
	}
	data := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fbPUTUrl, data)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error preparing PUT request to Firebase: %s", err)}})
		return
	}

	lg.Log(logging.Entry{Severity: logging.Debug, Payload: LogMessage{Message: fmt.Sprintf("making Put request to firebase: %s", fbPUTUrl)}})
	fbResp, err := httpClient.Do(req)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error trying to make PUT to %s: %s", fbPUTUrl, err)}})
		return
	}
	lg.Log(logging.Entry{Severity: logging.Debug, Payload: LogMessage{Message: fmt.Sprintf("Put response from firebase: %+v", fbResp)}})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statsAPIScheduleResp)
}

// firebaseURL returns the URL for firebase for this project
func firebaseURL(time time.Time) string {
	firebaseNamespace := "warning-track-backend"
	databaseCollection := "game-data-by-day"
	month := time.Format("01")
	day := time.Format("02")
	year := time.Format("2006")
	return "https://" + firebaseNamespace + ".firebaseio.com/" + databaseCollection + "/" + month + "-" + day + "-" + year + ".json"
}

// getReference
func getReference(ctx context.Context, app *firebase.App) {
	// Create a database client from App.
	client, err := app.Database(ctx)
	if err != nil {
		log.Fatalln("Error initializing database client:", err)
	}

	// Get a database reference to our blog.
	ref := client.NewRef("server/saving-data/fireblog")
	// [END get_reference]
	fmt.Println(ref.Path)
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
