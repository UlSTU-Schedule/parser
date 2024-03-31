package schedule

import (
	_ "embed"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/ulstu-schedule/parser/types"
	"strings"
	"time"
)

const (
	teacherScheduleURL          = "https://lk.ulstu.ru/timetable/shared/teachers/%s"
	headingTableTeacherFontSize = 38
)

// GetFullTeacherSchedule returns the full teacher's schedule.
func GetFullTeacherSchedule(teacher string) (*types.Schedule, error) {
	teacherURL, err := getTeacherURL(teacher)
	if err != nil {
		return nil, err
	}
	return getFullSchedule(teacher, teacherURL, types.Teacher)
}

func GetTeachers() ([]string, error) {
	teachers := make([]string, 0, 800)

	doc, err := getDocFromURL(fmt.Sprintf(teacherScheduleURL, "Praspisan.html"))
	if err != nil {
		return nil, err
	}

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		if i > 0 {
			foundTeacherName := s.Find("font").Text()
			formattedTeacherName := strings.Split(foundTeacherName, ",")[0]
			teachers = append(teachers, formattedTeacherName)
		}
	})

	return teachers, nil
}

// GetTextDayTeacherSchedule returns a text representation of the day schedule.
func GetTextDayTeacherSchedule(teacherName string, daysAfterCurr int) (string, error) {
	schedule, err := GetDayTeacherSchedule(teacherName, daysAfterCurr)
	if err != nil {
		return "", err
	}

	return ConvertDayTeacherScheduleToText(teacherName, *schedule, daysAfterCurr), nil
}

// GetCurrWeekTeacherScheduleImg return img of current week schedule
func GetCurrWeekTeacherScheduleImg(teacherName string) (string, error) {
	currWeekDate, _ := getWeekDateAndWeekDay(0)
	return GetWeekTeacherScheduleImg(teacherName, currWeekDate, true)
}

// GetNextWeekTeacherScheduleImg return img of next week schedule
func GetNextWeekTeacherScheduleImg(teacherName string) (string, error) {
	nextWeekDate, _ := getWeekDateAndWeekDay(7)
	return GetWeekTeacherScheduleImg(teacherName, nextWeekDate, false)
}

// GetCurrWeekTeacherSchedule return object of current week schedule
func GetCurrWeekTeacherSchedule(teacherName string) (*types.Week, error) {
	currWeekDate, _ := getWeekDateAndWeekDay(0)
	return GetWeekTeacherSchedule(teacherName, currWeekDate)
}

// GetNextWeekTeacherSchedule return object of next week schedule
func GetNextWeekTeacherSchedule(teacherName string) (*types.Week, error) {
	nextWeekDate, _ := getWeekDateAndWeekDay(7)
	return GetWeekTeacherSchedule(teacherName, nextWeekDate)
}

// GetWeekTeacherScheduleImg return path on img of schedule
func GetWeekTeacherScheduleImg(teacherName string, weekDate time.Time, isCurrWeek bool) (string, error) {
	schedule, err := GetWeekTeacherSchedule(teacherName, weekDate)
	if err != nil {
		return "", err
	}
	return getImgByWeekTeacherSchedule(schedule, teacherName, isCurrWeek)
}

// GetDayTeacherSchedule returns *types.Day received from the full schedule regarding how many days have passed relative to the current time.
func GetDayTeacherSchedule(teacherName string, daysAfterCurr int) (*types.Day, error) {
	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}
	return ParseDaySchedule(schedule, teacherName, daysAfterCurr)
}

// GetWeekTeacherSchedule return object of week schedule
func GetWeekTeacherSchedule(teacherName string, weekDate time.Time) (*types.Week, error) {
	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}

	return parseWeekSchedule(schedule, teacherName, weekDate)
}

// ParseCurrWeekTeacherScheduleImg returns the path to the image with the week schedule based on the current school week.
func ParseCurrWeekTeacherScheduleImg(schedule *types.Week, teacherName string) (string, error) {
	return getImgByWeekTeacherSchedule(schedule, teacherName, true)
}

// ParseNextWeekTeacherScheduleImg returns the path to the image with the week schedule based on the next school week.
func ParseNextWeekTeacherScheduleImg(schedule *types.Week, teacherName string) (string, error) {
	return getImgByWeekTeacherSchedule(schedule, teacherName, false)
}

// GetImgByWeekTeacherSchedule return path on img of schedule
func getImgByWeekTeacherSchedule(schedule *types.Week, teacherName string, isCurrWeek bool) (string, error) {
	return GetImgByWeekSchedule(schedule, teacherName, isCurrWeek, headingTableTeacherFontSize, drawTeacherLessonForWeekSchedule)
}

