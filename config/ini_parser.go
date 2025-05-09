package config

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sahaj-b/go-attend/utils"
)

type Config struct {
	StartDate              time.Time
	Schedule               map[string][]string
	UnscheduledAsCancelled bool
}

var (
	globalCfg Config
	loadOnce  sync.Once
	loadErr   error
)

const (
	keyStartDate              = "start_date"
	keyUnscheduledAsCancelled = "unscheduled_as_cancelled"
	sectionSchedule           = "schedule"
)

func getCfgFilePath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Failed to get user config dir: %w", err)
	}
	path := filepath.Join(cfgDir, "go-attend", "config.ini")
	return path, nil
}

// TODO: Review
func writeStartDateInFile(file *os.File) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to beginning of file: %w", err)
	}

	var lines []string
	targetLineIndex := -1
	keyFoundAndValueEmpty := false

	scanner := bufio.NewScanner(file)
	currentLineNumber := 0
	for scanner.Scan() {
		originalLine := scanner.Text()
		lines = append(lines, originalLine)

		if targetLineIndex == -1 {
			trimmedLine := strings.TrimSpace(originalLine)
			parts := strings.SplitN(trimmedLine, "=", 2)

			if len(parts) > 0 && strings.TrimSpace(parts[0]) == keyStartDate {
				targetLineIndex = currentLineNumber
				if len(parts) == 1 || (len(parts) == 2 && strings.TrimSpace(parts[1]) == "") {
					keyFoundAndValueEmpty = true
				}
			}
		}
		currentLineNumber++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading config file content: %w", err)
	}
	if keyFoundAndValueEmpty {
		newDateString := time.Now().Format("02-01-2006")
		lines[targetLineIndex] = keyStartDate + " = " + newDateString
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek to beginning: %w", err)
		}
		writer := bufio.NewWriter(file)
		for _, line := range lines {
			if _, err := writer.WriteString(line + "\n"); err != nil {
				writer.Flush()
				return fmt.Errorf("failed to write line back to file: %w", err)
			}
		}
		writer.Flush()
	}

	return nil
}

func parseIni(file *os.File) (Config, error) {
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
	scanner := bufio.NewScanner(file)
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
			case "":
				{
					switch key {

					case keyStartDate:
						if value == "" {
						} else {
							startDate, err := time.Parse("02-01-2006", value)
							if err != nil {
								return Config{}, fmt.Errorf("Invalid Start Date format: %v. Expected format: dd-mm-yyyy", value)
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
							return Config{}, fmt.Errorf("Invalid value for %v: %v. Expected true or false", key, value)
						}

					default:
						return Config{}, fmt.Errorf("Invalid key: %v in Default section", key)
					}
				}

			case sectionSchedule:
				{
					if _, err := time.Parse("monday", key); err != nil {
						return Config{}, fmt.Errorf("Invalid key in schedule: %v. Expected a day of the week(e.g., monday)", key)
					}
					subjects := strings.Split(value, ",")
					for i := range subjects {
						subjects[i] = strings.TrimSpace(subjects[i])
						if len(subjects[i]) == 0 {
							return Config{}, fmt.Errorf("Subject cannot be empty: %v", line)
						}
					}
					cfg.Schedule[key] = subjects
				}
			default:
				return Config{}, fmt.Errorf("Invalid section: [%v]", section)
			}
		}
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

func GetCfg() Config {
	loadOnce.Do(func() {
		cfg, err := loadAndParseConfig()
		if err != nil {
			loadErr = err
			log.Fatalf("Failed to load config: %v", err)
		}
		globalCfg = cfg
	})
	return globalCfg
}
