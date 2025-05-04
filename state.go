package main

import (
	"errors"
	"time"
)

type ItemStatus struct {
	bullet string
	text   string
	style  string
}

var (
	attended  ItemStatus = ItemStatus{style: green, bullet: "●", text: "Attended"}
	absent    ItemStatus = ItemStatus{style: white, bullet: "○", text: "Absent"}
	cancelled ItemStatus = ItemStatus{style: gray + strike, bullet: "×", text: "Cancelled"}
)

type Item struct {
	name     string
	selected bool
	status   ItemStatus
}

type State struct {
	date              time.Time
	items             []Item
	cachedDates       map[time.Time][]Item
	cursor            int
	lastRenderedLines int
}

func (state *State) toggleCancel() {
	if state.items[state.cursor].status == cancelled {
		state.items[state.cursor].status = absent
	} else {
		state.items[state.cursor].status = cancelled
	}
}

func (state *State) toggleItem() {
	switch state.items[state.cursor].status {
	case attended:
		state.items[state.cursor].status = absent
	case absent:
		state.items[state.cursor].status = attended
	case cancelled:
		state.items[state.cursor].status = attended
	}
}

func (state *State) moveCursor(direction string) {
	switch direction {
	case "down":
		if state.cursor < len(state.items)-1 {
			state.cursor++
		}
	case "up":
		if state.cursor > 0 {
			state.cursor--
		}
	}
}

func handleInput(state *State, input string) (confirm bool, quit bool) {
	confirm, quit = false, false
	switch input {
	case upArrowKey, "k":
		state.moveCursor("up")
	case downArrowKey, "j":
		state.moveCursor("down")
	case " ":
		state.toggleItem()
	case "c":
		state.toggleCancel()
	case kpEnterKey, "\n", "\r", "\r\n":
		confirm, quit = true, true
	case ctrlC, "q":
		confirm, quit = false, true
	}
	return confirm, quit
}

func (state *State) stepDay(direction string) (atMaxDate bool, err error) {
	// TODO: Complete this shit

	// ogDate := state.date

	// cool date equality check
	if state.date.Truncate(24 * time.Hour).Equal(time.Now().Truncate(24 * time.Hour)) {
		return true, nil
	}
	switch direction {
	case "next":
		state.date = state.date.AddDate(0, 0, 1)
	case "prev":
		state.date = state.date.AddDate(0, 0, -1)
	default:
		return false, errors.New("Invalid direction")
	}
	return false, nil
}
