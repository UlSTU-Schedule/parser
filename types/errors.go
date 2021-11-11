package types

import (
	"fmt"
	"net/http"
)

// StatusCodeError is returned when a http.Get returns a response with a status code other than 200.
type StatusCodeError struct {
	StatusCode int
	StatusText string
}

// StatusCodeError implements Error method.
func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("status code error: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
}

// IncorrectDateError is returned when the date does not match the "dd.mm" format or does not exist.
type IncorrectDateError struct {
	Date string
}

// IncorrectDateError implements Error method.
func (e *IncorrectDateError) Error() string {
	return fmt.Sprintf("incorrect date: %s", e.Date)
}

// IncorrectWeekNumberError is returned when the value of the school week number is out of the acceptable range.
type IncorrectWeekNumberError struct {
	WeekNum int
}

// IncorrectWeekNumberError implements Error method.
func (e *IncorrectWeekNumberError) Error() string {
	return fmt.Sprintf("incorrect value of the school week number: %d", e.WeekNum)
}

// UnavailableWeeklyScheduleError is returned when the weekly schedule is missing or not published.
type UnavailableWeeklyScheduleError struct {
	Object  string
	WeekNum int
}

// UnavailableWeeklyScheduleError implements Error method.
func (e *UnavailableWeeklyScheduleError) Error() string {
	return fmt.Sprintf("the weekly schedule is missing or not published: %s, week number is %d",
		e.Object, e.WeekNum)
}

// LinkPointsToIncorrectObjectError is returned when when the schedule on the link does not match the expected schedule.
type LinkPointsToIncorrectObjectError struct {
	ObjectName        string
	ObjectNameFromURL string
}

// LinkPointsToIncorrectObjectError implements Error method.
func (e *LinkPointsToIncorrectObjectError) Error() string {
	return fmt.Sprintf("mismatch between schedule objects: %s != %s", e.ObjectName, e.ObjectNameFromURL)
}
