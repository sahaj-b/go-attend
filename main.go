package main

import (
	"fmt"
	"time"
)

var CURR_DAY = time.Now().Truncate(time.Hour * 24)

func main() {
	restorer, err := initScreen()
	if err != nil {
		printRed("Error initializing terminal:" + err.Error())
		return
	}
	defer restorer()
	csvStore, err := NewCSVStore()
	if err != nil {
		printRed("Error creating CSV store:" + err.Error())
		return
	}
	records, err := csvStore.GetAllRecords()
	if err != nil {
		printRed("Error getting records:" + err.Error())
		return
	}
	if err := validateAndFixRecords(&records); err != nil {
		printRed("Error validating records:" + err.Error())
		return
	}
	state, err := getInitialState(csvStore)
	if err != nil {
		printRed("Error getting initial state:" + err.Error())
		return
	}
	confirm, quit := false, false

	for !quit {
		render(state)
		inp, err := getInput()
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}
		confirm, quit = handleInput(state, inp, csvStore)
	}
	fmt.Println()
	if confirm {
		if err := csvStore.SaveState(state); err != nil {
			printRed("Error saving items:" + err.Error())
		} else {
			printGreen("Saved successfully")
		}
	} else {
		printRed("Cancelled")
	}
	fmt.Println()
}
