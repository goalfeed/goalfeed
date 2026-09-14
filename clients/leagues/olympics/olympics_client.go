package olympics

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"goalfeed/config"
	"goalfeed/utils"
)

// Two API bases are used by the Olympics site (from HAR / schedule navigation):
// 1. wmr-api (v2): game-specific data - startList, playByPlay, results/scores.
//    Base: https://www.olympics.com/wmr-api/api/v2
//    Params: competitionCode=OWG2026, disciplineCode=IHO, eventCode=TEAM6-------------,
//    genderCode=M|W, phaseCode, eventUnitCode, languageCode=ENG.
// 2. wmr-owg2026: schedule lite by day (all sports).
//    GET https://www.olympics.com/wmr-owg2026/schedules/api/ENG/schedule/lite/day/YYYY-MM-DD
//    Response includes ODF document codes (odf-documentcode header).
//
// IHO event IDs from HAR (odf-documentcode values):
//   Women's: IHOWTEAM6-------------QFNL000300--  (phaseCode=QFNL, eventUnitCode=000300--)
//            Page: /results/iho/w/team6-------------/qfnl/000300--/team-lineups
//   Men's:   IHOMTEAM6-------------GPB-000600--  (phaseCode=GPB-, eventUnitCode=000600--)
//
// Live-game HAR (men's and women's): same wmr-api v2 base. Women's live load (e.g. live-game-load-wmn.har)
// uses genderCode=W with same endpoints; single game e.g. phaseCode=QFNL, eventUnitCode=000300--.
// Endpoints used during a live game:
//   - results/scores (single game: phaseCode+eventUnitCode; or phaseCode=*, eventUnitCode=* for all)
//   - playByPlay (phaseCode+eventUnitCode)
//   - events (no phase/unit), events/extended (genderCode; optional phaseCode+eventUnitCode for one game)
//   - results (full result for one game), schedules (game schedule), schedules/startList (with or without phase+unit)
//   - pdf (game PDF). API returns cache-control: max-age=5 and ETag; client can send If-None-Match for 304.
//   startList can be called with optional phaseCode+eventUnitCode for one game (live HAR used eventUnitCode=000500 no trailing --).
//   live-game-load HAR: initial page load returns 200 with body; playByPlay and results/scores use wrapper { ResponseCode, TotalRows, Data }.
const (
	baseURL         = "https://www.olympics.com/wmr-api/api/v2"
	scheduleLiteURL = "https://www.olympics.com/wmr-owg2026/schedules/api/ENG/schedule/lite/day"
	competitionCode = "OWG2026"
	disciplineCode  = "IHO"
	eventCode       = "TEAM6-------------"
	languageCode    = "ENG"
)

// OlympicsApiClient calls the Olympics.com WMR API for ice hockey
type OlympicsApiClient struct {
}

func (c *OlympicsApiClient) buildURL(path string, params map[string]string) string {
	v := url.Values{}
	v.Set("competitionCode", competitionCode)
	v.Set("disciplineCode", disciplineCode)
	v.Set("eventCode", eventCode)
	v.Set("languageCode", languageCode)
	for k, val := range params {
		v.Set(k, val)
	}
	return fmt.Sprintf("%s/%s?%s", baseURL, path, v.Encode())
}

func (c *OlympicsApiClient) doRequest(apiURL string) ([]byte, error) {
	headers := map[string]string{
		"Accept":     "application/json",
		"User-Agent": "Goalfeed/1.0",
	}
	if token := config.GetString("olympics.bearer_token"); token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	bodyChan := make(chan []byte)
	go utils.GetByteWithHeaders(apiURL, bodyChan, headers)
	body := <-bodyChan
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response from Olympics API")
	}
	return body, nil
}

