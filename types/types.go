package types

// GroupSchedule ...
type GroupSchedule struct {
	Weeks [2]GroupWeek
}

// GroupWeek ...
type GroupWeek struct {
	Days [6]GroupDay
}

// GroupDay ...
type GroupDay struct {
	Lessons [8]GroupLesson
}

// GroupLesson ...
type GroupLesson struct {
	SubLessons []GroupSubLesson
}

// LessonType ...
type LessonType int

const (
	Lecture LessonType = iota
	Laboratory
	Practice
)

// GroupSubLesson ...
type GroupSubLesson struct {
	Type    LessonType
	Group   string
	Name    string
	Teacher string
	Room    string
}

// TeacherSchedule ...
type TeacherSchedule struct {
	Weeks []TeacherWeek
}

// TeacherWeek ...
type TeacherWeek struct {
	Days []TeacherDay
}

// TeacherDay ...
type TeacherDay struct {
	Lessons []TeacherLesson
}

// TeacherLesson ...
type TeacherLesson struct {
	Type    LessonType
	Groups  []string
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
