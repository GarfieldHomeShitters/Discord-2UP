package MatchFinder

import (
	"Adam/discord-twoup/MatchFinder"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	TwoUp      QueryType = iota
	Qualifying QueryType = iota
	FreeBet    QueryType = iota
	ValueBet   QueryType = iota
)

const url = "https://api.oddsplatform.profitaccumulator.com/graphql"

func Find(queryType QueryType) []Match {
	switch queryType {
	case TwoUp:
		resp, err := Get2UpData()
		if err != nil {
			fmt.Println("Error getting 2UP Data:", err)
			return nil
		}
		return resp
	}
	return nil
}

func Get2UpData() ([]Match, error) {
	var queryResponse MatchResponse
	MinRating := "90"
	MinOdds := "2"
	MaxRating := "100"

	query := Query{
		RatingType:   "rating",
		MinOdds:      &MinOdds,
		MinRating:    &MinRating,
		MaxRating:    &MaxRating,
		Cap:          100,
		Limit:        500,
		ExcludeDraws: true,
		Bookmakers:   []string{"bet365"},
		Exchanges:    []string{"smarketsexchange"},
		LastUpdate:   21600,
		MarketGroups: []string{"match-odds"},
		Sports:       []string{"soccer"},
		CommissionRates: []CommisionRate{
			{Exchange: "betdaq", Rate: 0},
			{Exchange: "betfair", Rate: 5},
			{Exchange: "matchbook", Rate: 0},
			{Exchange: "smarkets", Rate: 0},
			{Exchange: "pocketbet", Rate: 0},
		},
	}

	response, err := makeQuery(query)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response, &queryResponse)
	if err != nil {
		return nil, err
	}

	return queryResponse.Matches, nil
}

func makeQuery(q Query) (json.RawMessage, error) {
	queryFile, err := os.Open("getBestMatches.graphql")
	if err != nil {
		return nil, err
	}
	defer queryFile.Close()

	queryBytes, err := io.ReadAll(queryFile)
	if err != nil {
		return nil, err
	}
	queryString := string(queryBytes)

	fQ := FullQuery{
		Query:         queryString,
		Variables:     q,
		OperationName: "GetBestMatches",
	}

	fQData, err := json.Marshal(fQ)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(fQData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
