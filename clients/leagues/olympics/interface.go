package olympics

// GenderCode for WMR API: M = men's, W = women's
const (
	GenderMen   = "M"
	GenderWomen = "W"
)

// IOlympicsApiClient defines the Olympics WMR API client interface
type IOlympicsApiClient interface {
	GetStartList(genderCode string) (StartListResponse, error)
	GetScores(genderCode string) (ScoresResponse, error)
	GetScoresForGame(genderCode, phaseCode, eventUnitCode string) (ScoresResponse, error)
	GetPlayByPlay(genderCode, phaseCode, eventUnitCode string) (PlayByPlayResponse, error)
	// GetScheduleLiteByDay returns raw JSON for wmr-owg2026 schedule by day (date YYYY-MM-DD)
	GetScheduleLiteByDay(date string) ([]byte, error)
}
