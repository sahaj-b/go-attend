package state

import (
	"fmt"
	"time"

	"github.com/sahaj-b/go-attend/config"
)

var CURR_DAY = time.Now().Truncate(time.Hour * 24)

const (
	// ansi keycodes
	upArrowKey    = "\x1b[A"
	downArrowKey  = "\x1b[B"
	leftArrowKey  = "\x1b[D"
	rightArrowKey = "\x1b[C"
	ctrlC         = "\x03"
	kpEnterKey    = "\x1bOM"
)

type Repository interface {
	GetItemsByDate(date time.Time) ([]Item, bool, error)
	SaveState(s *State) error
}

type StatusKind int

const (
	PresentStatus StatusKind = iota
	AbsentStatus
	CancelledStatus
)

type ItemStatus struct {
	Kind StatusKind
	Text string
}

var (
	Present   ItemStatus = ItemStatus{Kind: PresentStatus, Text: "Attended"}
	Absent    ItemStatus = ItemStatus{Kind: AbsentStatus, Text: "Absent"}
	Cancelled ItemStatus = ItemStatus{Kind: CancelledStatus, Text: "Cancelled"}
)

type Item struct {
	Name     string
	Selected bool
	Status   ItemStatus
}

type ItemsMap map[time.Time][]Item

type State struct {
	Date              time.Time
	AtMaxDate         bool
	Items             []Item
	CachedDates       ItemsMap
	Cursor            int
	LastRenderedLines int
}

func GetInitialState(repo Repository) (*State, error) {
	state := &State{
		Date:              CURR_DAY,
		AtMaxDate:         true,
		CachedDates:       make(ItemsMap),
		Cursor:            0,
		LastRenderedLines: -1, // to prevent clearing cli command
	}
	err := state.loadItems(repo)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (s *State) toggleCancel() {
	if s.Items[s.Cursor].Status == Cancelled {
		s.Items[s.Cursor].Status = Absent
	} else {
		s.Items[s.Cursor].Status = Cancelled
	}
}

func (s *State) toggleItem() {
	switch s.Items[s.Cursor].Status {
	case Present:
		s.Items[s.Cursor].Status = Absent
	case Absent:
		s.Items[s.Cursor].Status = Present
	case Cancelled:
		s.Items[s.Cursor].Status = Present
	}
}

func (s *State) moveCursor(direction string) {
	switch direction {
	case "down":
		if s.Cursor < len(s.Items)-1 {
			s.Cursor++
		}
	case "up":
		if s.Cursor > 0 {
			s.Cursor--
		}
	}
}

func HandleInput(s *State, input string, repo Repository) (confirm bool, quit bool) {
	confirm, quit = false, false
	switch input {
	case upArrowKey, "k":
		s.moveCursor("up")
	case downArrowKey, "j":
		s.moveCursor("down")
	case leftArrowKey, "h":
		if err := s.stepDay("prev", repo); err != nil {
			return false, true
		}
	case rightArrowKey, "l":
		if err := s.stepDay("next", repo); err != nil {
			return false, true
		}
	case " ":
		s.toggleItem()
	case "c":
		s.toggleCancel()
	case kpEnterKey, "\n", "\r", "\r\n":
		confirm, quit = true, true
	case ctrlC, "q":
		confirm, quit = false, true
	}
	return confirm, quit
}

func (s *State) loadItems(repo Repository) (err error) {
	if newItems, found := s.CachedDates[s.Date]; found {
		s.Items = newItems
	} else {
		newItems, found, err := repo.GetItemsByDate(s.Date)
		if err != nil {
			return fmt.Errorf("Error getting items by date: %w", err)
		}
		if found {
			s.Items = newItems
		} else {
			subjects, err := config.GetNewItems(s.Date.Format("Monday"))
			if err != nil {
				return fmt.Errorf("Error getting initial items: %w", err)
			}
			if len(subjects[0]) != 0 {
				s.Items = make([]Item, len(subjects))
				for i, name := range subjects {
					s.Items[i] = Item{
						Name:     name,
						Selected: false,
						Status:   Absent,
					}
				}
			} else {
				s.Items = []Item{}
			}
		}
	}
	newItemsLen := len(s.Items)
	if newItemsLen > 0 && s.Cursor >= newItemsLen {
		s.Cursor = len(s.Items) - 1
	}
	return nil
}

func (s *State) stepDay(direction string, repo Repository) error {
	if _, ok := s.CachedDates[s.Date]; !ok {
		s.CachedDates[s.Date] = s.Items
	}
	switch direction {
	case "next":
		if !s.AtMaxDate {
			s.Date = s.Date.AddDate(0, 0, 1)
		}
	case "prev":
		s.Date = s.Date.AddDate(0, 0, -1)
	default:
		return fmt.Errorf("Invalid direction")
	}

	if s.Date.Equal(CURR_DAY) {
		s.AtMaxDate = true
	} else {
		s.AtMaxDate = false
	}

	err := s.loadItems(repo)
	if err != nil {
		return fmt.Errorf("Error loading new items: %w", err)
	}
	return nil
}
