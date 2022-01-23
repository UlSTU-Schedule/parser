package schedule

import (
	_ "embed"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/ulstu-schedule/parser/types"
	"strconv"
	"strings"
	"time"
)

const (
	teacherScheduleURL          = "https://old.ulstu.ru/schedule/teachers/%s"
	headingTableTeacherFontSize = 38
)

//go:embed assets/week_schedule_teacher_template.png
var weekScheduleTeacherTemp []byte

// GetTextDayTeacherScheduleByDate returns a text representation of the day schedule based on the the string
// representation of the date.
func GetTextDayTeacherScheduleByDate(teacherName, inputDate string) (string, error) {
	schedule, err := GetDayTeacherScheduleByDate(teacherName, inputDate)
	if err != nil {
		return "", err
	}

	inputDateTime, err := getDateTime(inputDate)
	if err != nil {
		return "", err
	}

	nowDateTime := time.Now()

	// The difference in days between the entered date and the current date. Negative if the entered date is earlier
	// than the current one, and positive if the entered date is later than the current one.
	diffBetweenInputAndCurrDates := int(inputDateTime.Sub(nowDateTime).Hours() / 24)
	if nowDateTime.Before(inputDateTime) {
		diffBetweenInputAndCurrDates++
	}

	return convertDayTeacherScheduleToText(teacherName, *schedule, diffBetweenInputAndCurrDates), nil
}

// GetDayTeacherScheduleByDate returns *types.Day received from the full schedule based on the string representation
// of the date.
func GetDayTeacherScheduleByDate(teacherName, date string) (*types.Day, error) {
	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}

	weekNum, weekDayNum, err := getWeekAndWeekDayNumbersByDate(date)
	if err != nil {
		return nil, err
	}

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: teacherName, WeekNum: weekNum}
	}

	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDayTeacherScheduleByWeekDay returns a text representation of the day schedule based on the selected day of
// the current week.
func GetTextDayTeacherScheduleByWeekDay(teacherName, weekDay string) (string, error) {
	schedule, err := GetDayTeacherScheduleByWeekDay(teacherName, weekDay)
	if err != nil {
		return "", err
	}

	weekDayNumNow := int(time.Now().Weekday()) - 1
	if weekDayNumNow == -1 {
		weekDayNumNow = 6
	}
	weekDayNum := convertWeekDayToWeekDayIdx(weekDay)

	return convertDayTeacherScheduleToText(teacherName, *schedule, weekDayNum-weekDayNumNow), nil
}

// GetDayTeacherScheduleByWeekDay returns *types.Day received from the full schedule based on the selected day of the
// current week.
func GetDayTeacherScheduleByWeekDay(teacherName, weekDay string) (*types.Day, error) {
	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}

	weekNum, weekDayNum := getWeekAndWeekDayNumbersByWeekDay(weekDay)

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: teacherName, WeekNum: weekNum}
	}

	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDayTeacherSchedule returns a text representation of the day schedule.
func GetTextDayTeacherSchedule(teacherName string, daysAfterCurr int) (string, error) {
	schedule, err := GetDayTeacherSchedule(teacherName, daysAfterCurr)
	if err != nil {
		return "", err
	}

	return convertDayTeacherScheduleToText(teacherName, *schedule, daysAfterCurr), nil
}

// GetDayTeacherSchedule returns *types.Day received from the full schedule regarding how many days have passed relative to the current time.
func GetDayTeacherSchedule(teacherName string, daysAfterCurr int) (*types.Day, error) {
	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}

	weekNum, weekDayNum := getWeekAndWeekDayNumbers(daysAfterCurr)

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: teacherName, WeekNum: weekNum}
	}

	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetCurrWeekTeacherScheduleImg return img of current week schedule
func GetCurrWeekTeacherScheduleImg(teacherName string) (string, error) {
	currWeekNum, _ := getWeekAndWeekDayNumbers(0)
	return GetWeekTeacherScheduleImg(teacherName, currWeekNum)
}

// GetNextWeekTeacherScheduleImg return img of next week schedule
func GetNextWeekTeacherScheduleImg(teacherName string) (string, error) {
	currWeekNum, _ := getWeekAndWeekDayNumbers(7)
	return GetWeekTeacherScheduleImg(teacherName, currWeekNum)
}

// GetCurrWeekTeacherSchedule return object of current week schedule
func GetCurrWeekTeacherSchedule(teacherName string) (*types.Week, error) {
	currWeekNum, _ := getWeekAndWeekDayNumbers(0)
	return GetWeekTeacherSchedule(teacherName, currWeekNum)
}

// GetNextWeekTeacherSchedule return object of next week schedule
func GetNextWeekTeacherSchedule(teacherName string) (*types.Week, error) {
	nextWeekNum, _ := getWeekAndWeekDayNumbers(7)
	return GetWeekTeacherSchedule(teacherName, nextWeekNum)
}

// GetWeekTeacherSchedule return object of week schedule
func GetWeekTeacherSchedule(teacherName string, weekNum int) (*types.Week, error) {
	if weekNum < 0 || weekNum > 1 {
		return nil, &types.IncorrectWeekNumberError{WeekNum: weekNum}
	}

	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: teacherName, WeekNum: weekNum}
	}

	return &schedule.Weeks[weekNum], nil
}

