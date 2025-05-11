package config

import (
	"fmt"
	"strings"
)

type subjectsSet map[string]struct{} // golang trick for making a set, struct{} takes 0 bytes

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
