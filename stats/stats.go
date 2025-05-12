package stats

import (
	"fmt"
	"time"

	"github.com/sahaj-b/go-attend/core"
)

type StatsDataProvider interface {
	GetAttendanceItemsInRange(time.Time, time.Time) ([]core.AttendanceItem, error)
}

type Stat struct {
	Attended int
	Total    int
}

type (
	subjectStatsMap map[string]Stat
	weekdayStatsMap map[string]Stat
)

func GetSubjectWiseStats(dp StatsDataProvider, startDate time.Time, endDate time.Time) (subjectStatsMap, int, int, error) {
	attended, total := 0, 0
	items, err := dp.GetAttendanceItemsInRange(startDate, endDate)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("Failed to get attendance items: %w", err)
	}
	subjectStats := make(subjectStatsMap)
	for _, item := range items {
		if item.Status == core.Cancelled {
			continue
		}
		currStat := subjectStats[item.Subject]
		if item.Status == core.Present {
			attended++
			currStat.Attended++
		}
		total++
		currStat.Total++
		subjectStats[item.Subject] = currStat
	}
	return subjectStats, attended, total, nil
}

func GetWeekdayWiseStats(dp StatsDataProvider, startDate time.Time, endDate time.Time) (weekdayStatsMap, int, int, error) {
	attended, total := 0, 0
	items, err := dp.GetAttendanceItemsInRange(startDate, endDate)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("Failed to get attendance items: %w", err)
	}
	weekdayStats := make(weekdayStatsMap)
	for _, item := range items {
		weekdayKey := item.Date.Weekday().String()
		currStat := weekdayStats[weekdayKey]
		if item.Status == core.Cancelled {
			continue
		}
		if item.Status == core.Present {
			attended++
			currStat.Attended++
		}
		total++
		currStat.Total++
		weekdayStats[weekdayKey] = currStat
	}
	return weekdayStats, attended, total, nil
}
