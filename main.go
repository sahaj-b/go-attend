package main

import (
	"fmt"
	"time"
)

var CURR_TIME = time.Now()

func main() {
	restorer, err := initScreen()
	if err != nil {
		printRed("Error initializing terminal:" + err.Error())
		return
	}
	defer restorer()
	records, err := getAllRecords()
	if err != nil {
		printRed("Error getting records:" + err.Error())
		return
	}
	if err := records.validate(); err != nil {
		printRed("Error validating records:" + err.Error())
		return
	}
	state := State{
		items: []Item{
			{name: "English", selected: false, status: attended},
			{name: "Math", selected: false, status: absent},
			{name: "Science", selected: false, status: attended},
			{name: "History", selected: false, status: absent},
			{name: "Geography", selected: false, status: attended},
		},
		cursor:            0,
		lastRenderedLines: -1, // to prevent clearing cli command
	}
	confirm, quit := false, false

	for !quit {
		render(&state)
		inp, err := getInput()
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}
		confirm, quit = handleInput(&state, inp)
	}
	fmt.Println()
	if confirm {
		if err := records.handleSave(CURR_TIME.Format(DATE_FORMAT_STORE), state.items); err != nil {
			printRed("Error saving items:" + err.Error())
		} else {
			printGreen("Saved successfully")
		}
	} else {
		printRed("Cancelled by User")
	}
	fmt.Println()
}
