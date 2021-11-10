//go:generate easyjson -no_std_marshalers types.go

package types

import (
	"fmt"
	"strings"
)

// Schedule represents the full schedule that contains two school Weeks.
//easyjson:json
type Schedule struct {
	Weeks [2]Week `json:"weeks"`
}

// Week represents the school week (one of two schedule tables) that contains six Days (without Sunday).
type Week struct {
	Days [7]Day `json:"days"`
}

// Day represents the school day (the row in the schedule table) that contains eight Lessons.
type Day struct {
	Lessons [8]Lesson `json:"lessons"`
}

// Lesson represents the lesson (the cell in the schedule table) that can contain one or more SubLessons.
type Lesson struct {
	SubLessons []SubLesson `json:"sub_lessons"`
}

// StringGroupLesson returns a string representation of Lesson based on the structure of the lesson display for groups.
func (l Lesson) StringGroupLesson() string {
	if l.SubLessons != nil {
		var lessonBuilder strings.Builder
		_, _ = fmt.Fprintf(&lessonBuilder, "%d-ая пара (%s): ",
			int(l.SubLessons[0].Duration)+1, l.SubLessons[0].Duration.String())

		if len(l.SubLessons) == 1 {
			lessonBuilder.WriteString(l.SubLessons[0].StringGroupSubLesson())
		} else {
			var subLessonsBuilder strings.Builder
			for _, subLesson := range l.SubLessons {
				if strings.Contains(subLesson.Name, subLesson.Teacher) || strings.Contains(subLesson.Name, subLesson.Room) {
					continue
				}

				subgroupLessonInfo := fmt.Sprintf("%s; ", subLesson.StringGroupSubLesson())
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

func (l Lesson) StringTeacherLesson() string {
	if l.SubLessons != nil {

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
	Duration Duration   `json:"duration"`
	Type     LessonType `json:"type"`
	Group    string     `json:"group"`
	Name     string     `json:"name"`
	Teacher  string     `json:"teacher"`
	Room     string     `json:"room"`
}

// StringGroupSubLesson returns a string representation of SubLesson based on the structure of the lesson display for groups.
func (sl SubLesson) StringGroupSubLesson() string {
	if sl.Name != "" {
		if sl.Teacher != "" {
			return fmt.Sprintf("%s %s, %s, аудитория %s", sl.Type, sl.Name, sl.Teacher, sl.Room)
		} else {
			return fmt.Sprintf("%s %s, аудитория %s", sl.Type, sl.Name, sl.Room)
		}
	}
	return ""
}
