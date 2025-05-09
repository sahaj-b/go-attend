package config

import "fmt"

type subjectsSet map[string]struct{} // golang trick for making a set, struct{} takes 0 bytes

func GetAllSubjects() subjectsSet {
	cfg := GetCfg()
	subjectsSet := subjectsSet{}
	for _, daySubjects := range cfg.Schedule {
		for _, subject := range daySubjects {
			subjectsSet[subject] = struct{}{}
		}
	}
	return subjectsSet
}

func GetNewItems(weekday string) ([]string, error) {
	cfg := GetCfg()
	subjects, ok := cfg.Schedule[weekday]
	if !ok {
		return nil, fmt.Errorf("Invalid weekday: %v", weekday)
	}
	return subjects, nil
}
