package types

// GroupSchedule ...
type GroupSchedule struct {
	Weeks []GroupWeek
}

// GroupWeek ...
type GroupWeek struct {
	Days []GroupDay
}

// GroupDay ...
type GroupDay struct {
	Lessons []GroupLesson
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
	Weeks [2]TeacherWeek
}

// TeacherWeek ...
type TeacherWeek struct {
	Days [6]TeacherDay
}

// TeacherDay ...
type TeacherDay struct {
	Lessons [8]TeacherLesson
}

// TeacherLesson ...
type TeacherLesson struct {
	Groups []TeacherGroupLesson
}

// TeacherGroupLesson ...
type TeacherGroupLesson struct {
	Type    LessonType
	Group  string
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
