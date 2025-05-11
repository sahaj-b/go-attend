package config

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	loadOnce  sync.Once
	loadErr   error
	globalCfg Config
)

type Config struct {
	StartDate              time.Time
	Schedule               map[string][]string
	UnscheduledAsCancelled bool
}

type subjectsSet map[string]struct{} // golang trick for making a set, struct{} takes 0 bytes

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

func GetAllSubjectsSet() subjectsSet {
	cfg := GetCfg()
	subjectsSet := subjectsSet{}
	for _, daySubjects := range cfg.Schedule {
		for _, subject := range daySubjects {
			if strings.TrimSpace(subject) != "" {
				subjectsSet[subject] = struct{}{}
			}
		}
	}
	return subjectsSet
}

func GetNewSubjects(weekday string) ([]string, error) {
	cfg := GetCfg()
	subjects, ok := cfg.Schedule[strings.ToLower(weekday)]
	if !ok {
		return nil, fmt.Errorf("Invalid weekday: %v", weekday)
	}
	return subjects, nil
}
