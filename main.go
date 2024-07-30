package main

import (
	"Adam/discord-twoup/MatchFinder"
	"fmt"
)

func main() {
	qt := MatchFinder.TwoUp

	qR := MatchFinder.Find(qt)
	if qR != nil {
		fmt.Println("Executed:", qR)
	} else {
		fmt.Println("Failed!")
	}
}