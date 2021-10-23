package types

type (
	// Schedule represents the full schedule that contains two school Weeks.
	Schedule struct {
		Weeks [2]Week
	}

	// Week represents the school week (one of two schedule tables) that contains six Days (without Sunday).
	Week struct {
		Days [6]Day
	}

	// Day represents the school day (the row in the schedule table) that contains eight Lessons.
	Day struct {
		Lessons [8]Lesson
	}

	// Lesson represents the lesson (the cell in the schedule table) that can contain one or more SubLessons.
	Lesson struct {
		SubLessons []SubLesson
	}

	// LessonType is the type of the lesson. Can take 3 values: Lecture, Laboratory and Practice.
	LessonType int

	// SubLesson represents the nested lesson. During the time of one lesson, a group or teacher can have several
	// SubLessons at the same time.
	SubLesson struct {
		Type    LessonType
		Group   string
		Name    string
		Teacher string
		Room    string
	}

	// Faculty represents the faculty.
	Faculty struct {
		Name   string
		ID     int
		Groups []string
	}
)

const (
	Lecture LessonType = iota
	Laboratory
	Practice
)
