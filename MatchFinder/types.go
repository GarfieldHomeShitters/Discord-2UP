package MatchFinder

type MatchResponse struct {
	Data MatchList `json:"data"`
}

type MatchList struct {
	Matches []Match `json:"getBestMatches"`
}

type Match struct {
	Back          Back   `json:"back"`
	Lay           Lay    `json:"lay"`
	EventName     string `json:"eventName"`
	SelectionName string `json:"selectionName"`
	Rating        string `json:"rating"`
}

type Back struct {
	Odds string `json:"odds"`
}

type Lay struct {
	Odds      string `json:"odds"`
	Liquidity string `json:"liquidity"`
}

type Query struct {
	Bookmakers       []string        `json:"bookmaker"`
	Exchanges        []string        `json:"exchange"`
	MinRating        *string         `json:"minRating"`
	MaxRating        *string         `json:"maxRating"`
	TimeframeStart   *string         `json:"timeframeStart"`
	TimeframeEnd     *string         `json:"timeframeEnd"`
	EventNameFilter  *string         `json:"searchByEventName"`
	Limit            int             `json:"limit"`
	Cap              int             `json:"cap"`
	LastUpdate       int             `json:"updatedWithinSeconds"`
	ExcludeDraws     bool            `json:"excludeDraw"`
	MinimumLiquidity *string         `json:"minLiquidity"`
	RatingType       string          `json:"ratingType"`
	MinOdds          *string         `json:"minOdds"`
	MaxOdds          *string         `json:"maxOdds"`
	MarketGroups     []string        `json:"permittedMarketGroups"`
	Sports           []string        `json:"permittedSports"`
	Skip             int             `json:"skip"`
	EventGroups      []string        `json:"permittedEventGroups"`
	CommissionRates  []CommisionRate `json:"commissionRates"`
}

type CommisionRate struct {
	Exchange string `json:"exchange"`
	Rate     int    `json:"rate"`
}

type FullQuery struct {
	OperationName string `json:"operationName"`
	Variables     Query  `json:"variables"`
	Query         string `json:"query"`
}

type QueryType int
