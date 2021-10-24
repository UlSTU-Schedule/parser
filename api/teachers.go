package api

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"strings"
	"time"
)

const teacherScheduleURL = "https://old.ulstu.ru/schedule/teachers/%s"

// GetTextDailyTeacherScheduleByDate returns a text representation of the daily schedule based on the the string
// representation of the date.
func GetTextDailyTeacherScheduleByDate(teacherName, inputDate string) (string, error) {
	schedule, err := GetDailyTeacherScheduleByDate(teacherName, inputDate)
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
	return convertDailyTeacherScheduleToText(teacherName, *schedule, diffBetweenInputAndCurrDates), nil
}

// GetDailyTeacherScheduleByDate returns *types.Day received from the full schedule based on the string representation
// of the date.
func GetDailyTeacherScheduleByDate(teacherName, date string) (*types.Day, error) {
	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}

	weekNum, weekDayNum, err := getWeekAndWeekDayNumbersByDate(date)
	if err != nil {
		return nil, err
	}

	// returns weekDayNum = -1 when the day of the week is Sunday
	if weekDayNum == -1 {
		return &types.Day{}, nil
	}

	if isWeeklyScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, errors.New("the schedule for the selected week is empty or not loaded yet")
	}
	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDailyTeacherScheduleByWeekDay returns a text representation of the daily schedule based on the selected day of
// the current week.
func GetTextDailyTeacherScheduleByWeekDay(teacherName, weekDay string) (string, error) {
	schedule, err := GetDailyTeacherScheduleByWeekDay(teacherName, weekDay)
	if err != nil {
		return "", err
	}

	weekDayNumNow := int(time.Now().Weekday()) - 1
	weekDayNum := convertWeekdayToWeekDayIdx(weekDay)
	if weekDayNum == -1 && weekDayNumNow != -1 {
		weekDayNum = 6
	}
	return convertDailyTeacherScheduleToText(teacherName, *schedule, weekDayNum-weekDayNumNow), nil
}

// GetDailyTeacherScheduleByWeekDay returns *types.Day received from the full schedule based on the selected day of the
// current week.
func GetDailyTeacherScheduleByWeekDay(teacherName, weekDay string) (*types.Day, error) {
	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}

	weekNum, weekDayNum := getWeekAndWeekDayNumbersByWeekDay(weekDay)
	// returns weekDayNum = -1 when the day of the week is Sunday
	if weekDayNum == -1 {
		return &types.Day{}, nil
	}

	if isWeeklyScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, errors.New("the schedule for the selected week is empty or not loaded yet")
	}
	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDailyTeacherSchedule returns a text representation of the daily schedule.
func GetTextDailyTeacherSchedule(teacherName string, daysAfterCurr int) (string, error) {
	schedule, err := GetDailyTeacherSchedule(teacherName, daysAfterCurr)
	if err != nil {
		return "", err
	}
	return convertDailyTeacherScheduleToText(teacherName, *schedule, daysAfterCurr), nil
}

// GetDailyTeacherSchedule returns *types.Day received from the full schedule regarding how many days have passed relative to the current time.
func GetDailyTeacherSchedule(teacherName string, daysAfterCurr int) (*types.Day, error) {
	schedule, err := GetFullTeacherSchedule(teacherName)
	if err != nil {
		return nil, err
	}

	weekNum, weekDayNum := getWeekAndWeekDayNumbers(daysAfterCurr)
	// returns weekDayNum = -1 when the day of the week is Sunday
	if weekDayNum == -1 {
		return &types.Day{}, nil
	}

	if isWeeklyScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, errors.New("the schedule for the selected week is empty or not loaded yet")
	}
	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
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

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// first week lessons
		if 22 <= i && i <= 79 && 2 <= i%10 && i%10 <= 9 {
			dayIdx := i/10 - 2
			lessonIdx := i%10 - 2
			teacherSchedule.Weeks[0].Days[dayIdx].Lessons[lessonIdx] = *getTeacherLessonFromDoc(teacher, s)
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
			teacherSchedule.Weeks[1].Days[dayIdx].Lessons[lessonIdx] = *getTeacherLessonFromDoc(teacher, s)
		}
	})
	return teacherSchedule, nil
}

