package function

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
)

// GetGameDataByDay returns useful (to Warning-Track) game information for given date
//
// ex.: https://us-central1-warning-track-backend.cloudfunctions.net/GetGameDataByDay -d {'"date":"03-01-2020"'}
//
func GetGameDataByDay(w http.ResponseWriter, r *http.Request) {
	// setup Logger
	ctx := context.Background()
	logName := "get-game-data-by-day"
	projectID := "warning-track-backend"
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Error setuping logger")
		return
	}
	defer client.Close()
	logger := client.Logger(logName).StandardLogger(logging.Info)

	logger.Printf("Received request: %+v", r.Body)

	var d struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		logger.Printf("Error attempting to decode json body: %s", err)
		return
	}
	logger.Printf("Date requested: %+v", d.Date)

	parsedDate, err := time.Parse("01-02-2006", d.Date)
	if err != nil {
		logger.Printf("Error parsing date requested: %s", err)
	}

	URL := statsAPIScheduleURL(parsedDate)
	logger.Printf("Making Get request: %s", URL)
	resp, err := http.Get(URL)
	if err != nil {
		logger.Printf("Error in Get request: %s", err)
		return
	}
	defer resp.Body.Close()

	logger.Println("Parsing response from Get request")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Error reading Get response body: %s", err)
		return
	}

	logger.Println("successfully received response from Get")

	statsAPIScheduleResp := statsAPISchedule{}
	err = json.Unmarshal(body, &statsAPIScheduleResp)
	if err != nil {
		logger.Printf("Error trying to unmarshal response from statsAPI: %s", err)
		return
	}

	// Integrate with Firebase to persist data
	httpClient := &http.Client{}
	fbPUTUrl := firebaseURL(parsedDate)
	b, err := json.Marshal(statsAPIScheduleResp)
	if err != nil {
		logger.Printf("Error trying to marshal response from statsAPI: %s", err)
		return
	}
	data := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fbPUTUrl, data)
	if err != nil {
		logger.Printf("Error preparing PUT request to Firebase: %s", err)
	}
	_, err = httpClient.Do(req)
	if err != nil {
		logger.Printf("Error trying to make PUT to %s: %s", fbPUTUrl, err)
	}

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
	return "https://" + firebaseNamespace + ".firebaseio.com/" + databaseCollection + "/" + month + "/" + day + "/" + year + ".json"
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
