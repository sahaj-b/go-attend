package ui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/sahaj-b/go-attend/stats"
)

const (
	bar          = "🬋"
	maxBarLength = 50
)

func barComponent(coloredLength, maxBarLength int) string {
	grayBarLength := maxBarLength - coloredLength
	// return Bggray + Cyan + strings.Repeat(bar, coloredLength) + strings.Repeat(" ", grayBarLength) + ResetStyle + "\n"
	// return Bggray + Cyan + strings.Repeat(bar, coloredLength) + Gray + strings.Repeat(bar, grayBarLength) + ResetStyle + "\n"
	return Cyan + strings.Repeat(bar, coloredLength) + MoreGray + strings.Repeat(bar, grayBarLength) + ResetStyle + "\n"
}

func overallAttendanceComponent(attended, total int) string {
	percentage := float32(attended) / float32(total) * 100
	return headerComponent("Overall Attendance") +
		Yellow + "Percentage: " + ResetStyle + Cyan + Bold + Bggray + fmt.Sprintf(" %.1f%% ", percentage) + ResetStyle + "\n" +
		Yellow + "Classes attended " + ResetStyle + Cyan + Bold + Bggray + fmt.Sprintf(" %d/%d ", attended, total) + ResetStyle + "\n" +
		barComponent(int(percentage*maxBarLength/100), maxBarLength)
}

func barMapComponent(imap map[string]stats.Stat, weekday bool) string {
	keys := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	if !weekday {
		keys = make([]string, 0, len(imap))
		for k := range imap {
			keys = append(keys, k)
		}
		slices.Sort(keys)
	}

	output := strings.Builder{}
	for _, key := range keys {
		stat, exists := imap[key]
		if !exists {
			continue
		}
		subjectPercentage := float32(0)
		if stat.Total > 0 {
			subjectPercentage = float32(stat.Attended) / float32(stat.Total) * 100
		}
		output.WriteString(Bggray + Yellow + Bold + " " + key + " " + ResetStyle +
			Cyan + Bold + fmt.Sprintf(" %.1f%%\n", subjectPercentage) + ResetStyle +
			Yellow + " " + strconv.Itoa(stat.Attended) + "/" + strconv.Itoa(stat.Total) + " " + ResetStyle +
			barComponent(int(subjectPercentage*maxBarLength/100), maxBarLength) + "\n")
	}
	return output.String()
}

func headerComponent(header string) string {
	return Bggray + Yellow + Bold + " " + header + " " + ResetStyle + "\n\n"
}

func DisplaySubjectWiseStats(dp stats.StatsDataProvider, startDate, endDate time.Time) {
	subjectsMap, attended, total, err := stats.GetSubjectWiseStats(dp, startDate, endDate)
	output := strings.Builder{}
	if err != nil {
		Error("Error fetching stats: " + err.Error())
		return
	}
	if total == 0 || len(subjectsMap) == 0 {
		Warn("No attendance records found")
		return
	}
	output.WriteString("\n")
	output.WriteString(headerComponent("Subject Wise Attendance"))
	output.WriteString(barMapComponent(subjectsMap, false) + "\n")
	output.WriteString(overallAttendanceComponent(attended, total))
	fmt.Println(output.String())
}

func DisplayWeekdayWiseStats(dp stats.StatsDataProvider, startDate, endDate time.Time) {
	weekdaysMap, attended, total, err := stats.GetWeekdayWiseStats(dp, startDate, endDate)
	output := strings.Builder{}
	if err != nil {
		Error("Error fetching stats: " + err.Error())
		return
	}
	if total == 0 || len(weekdaysMap) == 0 {
		Warn("No attendance records found")
		return
	}

	output.WriteString("\n")
	output.WriteString(headerComponent("Weekday Wise Attendance"))
	output.WriteString(barMapComponent(weekdaysMap, true))
	output.WriteString("\n")
	output.WriteString(overallAttendanceComponent(attended, total))
	fmt.Println(output.String())
}
