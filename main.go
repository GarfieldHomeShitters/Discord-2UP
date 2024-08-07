package main

import (
	"Adam/discord-twoup/Database"
	"Adam/discord-twoup/Handlers"
	"Adam/discord-twoup/MatchFinder"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/oracle/nosql-go-sdk/nosqldb/common"
	"log"
	"time"
)

const url = "https://canary.discord.com/api/webhooks/1267996258203865201/ZenhYVAIrjy3YLLkqtAjSIEGJ3W6UW2Qn627G41IrcxiaAZrSTGjegYR9zDbwfUkmD4v"

type NotificationHandler struct {
	Handlers.MatchNotifier
	Handlers.ErrorHandler
}

type DatabaseHandler struct {
	Database.Database
}

type DbItem struct {
	ID             string `json:"id"`
	EventStartDate string `json:"EventStartDate"`
}

func main() {
	Notifier := CreateNotificationHandler()
	Db := CreateDatabaseHandler()

	dbError := Db.Connect()
	if dbError != nil {
		nErr := Notifier.LogTypedError(dbError)
		if nErr != nil {
			str := fmt.Sprintf("Error logging error: %v\n", dbError)
			log.Fatal(str)
		}
		return
	}
	defer Db.Close()

	typedErr := Db.SelectTable("PreviouslyNotified")
	if typedErr != nil {
		Notifier.LogTypedError(typedErr)
		panic(typedErr.Error())
	}

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	if err := PerformMatchTask(Notifier, Db); err != nil {
		Notifier.LogTypedError(err)
		panic(err.Error())
	}

	for {
		select {
		case <-ticker.C:
			if err := PerformMatchTask(Notifier, Db); err != nil {
				Notifier.LogTypedError(err)
				return
			}
		}
	}
}

func PerformMatchTask(Notifier *NotificationHandler, Db *DatabaseHandler) Handlers.TypedError {
	stake := 20.00
	qR, err := MatchFinder.Find(MatchFinder.TwoUp, stake)
	if err != nil {
		notificationErr := Notifier.LogError(err)
		if notificationErr != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Panic(errStr)
		}
	}
	// TODO: Later -> handle no new matches and softly inform user -> log in console instead of via discord.
	getErr, newMatches := filterMatches(qR, Db)
	if getErr != nil {
		notificationErr := Notifier.LogTypedError(getErr)
		if notificationErr != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Panic(errStr)
		}
		log.Panic(getErr.Error())
	}

	err = Notifier.NotifyUser(context.Background(), &newMatches, stake)
	if err != nil && err.Error() != "no_matches" {
		fmt.Println("Failed Webhook:", err)
		notificationErr := Notifier.LogError(err)
		if notificationErr != nil {
			errStr := fmt.Sprintf("Error logging error: %v\n", err)
			log.Panic(errStr)
		}
		log.Panic(err.Error())
	}

	for _, v := range newMatches {
		item := NewDatabaseItem(v)
		typedErr := Db.Put(item)
		if typedErr != nil {
			notificationErr := Notifier.LogTypedError(typedErr)
			if notificationErr != nil {
				errStr := fmt.Sprintf("Error logging error: %v\n", typedErr)
				log.Fatal(errStr)
			}
			log.Panic(typedErr.Error())
		}
	}

	return nil
}

func CreateNotificationHandler() *NotificationHandler {
	UserIDs := []string{"261512678269255681"}
	ping := Handlers.NewPingType(nil, UserIDs, nil)
	MatchHandler := Handlers.NewDiscordMatchNotifier(url, *ping, "1207889")
	ErrorHandler := Handlers.NewDiscordErrorHandler(url, "15548997")
	return &NotificationHandler{
		MatchNotifier: MatchHandler,
		ErrorHandler:  ErrorHandler,
	}
}

func CreateDatabaseHandler() *DatabaseHandler {
	path := "C:\\Users\\Adam\\.oci\\config.ini"
	db := Database.NewOracleConnection(path, common.RegionLHR, time.Hour*3)
	return &DatabaseHandler{
		Database: db,
	}
}

func filterMatches(matches []MatchFinder.Match, Db *DatabaseHandler) (*Database.DataError, []MatchFinder.Match) {
	var newMatches []MatchFinder.Match
	for i, v := range matches {
		item := NewDatabaseItem(v)
		getErr, _ := Db.Get("id", item.ID)
		if getErr == nil {
			continue
		}

		if getErr.ErrorType() == "No Row" {
			newMatches = append(newMatches, matches[i])
			continue
		}

		return getErr, nil
	}

	return nil, newMatches
}

func NewDatabaseItem(match MatchFinder.Match) *DbItem {
	shortIdBytes := md5.Sum([]byte(match.ID))
	shortID := hex.EncodeToString(shortIdBytes[:])
	return &DbItem{
		ID:             shortID,
		EventStartDate: match.StartDate,
	}
}