// ConvertDayTeacherScheduleToText converts the information that types.Day contains into text.
func ConvertDayTeacherScheduleToText(teacherName string, daySchedule types.Day, daysAfterCurr int) string {
	result := strings.Builder{}

	dateStr := getDateStr(daysAfterCurr)
	_, weekDayNum := getWeekDateAndWeekDay(daysAfterCurr)

	switch daysAfterCurr {
	case 0:
		_, _ = fmt.Fprintf(&result, "%s проводит следующие пары сегодня (%s, %s, %d-ая учебная неделя):\n\n",
			teacherName, weekDays[weekDayNum], dateStr, daySchedule.WeekNumber)
	case 1:
		_, _ = fmt.Fprintf(&result, "%s проводит следующие пары завтра (%s, %s, %d-ая учебная неделя):\n\n",
			teacherName, weekDays[weekDayNum], dateStr, daySchedule.WeekNumber)
	default:
		_, _ = fmt.Fprintf(&result, "%s проводит следующие пары %s (%s, %d-ая учебная неделя):\n\n", teacherName,
			dateStr, weekDays[weekDayNum], daySchedule.WeekNumber)
	}

	noLessons := true
	for lessonIndex := 0; lessonIndex < len(daySchedule.Lessons); lessonIndex++ {
		subLessons := daySchedule.Lessons[lessonIndex].SubLessons

		if len(subLessons) > 0 {
			noLessons = false
			result.WriteString(daySchedule.Lessons[lessonIndex].StringTeacherLesson())
		}
	}

	if noLessons {
		switch daysAfterCurr {
		case 0:
			result.WriteString("Сегодня пар нет")
		case 1:
			result.WriteString("Завтра пар нет")
		default:
			_, _ = fmt.Fprintf(&result, "%s пар нет", dateStr)
		}
	}

	return result.String()
}

// parseTeacherLesson returns *types.Lesson received from the HTML document.
func parseTeacherLesson(teacher string, lessonIdx int, s *goquery.Selection) *types.Lesson {
	lesson := types.Lesson{}
	tableCellHTML, _ := s.Find("font").Html()
	// if the table cell contains the lesson info
	if !strings.HasPrefix(tableCellHTML, "_") && tableCellHTML != "" {
		// <br/> separates the name of the lesson, the groups and the audience number
		splitLessonInfoHTML := strings.Split(tableCellHTML, "<br/>")
		lessonGroups := strings.Split(splitLessonInfoHTML[0], ",")
		lesson.SubLessons = make([]types.SubLesson, 0, len(lessonGroups))
		lessonTypeAndName := strings.Split(splitLessonInfoHTML[1], ".")
		lessonType := determineLessonType(lessonTypeAndName[0])
		r := strings.NewReplacer(",", ", ", ".", ". ", "- ", " - ", " -", " - ") // TODO: вынести в самый верх, чтобы постоянно его не создавать (см. пример groups.go)
		lessonName := r.Replace(strings.TrimSpace(lessonTypeAndName[len(lessonTypeAndName)-1]))
		for _, groupName := range lessonGroups {
			r = strings.NewReplacer(".", "", "_", "-", " - ", "-", " -", "-", "- ", "-") // TODO: вынести в самый верх, чтобы постоянного его не создавать (см. пример groups.go)
			groupLesson := types.SubLesson{
				Duration: types.Duration(lessonIdx),
				Type:     lessonType,
				Group:    groupName,
				Name:     lessonName,
				Teacher:  teacher,
				Room:     r.Replace(splitLessonInfoHTML[2]),
			}
			lesson.SubLessons = append(lesson.SubLessons, groupLesson)
		}
	}

	return &lesson
}

// getTeacherURL returns the url to the teacher's schedule on UlSTU site.
func getTeacherURL(teacherName string) (string, error) {
	teacherURL := ""

	doc, err := getDocFromURL(fmt.Sprintf(teacherScheduleURL, "Praspisan.html"))
	if err != nil {
		return "", err
	}

	doc.Find("td").EachWithBreak(func(i int, s *goquery.Selection) bool {
		foundTeacherName := s.Find("font").Text()
		formattedTeacherName := strings.Split(foundTeacherName, ",")[0]
		if formattedTeacherName == teacherName {
			url, _ := s.Find("a").Attr("href")
			teacherURL = fmt.Sprintf(teacherScheduleURL, url)
			return false
		}
		return true
	})

	if teacherURL != "" {
		doc, err = getDocFromURL(teacherURL)
		if err != nil {
			return "", err
		}

		teacherNameFromDoc := doc.Find("p").Get(0).LastChild.FirstChild.Data
		if !strings.Contains(teacherNameFromDoc, teacherName) {
			return "", &types.IncorrectLinkError{Name: teacherName, NameFromURL: teacherNameFromDoc}
		}
	}

	return teacherURL, nil
}

// drawLessonForWeekSchedule - rendering schedule of lesson
func drawTeacherLessonForWeekSchedule(lesson *types.Lesson, x, y float64, dc *gg.Context) {
	subLessons := lesson.SubLessons

	groups := lesson.GetGroupsTeacherLesson()

	infoAboutLesson := fmt.Sprintf("%s \n%s %s \nаудитория %s", groups, subLessons[0].Type.String(),
		subLessons[0].Name, subLessons[0].Room)

	wrappedInfoStr := dc.WordWrap(infoAboutLesson, cellWidth-20)

	hasFontChanged := false

	linesInLessonStr := len(wrappedInfoStr)
	if linesInLessonStr >= 6 {
		setFontSize(linesInLessonStr, dc)
		hasFontChanged = true
	}

	dc.DrawStringWrapped(infoAboutLesson, x, y-143, 0, 0, cellWidth-20, 1.3, 1)

	if hasFontChanged {
		setFont(defaultScheduleFontSize, dc)
	}
}