// GetWeekTeacherScheduleImg return path on img of schedule
func GetWeekTeacherScheduleImg(teacherName string, weekNum int) (string, error) {
	schedule, err := GetWeekTeacherSchedule(teacherName, weekNum)
	if err != nil {
		return "", err
	}

	return ParseTeacherWeekSchedule(schedule, teacherName, weekNum)
}

// ParseTeacherWeekSchedule return path on img of schedule
func ParseTeacherWeekSchedule(schedule *types.Week, teacherName string, weekNum int) (string, error) {
	// loads an template of an empty table that will be filled in pairs
	tableImg := getWeekScheduleTmplImg(weekScheduleTeacherTemp)
	dc := gg.NewContextForImage(tableImg)

	setFont(headingTableTeacherFontSize, dc)
	dc.SetRGB255(25, 89, 209)
	dc.DrawString(teacherName, 515, 60)
	dc.DrawString(fmt.Sprintf("%d-ая", weekNum+1), imgWidth-105, 60)

	setDefaultSettings(dc)

	currWeekNum, currWeekDayNum := getWeekAndWeekDayNumbers(0)

	for row := 352; row < imgHeight; row += cellHeight {
		rowNum := row/cellHeight - 2

		x, y := float64(130), float64(row)

		if rowNum == currWeekDayNum && currWeekDayNum != -1 && currWeekNum == weekNum {
			highlightRow(row, dc)
		}
		scheduleDay := schedule.Days[rowNum]

		for _, lesson := range scheduleDay.Lessons {
			if len(lesson.SubLessons) > 0 {
				drawLessonForWeekSchedule(&lesson, dc, x, y)
			}
			// переходит к следующей паре
			x += cellWidth
		}
	}

	weekSchedulePath := fmt.Sprintf("week_schedule%d.png", getRandInt())
	return weekSchedulePath, dc.SavePNG(weekSchedulePath)
}

// drawLessonForWeekSchedule - rendering schedule of lesson
func drawLessonForWeekSchedule(lesson *types.Lesson, dc *gg.Context, x, y float64) {
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

// GetFullTeacherSchedule returns the full teacher's schedule.
func GetFullTeacherSchedule(teacher string) (*types.Schedule, error) {
	teacherURL, err := getTeacherURL(teacher)
	if err != nil {
		return nil, err
	}

	doc, err := getDocFromURL(teacherURL)
	if err != nil {
		return nil, err
	}

	teacherSchedule := new(types.Schedule)

	pSelection := doc.Find("p")

	if pSelection.Length() == 182 {
		pSelection.Each(func(i int, s *goquery.Selection) {
			// first week lessons
			if 22 <= i && i <= 79 && 2 <= i%10 && i%10 <= 9 {
				dayIdx := i/10 - 2
				lessonIdx := i%10 - 2
				teacherSchedule.Weeks[0].Days[dayIdx].Lessons[lessonIdx] = *getTeacherLessonFromDoc(teacher, lessonIdx, s)
			}
			// second week lessons
			if 113 <= i && i <= 170 && (i%10 == 0 || i%10 >= 3) {
				var dayIdx, lessonIdx int
				if i%10 == 0 {
					lessonIdx = 7
				} else {
					lessonIdx = i%10 - 3
				}
				if i == 170 {
					dayIdx = 5
				} else {
					dayIdx = (i/10)%10 - 1
				}
				teacherSchedule.Weeks[1].Days[dayIdx].Lessons[lessonIdx] = *getTeacherLessonFromDoc(teacher, lessonIdx, s)
			}
		})
	} else {
		weekNumStr := pSelection.Get(0).LastChild.LastChild.Data
		weekNum, _ := strconv.Atoi(string(strings.Split(weekNumStr, ": ")[1][0]))

		pSelection.Each(func(i int, s *goquery.Selection) {
			if 22 <= i && i <= 79 && 2 <= i%10 && i%10 <= 9 {
				dayIdx := i/10 - 2
				lessonIdx := i%10 - 2
				teacherSchedule.Weeks[weekNum-1].Days[dayIdx].Lessons[lessonIdx] = *getTeacherLessonFromDoc(teacher, lessonIdx, s)
			}
		})
	}

	return teacherSchedule, nil
}

// convertDayTeacherScheduleToText converts the information that types.Day contains into text.
func convertDayTeacherScheduleToText(teacherName string, daySchedule types.Day, daysAfterCurr int) string {
	result := strings.Builder{}

	dateStr := getDateStr(daysAfterCurr)
	weekNum, weekDayNum := getWeekAndWeekDayNumbers(daysAfterCurr)

	switch daysAfterCurr {
	case 0:
		_, _ = fmt.Fprintf(&result, "%s проводит следующие пары сегодня (%s, %s, %d-ая учебная неделя):\n\n",
			teacherName, weekDays[weekDayNum], dateStr, weekNum+1)
	case 1:
		_, _ = fmt.Fprintf(&result, "%s проводит следующие пары завтра (%s, %s, %d-ая учебная неделя):\n\n",
			teacherName, weekDays[weekDayNum], dateStr, weekNum+1)
	default:
		_, _ = fmt.Fprintf(&result, "%s проводит следующие пары %s (%s, %d-ая учебная неделя):\n\n", teacherName,
			dateStr, weekDays[weekDayNum], weekNum+1)
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

// getTeacherLessonFromDoc returns *types.Lesson received from the HTML document.
func getTeacherLessonFromDoc(teacher string, lessonIdx int, s *goquery.Selection) *types.Lesson {
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
