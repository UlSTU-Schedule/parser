package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"regexp"
	"strings"
	"time"
)

const (
	groupScheduleURLTemplate = "https://old.ulstu.ru/schedule/students/part%d/%s"
	findTeacherRegexp        = `([А-Яа-яё]+ [А-Я] [А-Я])|([Прpеeпоoдаaватели]{13} [каaфеeдры]{7}|)`
	findRoomRegexp           = `(\d.*[-_].+)|(\d)`
)

// GetTextDailyGroupScheduleByDate returns a text representation of the daily schedule based on the the string
// representation of the date.
func GetTextDailyGroupScheduleByDate(groupName, inputDate string) (string, error) {
	schedule, err := GetDailyGroupScheduleByDate(groupName, inputDate)
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
	return convertDailyGroupScheduleToText(groupName, schedule, diffBetweenInputAndCurrDates), nil
}

// GetDailyGroupScheduleByDate returns *types.Day received from the full schedule based on the string representation
// of the date.
func GetDailyGroupScheduleByDate(groupName, date string) (*types.Day, error) {
	schedule, err := GetFullGroupSchedule(groupName)
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
	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDailyGroupScheduleByWeekDay returns a text representation of the daily schedule based on the selected day of the current week.
func GetTextDailyGroupScheduleByWeekDay(groupName, weekDay string) (string, error) {
	schedule, err := GetDailyGroupScheduleByWeekDay(groupName, weekDay)
	if err != nil {
		return "", err
	}
	weekDayNumNow := int(time.Now().Weekday()) - 1
	weekDayNum := convertWeekdayToIndex(weekDay)
	if weekDayNum == -1 {
		weekDayNum = 6
	}
	return convertDailyGroupScheduleToText(groupName, schedule, weekDayNum-weekDayNumNow), nil
}

// GetDailyGroupScheduleByWeekDay returns *types.Day received from the full schedule based on the selected day of the current week.
func GetDailyGroupScheduleByWeekDay(groupName, weekDay string) (*types.Day, error) {
	schedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	weekNum, weekDayNum := getWeekAndWeekDayNumbersByWeekDay(weekDay)
	// returns weekDayNum = -1 when the day of the week is Sunday
	if weekDayNum == -1 {
		return &types.Day{}, nil
	}
	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDailyGroupSchedule returns a text representation of the daily schedule.
func GetTextDailyGroupSchedule(groupName string, daysAfterCurr int) (string, error) {
	schedule, err := GetDailyGroupSchedule(groupName, daysAfterCurr)
	if err != nil {
		return "", err
	}
	return convertDailyGroupScheduleToText(groupName, schedule, daysAfterCurr), nil
}

// GetDailyGroupSchedule returns *types.Day received from the full schedule regarding how many days have passed relative to the current time.
func GetDailyGroupSchedule(groupName string, daysAfterCurr int) (*types.Day, error) {
	schedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	weekNum, weekDayNum := getWeekAndWeekdayNumbersBy(daysAfterCurr)
	// returns weekDayNum = -1 when the day of the week is Sunday
	if weekDayNum == -1 {
		return &types.Day{}, nil
	}
	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// convertDailyGroupScheduleToText converts the information that *types.Day contains into text.
func convertDailyGroupScheduleToText(groupName string, dailySchedule *types.Day, daysAfterCurr int) string {
	result := ""

	dateStr := getDateStrBy(daysAfterCurr)
	weekNum, weekDayNum := getWeekAndWeekdayNumbersBy(daysAfterCurr)
	weekDay := convertWeekDayIdxToWeekDay(weekDayNum)

	switch daysAfterCurr {
	case 0:
		result = fmt.Sprintf("Расписание %s на сегодня (%s, %s, %d-ая учебная неделя):\n\n", groupName,
			weekDay, dateStr, weekNum+1)
	case 1:
		result = fmt.Sprintf("Расписание %s на завтра (%s, %s, %d-ая учебная неделя):\n\n", groupName,
			weekDay, dateStr, weekNum+1)
	default:
		result = fmt.Sprintf("Расписание %s на %s (%s, %d-ая учебная неделя):\n\n", groupName,
			dateStr, weekDay, weekNum+1)
	}

	noLessons := true

	for lessonNum := 0; lessonNum < len(dailySchedule.Lessons); lessonNum++ {
		if len(dailySchedule.Lessons[lessonNum].SubLessons) > 0 {
			noLessons = false

			result += fmt.Sprintf("%d-ая пара (%s): ", lessonNum+1, lessonsTime[lessonNum])

			if len(dailySchedule.Lessons[lessonNum].SubLessons) == 1 {
				formattedLesson := strings.Replace(dailySchedule.Lessons[lessonNum].SubLessons[0].Name, ",", ", ", -1)

				formattedRoom := strings.Replace(dailySchedule.Lessons[lessonNum].SubLessons[0].Room, " ", "", -1)
				formattedRoom = strings.Replace(formattedRoom, ".", "", -1)

				lessonTypeStr := getLessonTypeStr(dailySchedule.Lessons[lessonNum].SubLessons[0].Type)

				result += fmt.Sprintf("%s %s, %s, аудитория %s", lessonTypeStr, formattedLesson,
					dailySchedule.Lessons[lessonNum].SubLessons[0].Teacher, formattedRoom)
			} else {
				var subgroupsLessonsOnCurrLesson string
				for _, subgroupLesson := range dailySchedule.Lessons[lessonNum].SubLessons {
					if strings.Contains(subgroupLesson.Name, subgroupLesson.Teacher) || strings.Contains(subgroupLesson.Name, subgroupLesson.Room) {
						continue
					}

					formattedLesson := strings.Replace(subgroupLesson.Name, ",", ", ", -1)

					formattedRoom := strings.Replace(subgroupLesson.Room, " ", "", -1)
					formattedRoom = strings.Replace(formattedRoom, ".", "", -1)

					subgroupLessonTypeStr := getLessonTypeStr(subgroupLesson.Type)

					subgroupLessonInfo := fmt.Sprintf("%s %s, %s, аудитория %s; ", subgroupLessonTypeStr,
						formattedLesson, subgroupLesson.Teacher, formattedRoom)

					if !strings.Contains(subgroupsLessonsOnCurrLesson, subgroupLessonInfo) {
						subgroupsLessonsOnCurrLesson += subgroupLessonInfo
					}
				}
				result += strings.TrimSuffix(subgroupsLessonsOnCurrLesson, "; ")
			}
			result += "\n\n"
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

// GetFullGroupSchedule returns the full group's schedule.
func GetFullGroupSchedule(groupName string) (*types.Schedule, error) {
	url, err := getGroupScheduleURL(groupName)
	if err != nil {
		return nil, err
	}

	doc, err := getDocFromURL(url)
	if err != nil {
		return nil, err
	}

	groupSchedule := new(types.Schedule)

	reFindTeacherAndRoom := regexp.MustCompile(fmt.Sprintf(`^%s %s$`, findTeacherRegexp, findRoomRegexp))
	reFindTeacher := regexp.MustCompile(findTeacherRegexp)
	reFindRoom := regexp.MustCompile(findRoomRegexp)

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		iMod10 := i % 10
		iDiv10 := i / 10

		// first week lessons
		if 22 <= i && i <= 79 && 2 <= iMod10 && iMod10 <= 9 {
			dayIdx := iDiv10 - 2
			lessonIdx := iMod10 - 2
			groupSchedule.Weeks[0].Days[dayIdx].Lessons[lessonIdx] = *getGroupLessonFromTableCell(groupName, reFindTeacherAndRoom, reFindTeacher, reFindRoom, s)
		}
		// second week lessons
		if 113 <= i && i <= 170 && (iMod10 == 0 || iMod10 >= 3) {
			var lessonIdx, dayIdx int
			if iMod10 == 0 {
				lessonIdx = 7
				dayIdx = iDiv10 - 12
			} else {
				lessonIdx = iMod10 - 3
				dayIdx = iDiv10 - 11
			}
			groupSchedule.Weeks[1].Days[dayIdx].Lessons[lessonIdx] = *getGroupLessonFromTableCell(groupName, reFindTeacherAndRoom, reFindTeacher, reFindRoom, s)
		}
	})
	return groupSchedule, nil
}

// getGroupLessonFromTableCell returns *types.Lesson received from the HTML table cell.
func getGroupLessonFromTableCell(groupName string, reFindTeacherAndRoom, reFindTeacher, reFindRoom *regexp.Regexp, s *goquery.Selection) *types.Lesson {
	lesson := new(types.Lesson)
	tableCellHTML, _ := s.Find("font").Html()
	// if the table cell contains the lesson info
	if tableCellHTML != "" {
		// the capacity of the slice is 5, because one subgroup cannot have more than 5 pairs at one time
		lesson.SubLessons = make([]types.SubLesson, 0, 5)
		var (
			// type of lessons that are located before teachers and rooms
			subLessonType types.LessonType
			// each lesson corresponds to 1 or more teachers and rooms that follow it
			subLessonName string
		)
		// <br/> separates the name of the lesson with the teacher and the audience number
		splitLessonInfoHTML := strings.Split(tableCellHTML, " <br/>")
		// if <br/> doesn't separate anything, so we do not take it into account
		for j := 0; j < len(splitLessonInfoHTML)-1; j++ {
			// if the row contains teacher and room
			if reFindTeacherAndRoom.MatchString(splitLessonInfoHTML[j]) {
				subLesson := types.SubLesson{
					Type:    subLessonType,
					Group:   groupName,
					Name:    subLessonName,
					Teacher: reFindTeacher.FindString(splitLessonInfoHTML[j]),
					Room:    reFindRoom.FindString(splitLessonInfoHTML[j]),
				}
				lesson.SubLessons = append(lesson.SubLessons, subLesson)
			} else {
				if j == 0 {
					subLessonTypeAndName := strings.Split(splitLessonInfoHTML[j], ".")
					subLessonName = subLessonTypeAndName[1]
					subLessonType = determineLessonType(subLessonTypeAndName[0])
				} else {
					subLessonName = splitLessonInfoHTML[j]
				}
			}
		}
	}
	return lesson
}

// getGroupScheduleURL returns the url to the group's schedule on UlSTU site.
func getGroupScheduleURL(groupName string) (string, error) {
	result := ""

	for schedulePartNum := 1; schedulePartNum < 4; schedulePartNum++ {
		doc, err := getDocFromURL(fmt.Sprintf(groupScheduleURLTemplate, schedulePartNum, "raspisan.html"))
		if err != nil {
			return "", err
		}

		doc.Find("td").EachWithBreak(func(i int, s *goquery.Selection) bool {
			foundGroupName := s.Find("font").Text()
			if foundGroupName != "" {
				if strings.Contains(foundGroupName, ", ") {
					foundGroupNames := strings.Split(foundGroupName, ", ")
					for _, foundGroupName = range foundGroupNames {
						if foundGroupName == groupName {
							href, _ := s.Find("a").Attr("href")
							result = fmt.Sprintf(groupScheduleURLTemplate, schedulePartNum, href)
							return false
						}
					}
				} else if foundGroupName == groupName {
					href, _ := s.Find("a").Attr("href")
					result = fmt.Sprintf(groupScheduleURLTemplate, schedulePartNum, href)
					return false
				}
			}
			return true
		})

		if result != "" {
			return result, nil
		}
	}
	return result, nil
}
