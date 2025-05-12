package core

import "time"

type AttendanceStatus int

const (
	Present AttendanceStatus = iota
	Absent
	Cancelled
)

type AttendanceItem struct {
	Subject string
	Status  AttendanceStatus
	Date    time.Time
}
