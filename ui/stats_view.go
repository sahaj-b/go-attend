package ui

import (
	"fmt"
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
	// output.WriteString(Bggray + Cyan + strings.Repeat(bar, coloredBarLength) + strings.Repeat(" ", grayBarLength) + ResetStyle + "\n")
	// output.WriteString(Bggray + Cyan + strings.Repeat(bar, coloredBarLength) + Gray + strings.Repeat(bar, grayBarLength) + ResetStyle + "\n")
	return Cyan + strings.Repeat(bar, coloredLength) + Gray + strings.Repeat(bar, grayBarLength) + ResetStyle + "\n"
}

func overallAttendanceComponent(attended, total int) string {
	percentage := float32(attended) / float32(total) * 100
	return Bggray + Yellow + Bold + "Overall Attendance: " + ResetStyle +
		Bggray + Yellow + strconv.Itoa(attended) + "/" + strconv.Itoa(total) + " Classes attended" + ResetStyle + "\n" +
		Yellow + Bold + fmt.Sprintf("Attendance Percentage: %.1f%%\n", percentage) + ResetStyle +
		barComponent(int(percentage*maxBarLength/100), maxBarLength)
}

func barMapComponent(imap map[string]stats.Stat) string {
	output := strings.Builder{}
	for subject, stats := range imap {
		subjectPercentage := float32(stats.Attended) / float32(stats.Total) * 100
		output.WriteString(Bggray + Yellow + Bold + subject + ResetStyle + "\n")
		output.WriteString(Bggray + Yellow + strconv.Itoa(stats.Attended) + "/" + strconv.Itoa(stats.Total) + " Classes attended" + ResetStyle + "\n")
		output.WriteString(Yellow + Bold + fmt.Sprintf("Attendance Percentage: %.1f%%\n", subjectPercentage) + ResetStyle)
		output.WriteString(barComponent(int(subjectPercentage*maxBarLength/100), maxBarLength))
	}
	return output.String()
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
	output.WriteString(overallAttendanceComponent(attended, total))
	output.WriteString(Bggray + Yellow + Bold + "Subject Wise Stats" + ResetStyle)
	output.WriteString(barMapComponent(subjectsMap))
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

	output.WriteString(overallAttendanceComponent(attended, total))
	output.WriteString(Bggray + Yellow + Bold + "Weekday Wise Stats" + ResetStyle)
	output.WriteString(barMapComponent(weekdaysMap))
	fmt.Println(output.String())
}
