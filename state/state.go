package state

import (
	"fmt"
	"slices"
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

type StateDataProvider interface {
	GetItemsByDate(date time.Time) ([]Item, bool, error)
	SaveState(s *State) error
}

type ItemStatus int

const (
	Present ItemStatus = iota
	Absent
	Cancelled
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
	changed           bool
	LastRenderedLines int
}

func GetInitialState(repo StateDataProvider) (*State, error) {
	state := &State{
		Date:              CURR_DAY,
		AtMaxDate:         true,
		CachedDates:       make(ItemsMap),
		Cursor:            0,
		changed:           false,
		LastRenderedLines: -1, // to prevent clearing cli command
	}
	err := state.loadItems(repo)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (s *State) toggleCancel() {
	s.changed = true
	if s.Items[s.Cursor].Status == Cancelled {
		s.Items[s.Cursor].Status = Absent
	} else {
		s.Items[s.Cursor].Status = Cancelled
	}
}

func (s *State) toggleItem() {
	s.changed = true
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

func HandleInput(s *State, input string, repo StateDataProvider) (confirm bool, quit bool) {
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

func (s *State) loadItems(repo StateDataProvider) (err error) {
	s.changed = true
	unscheduledAsCancelled := config.GetCfg().UnscheduledAsCancelled
	if newItems, found := s.CachedDates[s.Date]; found {
		s.Items = newItems
	} else {
		newItems, found, err := repo.GetItemsByDate(s.Date)
		if err != nil {
			return fmt.Errorf("Error getting items by date: %w", err)
		}

		allSubjectsSet := config.GetAllSubjectsSet()
		allSubjects := make([]string, 0, len(allSubjectsSet))
		// add to list and sort for persistent order
		for subject := range allSubjectsSet {
			allSubjects = append(allSubjects, subject)
		}
		slices.Sort(allSubjects)

		if found {
			s.Items = newItems
			s.changed = false
			for _, subject := range allSubjects {
				found := false
				for _, item := range s.Items {
					if item.Name == subject {
						found = true
						break
					}
				}
				if !found {
					s.Items = append(s.Items, Item{
						Name:     subject,
						Selected: false,
						Status:   Cancelled,
					})
				}
			}
		} else {
			scheduledSubjects, err := config.GetNewSubjects(s.Date.Format("Monday"))
			if err != nil {
				return fmt.Errorf("Error getting initial items: %w", err)
			}
			s.Items = []Item{}
			if len(scheduledSubjects) > 0 {
				s.Items = make([]Item, len(scheduledSubjects))
				for i, name := range scheduledSubjects {
					s.Items[i] = Item{
						Name:     name,
						Selected: false,
						Status:   Absent,
					}
				}
			}
			if unscheduledAsCancelled {
				for _, subject := range allSubjects {
					if !slices.Contains(scheduledSubjects, subject) {
						s.Items = append(s.Items, Item{
							Name:     subject,
							Selected: false,
							Status:   Cancelled,
						})
					}
				}
			}
		}
	}
	newItemsLen := len(s.Items)
	if newItemsLen > 0 && s.Cursor >= newItemsLen {
		s.Cursor = len(s.Items) - 1
	}
	return nil
}

func (s *State) stepDay(direction string, repo StateDataProvider) error {
	if !s.AtMaxDate {
		if s.changed {
			s.CachedDates[s.Date] = s.Items
		}
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
