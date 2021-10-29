package types

import (
	"fmt"
	"net/http"
	"strings"
)

// Schedule represents the full schedule that contains two school Weeks.
type Schedule struct {
	Weeks [2]Week
}

// Week represents the school week (one of two schedule tables) that contains six Days (without Sunday).
type Week struct {
	Days [6]Day
}

// Day represents the school day (the row in the schedule table) that contains eight Lessons.
type Day struct {
	Lessons [8]Lesson
}

// Lesson represents the lesson (the cell in the schedule table) that can contain one or more SubLessons.
type Lesson struct {
	SubLessons []SubLesson
}

// String returns a string representation of Lesson.
func (l Lesson) String() string {
	if l.SubLessons != nil {
		var lessonBuilder strings.Builder
		_, _ = fmt.Fprintf(&lessonBuilder, "%d-ая пара (%s): ",
			int(l.SubLessons[0].Duration)+1, l.SubLessons[0].Duration.String())

		if len(l.SubLessons) == 1 {
			lessonBuilder.WriteString(l.SubLessons[0].String())
		} else {
			var subLessonsBuilder strings.Builder
			for _, subLesson := range l.SubLessons {
				if strings.Contains(subLesson.Name, subLesson.Teacher) || strings.Contains(subLesson.Name, subLesson.Room) {
					continue
				}

				subgroupLessonInfo := fmt.Sprintf("%s; ", subLesson.String())
				if !strings.Contains(subLessonsBuilder.String(), subgroupLessonInfo) {
					subLessonsBuilder.WriteString(subgroupLessonInfo)
				}
			}
			lessonBuilder.WriteString(strings.TrimSuffix(subLessonsBuilder.String(), "; "))
		}
		lessonBuilder.WriteString("\n\n")
		return lessonBuilder.String()
	}
	return ""
}

// LessonType is the type of the lesson. Can take 3 values: Lecture, Laboratory and Practice.
type LessonType int

const (
	Lecture LessonType = iota
	Laboratory
	Practice
)

func (lt LessonType) String() string {
	return [...]string{"Лек.", "Лаб.", "Пр."}[lt]
}

// Duration represents the lesson's duration.
type Duration int

func (d Duration) String() string {
	return [...]string{"08:30-09:50", "10:00-11:20", "11:30-12:50", "13:30-14:50", "15:00-16:20", "16:30-17:50",
		"18:00-19:20", "19:30-20:50"}[d]
}

// SubLesson represents the nested lesson. During the time of one lesson, a group or teacher can have several
// SubLessons at the same time.
type SubLesson struct {
	Duration Duration
	Type     LessonType
	Group    string
	Name     string
	Teacher  string
	Room     string
}

// String returns a string representation of SubLesson.
func (sl SubLesson) String() string {
	if sl.Name != "" {
		if sl.Teacher != "" {
			return fmt.Sprintf("%s %s, %s, аудитория %s", sl.Type, sl.Name, sl.Teacher, sl.Room)
		} else {
			return fmt.Sprintf("%s %s, аудитория %s", sl.Type, sl.Name, sl.Room)
		}
	}
	return ""
}

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
