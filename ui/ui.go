package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sahaj-b/go-attend/state"
)

const (
	// ansi control codes
	DATE_FORMAT_UI   = "02 Jan 2006"
	WEEKDAY_FORMAT   = "Mon"
	hideCursor       = "\x1b[?25l"
	showCursor       = "\x1b[?25h"
	saveCursorPos    = "\x1b[s"
	restoreCursorPos = "\x1b[u"
	clearDown        = "\x1b[J"
	moveUp           = "\x1b[%dA"

	highlight          = Yellow
	cursorChar         = highlight + "❯" + ResetStyle
	leftArrow          = Gray + "←" + ResetStyle
	rightArrow         = Gray + "→" + ResetStyle
	disabledRightArrow = Disabled + "→" + ResetStyle
)

type Hint struct {
	key string
	val string
}

var hints = []Hint{
	{"Space", "Present/Absent"},
	{"c", "Mark Cancelled"},
	{"Enter", "Confirm"},
	{"q", "Quit"},
}

func InitScreen() (restorer func(), err error) {
	fmt.Print(hideCursor + saveCursorPos)
	cmd := exec.Command("stty", "-F", "/dev/tty", "-g")
	initStateBytes, err := cmd.Output()
	initStateBytes = initStateBytes[:len(initStateBytes)-1]

	if err != nil {
		return nil, fmt.Errorf("Failed to get terminal state")
	}

	err = exec.Command("stty", "-F", "/dev/tty", "raw", "-echo").Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to set terminal to raw mode")
	}
	return func() {
		err := exec.Command("stty", "-F", "/dev/tty", string(initStateBytes)).Run()
		fmt.Print(showCursor)
		if err != nil {
			fmt.Println("Failed to reset terminal state:", err)
		}
	}, nil
}

func hintComponent(hints []Hint) string {
	result := " "
	for _, hint := range hints {
		result += Bggray + highlight + Bold + " " + hint.key + ResetStyle + Bggray + ": " + hint.val + " " + ResetStyle + " "
	}
	result += "\r\n"
	return result
}

func dateComponent(date time.Time, atMaxDate bool) string {
	today := date.Format(DATE_FORMAT_UI)
	weekday := date.Format(WEEKDAY_FORMAT)
	rightArrow := rightArrow
	if atMaxDate {
		rightArrow = disabledRightArrow
	}
	return " " + leftArrow + " " +
		Bggray +
		highlight + " " + weekday + " " + ResetStyle +
		" " + highlight + today + " " + ResetStyle +
		" " + rightArrow
}

func noClassesComponent(weekday string) string {
	return "   " + Yellow + Bold + "No classes for " + weekday + ResetStyle
}

func Render(s *state.State) {
	var output strings.Builder
	output.WriteString("\r\n")
	fmt.Printf(moveUp+clearDown, s.LastRenderedLines)
	output.WriteString(dateComponent(s.Date, s.AtMaxDate) + "\r\n")
	output.WriteString("\r\n")
	if len(s.Items) == 0 {
		output.WriteString("\r\n" + noClassesComponent(s.Date.Format("Monday")) + "\r\n\r\n")
	} else {
		for i, item := range s.Items {
			itemStyle := ""
			itemBullet := ""
			switch item.Status.Kind {
			case state.PresentStatus:
				itemStyle = Green
				itemBullet = "●"
			case state.AbsentStatus:
				itemStyle = ""
				itemBullet = "○"
			case state.CancelledStatus:
				itemStyle = Gray + Strike
				itemBullet = "✗"
			}
			if i == s.Cursor {
				output.WriteString(" " + cursorChar + Bold + " " + itemStyle + itemBullet + " " + item.Name + ResetStyle + "\r\n")
			} else {
				output.WriteString("   " + itemStyle + itemBullet + " " + item.Name + ResetStyle + "\r\n")
			}
		}
	}
	output.WriteString("\r\n")
	output.WriteString(hintComponent(hints))
	// renderedLines := len(state.items) + 3
	outputStr := output.String()
	s.LastRenderedLines = strings.Count(outputStr, "\r\n")
	fmt.Print(outputStr)
}

func GetInput() (string, error) {
	inputBuf := make([]byte, 3)
	n, err := os.Stdin.Read(inputBuf)
	if err != nil {
		return "", err
	}
	return string(inputBuf[:n]), nil
}
