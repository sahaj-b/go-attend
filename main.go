package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sahaj-b/go-attend/config"
	"github.com/sahaj-b/go-attend/state"
	"github.com/sahaj-b/go-attend/store"
	"github.com/sahaj-b/go-attend/ui"
)

const (
	DATE_FORMAT_ARG      = "02-01-2006"
	DATE_FORMAT_ARG_SHOW = "DD-MM-YYYY"
)

func main() {
	args := os.Args
	date := state.CURR_DAY
	if len(args) > 1 {
		switch args[1] {
		case "stats":
			handleStatsArgs(args)
			return
		case "rename":
			handleRenameArgs(args)
			return
		case "help", "-h", "--help":
			printHelp()
			return
		case "config-file":
			path, err := config.GetCfgFilePath()
			if err != nil {
				ui.Error("Error getting config file path: " + err.Error())
			}
			fmt.Println("Config file path:", path)
			return
		default:
			argDate, err := time.Parse(DATE_FORMAT_ARG, args[1])
			if err != nil {
				ui.Error("Invalid argument: " + args[1])
				printHelp()
				return
			}
			date = argDate
		}
	}
	config.GetCfg()
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
	currState, err := state.GetInitialState(csvStore, date)
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
			ui.Error("Error saving items: " + err.Error())
		} else {
			ui.Success("Saved successfully")
		}
	} else {
		ui.Error("Cancelled")
	}
	fmt.Println()
}

func printHelp() {
	fmt.Println("Usage: go-attend [date|options]")
	fmt.Println("Date format: " + DATE_FORMAT_ARG_SHOW)
	fmt.Println("Options:")
	fmt.Println("  stats               Show stats")
	fmt.Println("  stats -h            Show stats usage and flags")
	fmt.Println("  rename [old] [new]  Rename a subject from 'old' to 'new'")
	fmt.Println("  config-file         Show config file path")
	fmt.Println("  -h, -help           Show this help message")
	fmt.Println()
}

func handleStatsArgs(args []string) {
	cfg := config.GetCfg()
	weekday := false
	startDate := cfg.StartDate
	endDate := time.Time{}
	if len(args) > 2 {
		statsCmd := flag.NewFlagSet("stats", flag.ExitOnError)
		statsCmd.BoolVar(&weekday, "weekday", false, "Show weekday wise stats")
		startDateStr := statsCmd.String("start", "", "Start date for the stats (format: "+DATE_FORMAT_ARG_SHOW+")")
		endDateStr := statsCmd.String("end", "", "End date for the stats (format: "+DATE_FORMAT_ARG_SHOW+")")
		statsCmd.Usage = func() {
			fmt.Println("Usage: go-attend stats [flags]")
			fmt.Println("Flags:")
			statsCmd.PrintDefaults()
		}
		err := statsCmd.Parse(args[2:])
		if err != nil {
			return
		}
		if *startDateStr != "" {
			argStartDate, err := time.Parse(DATE_FORMAT_ARG, *startDateStr)
			if err != nil {
				ui.Error("Invalid start date: " + *startDateStr)
				fmt.Println("Format must be: " + DATE_FORMAT_ARG_SHOW)
				return
			}
			startDate = argStartDate
		}
		if *endDateStr != "" {
			argEndDate, err := time.Parse(DATE_FORMAT_ARG, *endDateStr)
			if err != nil {
				ui.Error("Invalid end date: " + *endDateStr)
				fmt.Println("Format must be: " + DATE_FORMAT_ARG_SHOW)
				return
			}
			endDate = argEndDate
		}
	}
	csvStore, err := store.NewCSVStore()
	if err != nil {
		ui.Error("Error creating CSV store:" + err.Error())
		return
	}
	if weekday {
		ui.DisplayWeekdayWiseStats(csvStore, startDate, endDate)
	} else {
		ui.DisplaySubjectWiseStats(csvStore, startDate, endDate)
	}
}

func handleRenameArgs(args []string) {
	if len(args) < 4 {
		ui.Error("Not enough arguments for rename")
		fmt.Println("Usage: go-attend rename <old_name> <new_name>")
		return
	}

	oldName := args[2]
	newName := args[3]

	if oldName == newName {
		ui.Error("Old and new names cannot be the same")
		return
	}

	if oldName == "" || newName == "" {
		ui.Error("Subject names cannot be empty")
		return
	}

	csvStore, err := store.NewCSVStore()
	if err != nil {
		ui.Error("Error creating CSV store: " + err.Error())
		return
	}

	err = csvStore.RenameSubject(oldName, newName)
	if err != nil {
		ui.Error("Error renaming subject: " + err.Error())
		return
	}

	ui.Success(fmt.Sprintf("Successfully renamed '%s' to '%s'", oldName, newName))
}
