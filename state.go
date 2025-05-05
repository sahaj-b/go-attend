package main

import (
	"errors"
	"time"
)

const (
	// ansi keycodes
	upArrowKey    = "\x1b[A"
	downArrowKey  = "\x1b[B"
	leftArrowKey  = "\x1b[D"
	rightArrowKey = "\x1b[C"
	ctrlC         = "\x03"
	kpEnterKey    = "\x1bOM"
)

type ItemStatus struct {
	bullet string
	text   string
	style  string
}

var (
	present   ItemStatus = ItemStatus{style: green, bullet: "●", text: "Attended"}
	absent    ItemStatus = ItemStatus{style: white, bullet: "○", text: "Absent"}
	cancelled ItemStatus = ItemStatus{style: gray + strike, bullet: "×", text: "Cancelled"}
)

type Item struct {
	name     string
	selected bool
	status   ItemStatus
}

type ItemsMap map[time.Time][]Item

type State struct {
	date              time.Time
	atMaxDate         bool
	items             []Item
	cachedDates       ItemsMap
	cursor            int
	lastRenderedLines int
}

func getInitialState(store DataStore) (*State, error) {
	state := &State{
		date:              CURR_DAY,
		atMaxDate:         true,
		cachedDates:       make(ItemsMap),
		cursor:            0,
		lastRenderedLines: -1, // to prevent clearing cli command
	}
	err := state.loadItems(store)
	if err != nil {
		return nil, errors.New("Error loading items: " + err.Error())
	}
	return state, nil
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
	case present:
		state.items[state.cursor].status = absent
	case absent:
		state.items[state.cursor].status = present
	case cancelled:
		state.items[state.cursor].status = present
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

func handleInput(state *State, input string, store DataStore) (confirm bool, quit bool) {
	confirm, quit = false, false
	switch input {
	case upArrowKey, "k":
		state.moveCursor("up")
	case downArrowKey, "j":
		state.moveCursor("down")
	case leftArrowKey, "h":
		if err := state.stepDay("prev", store); err != nil {
			printRed("Error going previous day: " + err.Error())
			return false, true
		}
	case rightArrowKey, "l":
		if err := state.stepDay("next", store); err != nil {
			printRed("Error going next day: " + err.Error())
			return false, true
		}
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

func (state *State) loadItems(store DataStore) (err error) {
	if newItems, found := state.cachedDates[state.date]; found {
		state.items = newItems
	} else {
		newItems, found, err := store.GetItemsByDate(state.date.Format(DATE_FORMAT_STORE))
		if err != nil {
			return errors.New("Error getting items by date: " + err.Error())
		}
		if found {
			state.items = newItems
		} else {
			state.items, err = getNewItemsFromCfg(state.date.Format(WEEKDAY_FORMAT))
			if err != nil {
				return errors.New("Error getting initial items: " + err.Error())
			}
		}
	}
	newItemsLen := len(state.items)
	if newItemsLen > 0 && state.cursor >= newItemsLen {
		state.cursor = len(state.items) - 1
	}
	return nil
}

func (state *State) stepDay(direction string, store DataStore) error {
	if _, ok := state.cachedDates[state.date]; !ok {
		state.cachedDates[state.date] = state.items
	}
	switch direction {
	case "next":
		if !state.atMaxDate {
			state.date = state.date.AddDate(0, 0, 1)
		}
	case "prev":
		state.date = state.date.AddDate(0, 0, -1)
	default:
		return errors.New("Invalid direction")
	}

	if state.date.Equal(CURR_DAY) {
		state.atMaxDate = true
	} else {
		state.atMaxDate = false
	}

	err := state.loadItems(store)
	if err != nil {
		return errors.New("Error loading new items: " + err.Error())
	}
	return nil
}
