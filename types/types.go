package types

// Schedule ...
type Schedule struct {
	Weeks [2]Week
}

// Week ...
type Week struct {
	Days [6]Day
}

// Day ...
type Day struct {
	Lessons [8]Lesson
}

// Lesson ...
type Lesson struct {
	SubLessons []SubLesson
}

// LessonType ...
type LessonType int

const (
	Lecture LessonType = iota
	Laboratory
	Practice
)

// SubLesson ...
type SubLesson struct {
	Type    LessonType
	Group   string
	Name    string
	Teacher string
	Room    string
}

// Faculty ...
type Faculty struct {
	Name   string
	ID     int
	Groups []string
}