// GetStartList returns the start list (schedule of games) for the given gender
func (c *OlympicsApiClient) GetStartList(genderCode string) (StartListResponse, error) {
	apiURL := c.buildURL("schedules/startList", map[string]string{"genderCode": genderCode})
	body, err := c.doRequest(apiURL)
	if err != nil {
		return nil, err
	}
	var list StartListResponse
	if err := json.Unmarshal(body, &list); err == nil && len(list) > 0 {
		return list, nil
	}
	// Try wrapper object e.g. { "startList": [...] } or { "data": [...] }
	var wrapper struct {
		StartList StartListResponse `json:"startList"`
		Data      StartListResponse `json:"data"`
		Items     StartListResponse `json:"items"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("parse startList: %w", err)
	}
	if len(wrapper.StartList) > 0 {
		return wrapper.StartList, nil
	}
	if len(wrapper.Data) > 0 {
		return wrapper.Data, nil
	}
	if len(wrapper.Items) > 0 {
		return wrapper.Items, nil
	}
	return list, nil
}

// GetScores returns all scores for the discipline/gender (phaseCode=*, eventUnitCode=*)
func (c *OlympicsApiClient) GetScores(genderCode string) (ScoresResponse, error) {
	apiURL := c.buildURL("results/scores", map[string]string{
		"genderCode":    genderCode,
		"phaseCode":     "*",
		"eventUnitCode": "*",
	})
	body, err := c.doRequest(apiURL)
	if err != nil {
		return ScoresResponse{}, err
	}
	var out ScoresResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return ScoresResponse{}, fmt.Errorf("parse scores: %w", err)
	}
	return out, nil
}

// GetScoresForGame returns score for a single game (phaseCode + eventUnitCode)
func (c *OlympicsApiClient) GetScoresForGame(genderCode, phaseCode, eventUnitCode string) (ScoresResponse, error) {
	apiURL := c.buildURL("results/scores", map[string]string{
		"genderCode":    genderCode,
		"phaseCode":     phaseCode,
		"eventUnitCode": eventUnitCode,
	})
	body, err := c.doRequest(apiURL)
	if err != nil {
		return ScoresResponse{}, err
	}
	// Live-game-load HAR: API returns wrapper { ResponseCode, TotalRows, Data: [{ PhaseCode, EventUnitCode, Home, Away, Scorers: [{ Goals: [{ Order, ScoreH, ScoreA }] }] }] }
	var wrapper apiResponseWrapper
	if err := json.Unmarshal(body, &wrapper); err == nil && wrapper.ResponseCode == 200 && len(wrapper.Data) > 0 {
		var dataList []scoresDataEntry
		if err := json.Unmarshal(wrapper.Data, &dataList); err == nil && len(dataList) > 0 {
			d := &dataList[0]
			home, away := scoreFromScoresDataEntry(d)
			return ScoresResponse{
				PhaseCode:  d.PhaseCode,
				UnitCode:   d.EventUnitCode,
				HomeScore:  home,
				AwayScore:  away,
			}, nil
		}
	}
	var out ScoresResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return ScoresResponse{}, fmt.Errorf("parse scores: %w", err)
	}
	return out, nil
}

// scoreFromScoresDataEntry returns home/away score from the last goal (max Order) in Scorers
func scoreFromScoresDataEntry(d *scoresDataEntry) (home, away int) {
	var maxOrder int
	var lastH, lastA string
	for _, s := range d.Scorers {
		for _, g := range s.Goals {
			if g.Order >= maxOrder {
				maxOrder = g.Order
				lastH, lastA = g.ScoreH, g.ScoreA
			}
		}
	}
	home, _ = strconv.Atoi(lastH)
	away, _ = strconv.Atoi(lastA)
	return home, away
}

// GetPlayByPlay returns play-by-play for a single game
func (c *OlympicsApiClient) GetPlayByPlay(genderCode, phaseCode, eventUnitCode string) (PlayByPlayResponse, error) {
	apiURL := c.buildURL("playByPlay", map[string]string{
		"genderCode":    genderCode,
		"phaseCode":     phaseCode,
		"eventUnitCode": eventUnitCode,
	})
	body, err := c.doRequest(apiURL)
	if err != nil {
		return PlayByPlayResponse{}, err
	}
	// Live-game-load HAR: API returns wrapper { ResponseCode, TotalRows, Data: [{ ActionCode, When, ScoreHome, ScoreAway, ... }] }
	var wrapper apiResponseWrapper
	if err := json.Unmarshal(body, &wrapper); err == nil && wrapper.ResponseCode == 200 && len(wrapper.Data) > 0 {
		var dataList []playByPlayDataEntry
		if err := json.Unmarshal(wrapper.Data, &dataList); err == nil && len(dataList) > 0 {
			events := make([]PlayByPlayEntry, 0, len(dataList))
			for _, e := range dataList {
				entry := PlayByPlayEntry{
					Sequence:   e.ActionOrder,
					Time:       e.When,
					Type:       e.ActionCode,
					Description: e.ActionCode,
					IsGoal:     e.ActionCode == "GOAL",
				}
				if e.PeriodCode != "" {
					entry.Period = periodFromCode(e.PeriodCode)
				}
				if e.ActionCompetitors != nil {
					entry.TeamCode = e.ActionCompetitors.OrganisationCode
				}
				events = append(events, entry)
			}
			return PlayByPlayResponse{Events: events}, nil
		}
	}
	var out PlayByPlayResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return PlayByPlayResponse{}, fmt.Errorf("parse playByPlay: %w", err)
	}
	return out, nil
}

func periodFromCode(code string) int {
	if len(code) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(code[1:])
	return n
}

// GetScheduleLiteByDay fetches the day schedule from wmr-owg2026 (all sports).
// Returns raw JSON; the response structure is large and includes multiple disciplines.
// Date format: YYYY-MM-DD (e.g. "2026-02-14").
func (c *OlympicsApiClient) GetScheduleLiteByDay(date string) ([]byte, error) {
	apiURL := fmt.Sprintf("%s/%s", scheduleLiteURL, date)
	return c.doRequest(apiURL)
}
