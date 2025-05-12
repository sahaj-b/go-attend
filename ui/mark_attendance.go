package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/sahaj-b/go-attend/core"
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

	var cmdGetState, cmdSetRaw *exec.Cmd
	var sttyBaseArgs []string

	switch runtime.GOOS {
	case "linux":
		sttyBaseArgs = []string{"stty", "-F", "/dev/tty"}
		getArg := append(sttyBaseArgs, "-g")
		cmdGetState = exec.Command(getArg[0], getArg[1:]...)
		rawArg := append(sttyBaseArgs, "raw", "-echo")
		cmdSetRaw = exec.Command(rawArg[0], rawArg[1:]...)
	case "darwin":
		cmdGetState = exec.Command("sh", "-c", "stty -g < /dev/tty")
		cmdSetRaw = exec.Command("sh", "-c", "stty raw -echo < /dev/tty")
	default:
		fmt.Print(showCursor)
		return nil, fmt.Errorf("Unsupported OS: %s. This code's too good for it", runtime.GOOS)
	}

	initStateBytes, err := cmdGetState.Output()
	if err != nil {
		fmt.Print(showCursor)
		return nil, fmt.Errorf("Failed to get terminal state (cmd: '%s'): %v. Output: '%s'", strings.Join(cmdGetState.Args, " "), err, string(initStateBytes))
	}

	initStateStr := strings.TrimSpace(string(initStateBytes))

	if err = cmdSetRaw.Run(); err != nil {
		fmt.Print(showCursor)
		restoreArgs := sttyBaseArgs
		if initStateStr != "" {
			settings := strings.Fields(initStateStr)
			if len(settings) > 0 {
				restoreArgs = append(restoreArgs, settings...)
			} else {
				restoreArgs = append(restoreArgs, "sane")
			}
		} else {
			restoreArgs = append(restoreArgs, "sane")
		}
		if len(restoreArgs) > len(sttyBaseArgs) {
			exec.Command(restoreArgs[0], restoreArgs[1:]...).Run()
		} else {
			saneFallbackArgs := append(sttyBaseArgs, "sane")
			exec.Command(saneFallbackArgs[0], saneFallbackArgs[1:]...).Run()
		}
		return nil, fmt.Errorf("Failed to set terminal to raw mode (cmd: '%s'): %v", strings.Join(cmdSetRaw.Args, " "), err)
	}

	restorer = func() {
		settings := strings.Fields(initStateStr)
		var cmdRestore *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			restoreCmdStr := "stty "
			if len(settings) > 0 {
				restoreCmdStr += strings.Join(settings, " ")
			} else {
				restoreCmdStr += "sane"
			}
			restoreCmdStr += " < /dev/tty"
			cmdRestore = exec.Command("sh", "-c", restoreCmdStr)
		case "linux":
			restoreCmdArgs := sttyBaseArgs
			if len(settings) > 0 {
				restoreCmdArgs = append(restoreCmdArgs, settings...)
			} else {
				restoreCmdArgs = append(restoreCmdArgs, "sane")
			}
			cmdRestore = exec.Command(restoreCmdArgs[0], restoreCmdArgs[1:]...)
		default:
			fmt.Print(showCursor)
			fmt.Fprintf(os.Stderr, "Unsupported OS: Cannot determine how to restore terminal state.\n")
			return
		}

		errRestore := cmdRestore.Run()
		fmt.Print(showCursor)
		if errRestore != nil {
			fmt.Fprintf(os.Stderr, "CRITICAL: Failed to reset terminal state (cmd: '%s'): %v. Good luck\n", strings.Join(cmdRestore.Args, " "), errRestore)
		}
	}
	return restorer, nil
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

func getStyleAndBullet(item state.Item) (string, string) {
	itemStyle := ""
	itemBullet := ""
	switch item.Status {
	case core.Present:
		itemStyle = Green
		itemBullet = "●"
	case core.Absent:
		itemStyle = ""
		itemBullet = "○"
	case core.Cancelled:
		itemStyle = Gray + Strike
		itemBullet = "✗"
	}
	return itemStyle, itemBullet
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
			itemStyle, itemBullet := getStyleAndBullet(item)
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
