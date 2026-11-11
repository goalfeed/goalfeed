package nhl

import (
	"encoding/json"
	"errors"
	"fmt"
	"goalfeed/utils"
)

type NHLApiClient struct {
}

// fetchers for testability
var fetchByte = utils.GetByte

func (c NHLApiClient) GetNHLScoreBoard(gameId string) NHLScoreboardResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://api-web.nhle.com/v1/gamecenter/%s/landing", gameId) // Updated URL
	go fetchByte(url, body)

	bodyByte := <-body
	var response NHLScoreboardResponse
	if err := json.Unmarshal(bodyByte, &response); err != nil && len(bodyByte) > 0 {
		// Unmarshal fills every field it can before reporting a mismatch, so
		// the partial response is still used; the error is only logged.
		utils.GetLogger().Warn(fmt.Sprintf("NHL landing %s: decode: %v", gameId, err))
	}
	return response
}

func (c NHLApiClient) GetNHLSchedule() NHLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := "https://api-web.nhle.com/v1/schedule/now" // Updated URL
	go fetchByte(url, body)

	bodyByte := <-body
	var response NHLScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetNHLScheduleByDate(date string) NHLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://api-web.nhle.com/v1/schedule/%s", date) // Format: YYYY-MM-DD
	go fetchByte(url, body)

	bodyByte := <-body
	var response NHLScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetTeam(teamAbbr string) NHLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://api-web.nhle.com/v1/club-schedule-season/%s/now", teamAbbr) // Updated URL
	go fetchByte(url, body)

	bodyByte := <-body
	var response NHLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetDiffPatch(gameId string, timestamp string) (NHLDiffPatch, error) {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://api-web.nhle.com/v1/game/%s/feed/live/diffPatch?site=en_nhl&startTimecode=%s", gameId, timestamp) // Updated URL
	go fetchByte(url, body)

	bodyByte := <-body
	var response NHLDiffPatch
	err := json.Unmarshal(bodyByte, &response)
	return response, err
}

func (c NHLApiClient) GetAllTeams() NHLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := "https://api-web.nhle.com/v1/teams"
	go fetchByte(url, body)

	bodyByte := <-body
	var response NHLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

// GetNHLRightRail fetches /v1/gamecenter/{id}/right-rail, which carries the
// team game stats panel (hits, faceoffs, power play, PIM, blocks, giveaways,
// takeaways), shots by period and the linescore.
//
// A type mismatch in one field does not discard the rest of the payload:
// json.Unmarshal fills every field it can before reporting the mismatch, so
// the partial response is returned along with the error for the caller to
// log. Only an empty body is treated as no data at all.
func (c NHLApiClient) GetNHLRightRail(gameId string) (NHLRightRailResponse, error) {
	body := make(chan []byte)
	url := fmt.Sprintf("https://api-web.nhle.com/v1/gamecenter/%s/right-rail", gameId)
	go fetchByte(url, body)

	bodyByte := <-body
	var response NHLRightRailResponse
	if len(bodyByte) == 0 {
		return response, errors.New("nhl right-rail: empty response")
	}
	if err := json.Unmarshal(bodyByte, &response); err != nil {
		return response, fmt.Errorf("nhl right-rail %s: %w", gameId, err)
	}
	return response, nil
}
