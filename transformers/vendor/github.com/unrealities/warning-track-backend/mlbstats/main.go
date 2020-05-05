package mlbstats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// GetSchedule returns a Schedule that contains all the requested day's games
func GetSchedule(date time.Time) (Schedule, error) {
	URL := statsAPIScheduleURL(date)
	resp, err := http.Get(URL)
	if err != nil {
		return Schedule{}, fmt.Errorf("mlbStats#GetSchedule: Get %s, error: %s", URL, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Schedule{}, fmt.Errorf("mlbStats#GetSchedule: reading Get response body error: %s", err)
	}

	statsAPIScheduleResp := Schedule{}
	err = json.Unmarshal(body, &statsAPIScheduleResp)
	if err != nil {
		return Schedule{}, fmt.Errorf("mlbStats#GetSchedule: unmarshal error: %s", err)
	}

	return statsAPIScheduleResp, nil
}
