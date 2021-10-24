package types

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

// SubLesson represents the nested lesson. During the time of one lesson, a group or teacher can have several
// SubLessons at the same time.
type SubLesson struct {
	Type    LessonType
	Group   string
	Name    string
	Teacher string
	Room    string
}

// Faculty represents the faculty.
type Faculty struct {
	Name   string
	ID     int
	Groups []string
}
