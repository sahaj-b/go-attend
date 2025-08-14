package ui

import "os"

var (
	ResetStyle string
	Bold       string
	Magenta    string
	White      string
	Red        string
	Green      string
	Cyan       string
	Gray       string
	Yellow     string
	Bggray     string
	MoreGray   string
	Strike     string
	Disabled   string
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		// NO_COLOR is set, disable all styles
		ResetStyle = ""
		Bold = ""
		Magenta = ""
		White = ""
		Red = ""
		Green = ""
		Cyan = ""
		Gray = ""
		Yellow = ""
		Bggray = ""
		MoreGray = ""
		Strike = ""
		Disabled = ""
	} else {
		// Normal colored mode
		ResetStyle = "\x1b[0m"
		Bold = "\x1b[1m"
		Magenta = "\x1b[38;2;255;0;255m"
		White = "\x1b[38;5;255m"
		Red = "\x1b[31m"
		Green = "\x1b[32m"
		Cyan = "\x1b[36m"
		Gray = "\x1b[38;5;247m"
		Yellow = "\x1b[33m"
		Bggray = "\x1b[48;5;236m"
		MoreGray = "\x1b[38;5;241m"
		Strike = "\x1b[9m"
		Disabled = "\x1b[38;5;240m"
	}
}
