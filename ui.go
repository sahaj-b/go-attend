package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	// ansi control codes
	DATE_FORMAT_UI   = "02 Jan 2006"
	WEEKDAY_FORMAT   = "Monday"
	hideCursor       = "\x1b[?25l"
	showCursor       = "\x1b[?25h"
	saveCursorPos    = "\x1b[s"
	restoreCursorPos = "\x1b[u"
	clearDown        = "\x1b[J"
	moveUp           = "\x1b[%dA"

	// styles
	resetStyle = "\x1b[0m"
	bold       = "\x1b[1m"
	magenta    = "\x1b[38;2;255;0;255m"
	white      = "\x1b[38;5;255m"
	red        = "\x1b[31m"
	green      = "\x1b[32m"
	cyan       = "\x1b[36m"
	gray       = "\x1b[38;5;247m"
	yellow     = "\x1b[33m"

	bggray = "\x1b[48;5;236m"
	strike = "\x1b[9m"

	// keycodes
	upArrowKey   = "\x1b[A"
	downArrowKey = "\x1b[B"
	ctrlC        = "\x03"
	kpEnterKey   = "\x1bOM"

	highlight  = yellow
	cursorChar = highlight + "❯" + resetStyle
	leftArrow  = gray + "←" + resetStyle
	rightArrow = gray + "→" + resetStyle
)

func initScreen() (restorer func(), err error) {
	fmt.Print(hideCursor + saveCursorPos)
	cmd := exec.Command("stty", "-F", "/dev/tty", "-g")
	initStateBytes, err := cmd.Output()
	initStateBytes = initStateBytes[:len(initStateBytes)-1]

	if err != nil {
		return nil, errors.New("Failed to get terminal state")
	}

	err = exec.Command("stty", "-F", "/dev/tty", "raw", "-echo").Run()
	if err != nil {
		return nil, errors.New("Failed to set terminal to raw mode")
	}
	return func() {
		err := exec.Command("stty", "-F", "/dev/tty", string(initStateBytes)).Run()
		fmt.Print(showCursor)
		if err != nil {
			fmt.Println("Failed to reset terminal state:", err)
		}
	}, nil
}

type Hint struct {
	key string
	val string
}

func hintComponent(hints []Hint) string {
	result := " "
	for _, hint := range hints {
		result += bggray + highlight + bold + " " + hint.key + resetStyle + bggray + ": " + hint.val + " " + resetStyle + " "
	}
	result += "\r\n"
	return result
}

func dateComponent(date time.Time) string {
	today := date.Format(DATE_FORMAT_UI)
	weekday := date.Format(WEEKDAY_FORMAT)
	return " " + leftArrow + " " +
		bggray +
		highlight + " " + weekday + " " + resetStyle +
		" " + highlight + today + " " + resetStyle +
		" " + rightArrow
}

func render(state *State) {
	output := "\r\n"
	fmt.Printf(moveUp+clearDown, state.lastRenderedLines)
	output += dateComponent(state.date) + "\r\n"
	// hints1 := []Hint{
	// 	{"↑/↓/j/k", "move cursor"},
	// 	{"←/→/h/l", "cycle dates"},
	// }
	//
	hints2 := []Hint{
		{"Space", "toggle attendance"},
		{"c", "toggle cancelled"},
		{"Enter", "confirm"},
		{"q", "quit"},
	}

	// output += hintComponent(hints1)
	output += "\r\n"
	for i, item := range state.items {
		if i == state.cursor {
			output += " " + cursorChar + bold + " " + item.status.style + item.status.bullet + " " + item.name + resetStyle + "\r\n"
		} else {
			output += "   " + item.status.style + item.status.bullet + " " + item.name + resetStyle + "\r\n"
		}
	}
	output += "\r\n"
	output += hintComponent(hints2)
	// renderedLines := len(state.items) + 3
	state.lastRenderedLines = strings.Count(output, "\r\n")
	fmt.Print(output)
}

func getInput() (string, error) {
	inputBuf := make([]byte, 3)
	n, err := os.Stdin.Read(inputBuf)
	if err != nil {
		return "", err
	}
	return string(inputBuf[:n]), nil
}
