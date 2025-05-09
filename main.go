package main

import (
	"fmt"

	"github.com/sahaj-b/go-attend/state"
	"github.com/sahaj-b/go-attend/store"
	"github.com/sahaj-b/go-attend/ui"
)

func main() {
	restorer, err := ui.InitScreen()
	if err != nil {
		ui.Error("Error initializing terminal:" + err.Error())
		return
	}
	defer restorer()
	csvStore, err := store.NewCSVStore()
	if err != nil {
		ui.Error("Error creating CSV store:" + err.Error())
		return
	}
	currState, err := state.GetInitialState(csvStore)
	if err != nil {
		ui.Error("Error getting initial state:" + err.Error())
		return
	}
	confirm, quit := false, false

	for !quit {
		ui.Render(currState)
		inp, err := ui.GetInput()
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}
		confirm, quit = state.HandleInput(currState, inp, csvStore)
	}
	fmt.Println()
	if confirm {
		if err := csvStore.SaveState(currState); err != nil {
			ui.Error("Error saving items:" + err.Error())
		} else {
			ui.Success("Saved successfully")
		}
	} else {
		ui.Error("Cancelled")
	}
	fmt.Println()
}