// convertDailyTeacherScheduleToText converts the information that types.Day contains into text.
func convertDailyTeacherScheduleToText(teacherName string, dailySchedule types.Day, daysAfterCurr int) string {
	result := ""

	dateStr := getDateStr(daysAfterCurr)
	weekNum, weekDayNum := getWeekAndWeekDayNumbers(daysAfterCurr)
	weekDay := convertWeekDayIdxToWeekDay(weekDayNum)

	switch daysAfterCurr {
	case 0:
		result += fmt.Sprintf("%s проводит следующие пары сегодня (%s, %s, %d-ая учебная неделя):\n\n", teacherName,
			weekDay, dateStr, weekNum+1)
	case 1:
		result += fmt.Sprintf("%s проводит следующие пары завтра (%s, %s, %d-ая учебная неделя):\n\n", teacherName,
			weekDay, dateStr, weekNum+1)
	default:
		result += fmt.Sprintf("%s проводит следующие пары %s (%s, %d-ая учебная неделя):\n\n", teacherName,
			dateStr, weekDay, weekNum+1)
	}

	noLessons := true

	for lessonIndex := 0; lessonIndex < len(dailySchedule.Lessons); lessonIndex++ {
		subLessons := dailySchedule.Lessons[lessonIndex].SubLessons

		if len(subLessons) > 0 {
			noLessons = false
			groups := ""
			for indexSubLesson, subLesson := range subLessons {
				if indexSubLesson != len(subLessons)-1 {
					groups += fmt.Sprintf("%s, ", subLesson.Group)
				} else {
					groups += fmt.Sprintf("%s", subLesson.Group)
				}
			}

			lessonNumber := lessonIndex + 1
			lessonTime := lessonsTime[lessonIndex]
			lessonType := subLessons[0].Type
			lessonName := subLessons[0].Name
			lessonTypeWithName := fmt.Sprintf("%s %s", lessonType, lessonName)
			lessonRoom := strings.Replace(subLessons[0].Room, " ", "", -1)
			lessonRoom = strings.Replace(lessonRoom, ".", "", -1)

			if len(strings.Split(groups, ",")) > 1 {
				result += fmt.Sprintf("%d-ая пара (%s): %s, аудитория %s. Группы: %s\n\n",
					lessonNumber, lessonTime, lessonTypeWithName, lessonRoom, groups)
			} else {
				result += fmt.Sprintf("%d-ая пара (%s): %s, %s, аудитория %s\n\n",
					lessonNumber, lessonTime, lessonTypeWithName, groups, lessonRoom)
			}
		}
	}

	if noLessons {
		switch daysAfterCurr {
		case 0:
			result += "Сегодня пар нет"
		case 1:
			result += "Завтра пар нет"
		default:
			result += fmt.Sprintf("%s пар нет", dateStr)
		}
	}

	return result
}

// getTeacherLessonFromDoc returns *types.Lesson received from the HTML document.
func getTeacherLessonFromDoc(teacher string, s *goquery.Selection) *types.Lesson {
	lesson := new(types.Lesson)
	tableCellHTML, _ := s.Find("font").Html()
	// if the table cell contains the lesson info
	if !strings.HasPrefix(tableCellHTML, "_") && tableCellHTML != "" {
		// <br/> separates the name of the lesson, the groups and the audience number
		splitLessonInfoHTML := strings.Split(tableCellHTML, "<br/>")
		lessonGroups := strings.Split(splitLessonInfoHTML[0], ",")
		lesson.SubLessons = make([]types.SubLesson, 0, len(lessonGroups))
		lessonTypeAndName := strings.Split(splitLessonInfoHTML[1], ".")
		lessonType := determineLessonType(lessonTypeAndName[0])
		lessonName := strings.TrimSpace(lessonTypeAndName[len(lessonTypeAndName)-1])
		for _, groupName := range lessonGroups {
			groupLesson := types.SubLesson{
				Type:    lessonType,
				Group:   groupName,
				Name:    lessonName,
				Teacher: teacher,
				Room:    splitLessonInfoHTML[2],
			}
			lesson.SubLessons = append(lesson.SubLessons, groupLesson)
		}
	}
	return lesson
}

// getTeacherURL returns the url to the teacher's schedule on UlSTU site.
func getTeacherURL(teacherName string) (string, error) {
	var teacherURL string

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
	return teacherURL, nil
}
