package MatchFinder

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

const (
	TwoUp      QueryType = iota
	Qualifying QueryType = iota
	FreeBet    QueryType = iota
	ValueBet   QueryType = iota
)

type MatchError struct {
	err string
}

func (e MatchError) Error() string {
	return fmt.Sprintf("[Match API Error] %s", e.err)
}

const url = "https://api.oddsplatform.profitaccumulator.com/graphql"

func Find(queryType QueryType, stake float64) ([]Match, error) {
	switch queryType {
	case TwoUp:
		resp, err := Get2UpData(stake)
		if err != nil {
			return nil, MatchError{fmt.Sprintf("Error getting 2UP Data: %s", err.Error())}
		}
		return resp, nil
	}
	return nil, MatchError{fmt.Sprintf("Invalid query type")}
}

func Get2UpData(stake float64) ([]Match, error) {
	var queryResponse MatchResponse
	MinRating := "96"
	MinOdds := "2"
	MaxOdds := "6"
	MaxRating := "100"

	query := Query{
		RatingType:   "rating",
		MinOdds:      &MinOdds,
		MaxOdds:      &MaxOdds,
		MinRating:    &MinRating,
		MaxRating:    &MaxRating,
		Cap:          100,
		Limit:        500,
		Skip:         0,
		ExcludeDraws: true,
		Bookmakers:   []string{"bet365"},
		Exchanges:    []string{"smarketsexchange"},
		LastUpdate:   21600,
		MarketGroups: []string{"match-odds"},
		Sports:       []string{"soccer"},
		EventGroups:  []string{"english-premier-league", "uefa-champions-league", "uefa-europa-league", "german-bundesliga", "italian-serie-a", "spanish-la-liga", "french-ligue-1", "english-fa-cup", "scottish-premiership"},
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

	for i := range queryResponse.Data.Matches {
		rating, pfErr := strconv.ParseFloat(queryResponse.Data.Matches[i].Rating, 32)
		if pfErr != nil {
			return nil, pfErr
		}
		queryResponse.Data.Matches[i].QualLoss = float32((stake * (rating - 100)) / 100)
	}

	return queryResponse.Data.Matches, nil
}

func makeQuery(q Query) (json.RawMessage, error) {
	queryFile, err := os.Open("MatchFinder/getBestMatches.graphql")
	if err != nil {
		return nil, err
	}
	defer func(queryFile *os.File) {
		err := queryFile.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(queryFile)

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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(fQData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing Body:", err)
		}
	}(resp.Body)

	var response io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer func(gzipReader *gzip.Reader) {
			err := gzipReader.Close()
			if err != nil {
				fmt.Printf("failed to close reader: %v\n", err)
			}
		}(gzipReader)
		response = gzipReader
	}

	body, err := io.ReadAll(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzipped response body: %v", err)
	}

	return body, nil
}
