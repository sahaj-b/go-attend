package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func mustParseTime(t *testing.T, value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	tm, err := time.Parse("02-01-2006", value)
	if err != nil {
		t.Fatalf("Test setup error: Failed to parse time literal '%s': %v", value, err)
	}
	return tm
}

func defaultSchedule() map[string][]string {
	return map[string][]string{
		"monday":    {},
		"tuesday":   {},
		"wednesday": {},
		"thursday":  {},
		"friday":    {},
		"saturday":  {},
		"sunday":    {},
	}
}

func TestParseIni(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		expectedCfg   Config
		isErr         bool
	}{
		{
			name: "Valid full config",
			configContent: `
[general]
start_date = 01-08-2023
unscheduled_as_cancelled = true
[schedule]
monday = Math, Physics 
tuesday = Chemistry
# wednesday is a holiday
thursday =
friday = Biology, Lab
saturday = Sports
sunday =
			`,
			expectedCfg: Config{
				StartDate: mustParseTime(t, "01-08-2023"),
				Schedule: func() map[string][]string {
					s := defaultSchedule()
					s["monday"] = []string{"Math", "Physics"}
					s["tuesday"] = []string{"Chemistry"}
					s["friday"] = []string{"Biology", "Lab"}
					s["saturday"] = []string{"Sports"}
					return s
				}(),
				UnscheduledAsCancelled: true,
			},
			isErr: false,
		},
		{
			name: "Valid config with no subjects on some days and empty start date",
			configContent: `
[general]
start_date =
unscheduled_as_cancelled = false
[schedule]
monday = Geography
tuesday =
			`,
			expectedCfg: Config{
				StartDate: time.Time{},
				Schedule: func() map[string][]string {
					s := defaultSchedule()
					s["monday"] = []string{"Geography"}
					return s
				}(),
				UnscheduledAsCancelled: false,
			},
			isErr: false,
		},
		{
			name: "Error: Invalid start_date format",
			configContent: `
[general]
start_date = 2023-08-01 
[schedule]
monday = Math
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Invalid unscheduled_as_cancelled value",
			configContent: `
[general]
start_date = 01-08-2023
unscheduled_as_cancelled = maybe
[schedule]
monday = Math
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Unknown key in general section",
			configContent: `
[general]
start_date = 01-08-2023
favorite_color = blue 
[schedule]
monday = Math
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Unknown day in schedule section (not a valid weekday key)",
			configContent: `
[general]
start_date = 01-08-2023
[schedule]
funday = Math 
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Empty subject string in a list",
			configContent: `
[general]
start_date = 01-08-2023
[schedule]
monday = Math, , Physics 
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Duplicate subject",
			configContent: `
[general]
start_date = 01-08-2023
[schedule]
monday = Math, Physics, Math
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: No subjects defined at all (all days empty or commented)",
			configContent: `
[general]
start_date = 01-08-2023
unscheduled_as_cancelled = true
[schedule]
monday = 
tuesday =
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Key outside of any section",
			configContent: `
oops_a_key = value
[general]
start_date = 01-08-2023
[schedule]
monday = Math
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Malformed key-value pair (no equals)",
			configContent: `
[general]
start_date 01-08-2023
[schedule]
monday = Math
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name:          "Error: Empty input string (no subjectFound)",
			configContent: ``,
			expectedCfg:   Config{Schedule: defaultSchedule()},
			isErr:         true,
		},
		{
			name: "Error: Input with only comments (no subjectFound)",
			configContent: `
# This is a comment
; So is this
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Schedule key with completely empty value (no subjects defined anywhere)",
			configContent: `
[general]
start_date = 01-01-2025
[schedule]
monday = 
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Valid: Schedule key with empty value, but another day has subjects",
			configContent: `
[general]
start_date = 01-01-2025
unscheduled_as_cancelled = false
[schedule]
monday = 
tuesday = Science
			`,
			expectedCfg: Config{
				StartDate: mustParseTime(t, "01-01-2025"),
				Schedule: func() map[string][]string {
					s := defaultSchedule()
					s["tuesday"] = []string{"Science"}
					return s
				}(),
				UnscheduledAsCancelled: false,
			},
			isErr: false,
		},
		{
			name: "Error: Malformed section name (missing closing bracket)",
			configContent: `
[general
start_date = 01-01-2023
[schedule]
monday = Math
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
		{
			name: "Error: Section name with leading/trailing spaces",
			configContent: `
[ general ]
start_date = 01-01-2023
[schedule]
monday = Math
			`,
			expectedCfg: Config{},
			isErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.configContent)
			actualCfg, err := parseIni(reader)

			if tt.isErr {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error ,but got: %v", err)
				}
				if !reflect.DeepEqual(actualCfg, tt.expectedCfg) {
					t.Errorf("Config mismatch: Expected: %+v Actual:   %+v", tt.expectedCfg, actualCfg)
					if !reflect.DeepEqual(actualCfg.Schedule, tt.expectedCfg.Schedule) {
						t.Logf("Schedule Expected: %+v", tt.expectedCfg.Schedule)
						t.Logf("Schedule Actual:   %+v", actualCfg.Schedule)
					}
					if actualCfg.StartDate != tt.expectedCfg.StartDate {
						t.Logf("StartDate Expected: %s", tt.expectedCfg.StartDate)
						t.Logf("StartDate Actual:   %s", actualCfg.StartDate)
					}
					if actualCfg.UnscheduledAsCancelled != tt.expectedCfg.UnscheduledAsCancelled {
						t.Logf("UnscheduledAsCancelled Expected: %t", tt.expectedCfg.UnscheduledAsCancelled)
						t.Logf("UnscheduledAsCancelled Actual:   %t", actualCfg.UnscheduledAsCancelled)
					}
				}
			}
		})
	}
}
