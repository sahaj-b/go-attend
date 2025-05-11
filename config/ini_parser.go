package config

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sahaj-b/go-attend/utils"
)

//go:embed config_template.ini
var configTemplateContent string

const (
	keyStartDate              = "start_date"
	keyUnscheduledAsCancelled = "unscheduled_as_cancelled"
	sectionSchedule           = "schedule"
	sectionGeneral            = "general"
)

func getCfgFilePath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Failed to get user config dir: %w", err)
	}
	path := filepath.Join(cfgDir, "go-attend", "config.ini")
	return path, nil
}

func ensureConfigFileWithTemplate() error {
	cfgFilePath, err := getCfgFilePath()
	if err != nil {
		return fmt.Errorf("Failed to get config file path: %w", err)
	}
	if _, err := os.Stat(cfgFilePath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(cfgFilePath), 0o755)
		if err != nil {
			return fmt.Errorf("Failed to create config directory: %w", err)
		}
		file, err := os.Create(cfgFilePath)
		if err != nil {
			return fmt.Errorf("Failed to create config file: %w", err)
		}
		defer file.Close()

		_, err = file.WriteString(configTemplateContent)
		if err != nil {
			return fmt.Errorf("Failed to write config template to %s: %w", cfgFilePath, err)
		}
	}
	return nil
}

func parseGeneralEntry(key, value string, cfg *Config) error {
	switch key {

	case keyStartDate:
		if value == "" {
		} else {
			startDate, err := time.Parse("02-01-2006", value)
			if err != nil {
				return fmt.Errorf("Invalid Start Date format: %v. Expected format: dd-mm-yyyy", value)
			}
			cfg.StartDate = startDate
		}

	case keyUnscheduledAsCancelled:
		switch value {
		case "true":
			cfg.UnscheduledAsCancelled = true
		case "false":
			cfg.UnscheduledAsCancelled = false
		default:
			return fmt.Errorf("Invalid value for %v: %v. Expected true or false", key, value)
		}

	default:
		return fmt.Errorf("Invalid key: %v in [%v] section", key, sectionGeneral)
	}
	return nil
}

func parseScheduleEntry(key, value string, cfg *Config, subjectFound *bool) error {
	if _, exists := cfg.Schedule[key]; !exists {
		return fmt.Errorf("Invalid key in schedule: %v. Expected a day of the week(e.g., monday)", key)
	}
	subjects := strings.Split(value, ",")
	if len(subjects[0]) != 0 {
		for i := range subjects {
			subjects[i] = strings.TrimSpace(subjects[i])
			if len(subjects[i]) == 0 {
				return fmt.Errorf("Subject cannot be empty (on line: '%v=%v')", key, value)
			}
		}
		subjectsSet := make(map[string]struct{})
		for _, subject := range subjects {
			if _, exists := subjectsSet[subject]; exists {
				return fmt.Errorf("Duplicate subject found: %v on %v", subject, key)
			}
			subjectsSet[subject] = struct{}{}
		}
		cfg.Schedule[key] = subjects
		*subjectFound = true
	}
	return nil
}

func parseIni(reader io.Reader) (Config, error) {
	cfg := Config{
		StartDate: time.Time{},
		Schedule: map[string][]string{
			"monday":    {},
			"tuesday":   {},
			"wednesday": {},
			"thursday":  {},
			"friday":    {},
			"saturday":  {},
			"sunday":    {},
		},
		UnscheduledAsCancelled: false,
	}
	section := ""
	scanner := bufio.NewScanner(reader)
	subjectFound := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == ';' || line[0] == '#' {
			continue
		} else if line[0] == '[' && line[len(line)-1] == ']' {
			section = strings.ToLower(line[1 : len(line)-1])
		} else if strings.Contains(line, "=") {
			keyValue := strings.SplitN(line, "=", 2)
			if len(keyValue) != 2 {
				return Config{}, fmt.Errorf("Invalid key-value pair: %v", line)
			}
			key := strings.ToLower(strings.TrimSpace(keyValue[0]))
			value := strings.TrimSpace(keyValue[1])
			switch section {
			case sectionGeneral:
				if err := parseGeneralEntry(key, value, &cfg); err != nil {
					return Config{}, err
				}
			case sectionSchedule:
				if err := parseScheduleEntry(key, value, &cfg, &subjectFound); err != nil {
					return Config{}, err
				}
			default:
				return Config{}, fmt.Errorf("The key %v is not under a valid section", key)
			}
		} else {
			return Config{}, fmt.Errorf("Invalid line: %v", line)
		}
	}

	if !subjectFound {
		return Config{}, fmt.Errorf("At least one subject must be defined in the config")
	}
	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("Error reading config file: %w", err)
	}
	return cfg, nil
}

func loadAndParseConfig() (Config, error) {
	cfgFilePath, err := getCfgFilePath()
	if err != nil {
		return Config{}, err
	}

	err = ensureConfigFileWithTemplate()
	if err != nil {
		return Config{}, err
	}

	cfgFile, err := utils.EnsureAndGetFile(cfgFilePath, "r")
	if err != nil {
		return Config{}, fmt.Errorf("Failed to open config file: %v: %w", cfgFilePath, err)
	}
	defer cfgFile.Close()

	parsedCfg, err := parseIni(cfgFile)
	if err != nil {
		return Config{}, fmt.Errorf("INVALID CONFIG\n%w", err)
	}
	return parsedCfg, nil
}
