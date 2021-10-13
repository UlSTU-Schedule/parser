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

var lessonsTime = [8]string{"08:30-09:50", "10:00-11:20", "11:30-12:50", "13:30-14:50", "15:00-16:20", "16:30-17:50", "18:00-19:20", "19:30-20:50"}

// GetTextDailyGroupSchedule returns a text representation of the daily schedule.
func GetTextDailyGroupSchedule(groupName string, daysAfterCurr int) (string, error) {
	dailySchedule, err := GetDailyGroupSchedule(groupName, daysAfterCurr)
	if err != nil {
		return "", err
	}

	result := ""

	dateStr := getDateStrBy(daysAfterCurr)
	weekNum, _ := getWeekAndWeekdayNumbersBy(daysAfterCurr)

	switch daysAfterCurr {
	case 0:
		result = fmt.Sprintf("Расписание %s на сегодня (%s, %d-ая учебная неделя):\n\n", groupName,
			dateStr, weekNum+1)
	case 1:
		result = fmt.Sprintf("Расписание %s на завтра (%s, %d-ая учебная неделя):\n\n", groupName,
			dateStr, weekNum+1)
	default:
		result = fmt.Sprintf("Расписание %s на %s (%d-ая учебная неделя):\n\n", groupName,
			dateStr, weekNum+1)
	}

	noLessons := true

	for lessonNum := 0; lessonNum < len(dailySchedule.Lessons); lessonNum++ {
		if len(dailySchedule.Lessons[lessonNum].SubLessons) > 0 {
			noLessons = false

			result += fmt.Sprintf("%d-ая пара (%s): ", lessonNum+1, lessonsTime[lessonNum])

			if len(dailySchedule.Lessons[lessonNum].SubLessons) == 1 {
				formattedLesson := strings.Replace(dailySchedule.Lessons[lessonNum].SubLessons[0].Name, ".", ". ", -1)
				formattedLesson = strings.Replace(formattedLesson, ",", ", ", -1)

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

					formattedLesson := strings.Replace(subgroupLesson.Name, ".", ". ", -1)
					formattedLesson = strings.Replace(formattedLesson, ",", ", ", -1)

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
	return result, nil
}

// GetDailyGroupSchedule returns *types.Day received from the full schedule regarding how many days have passed relative to the current time.
func GetDailyGroupSchedule(groupName string, daysAfterCurr int) (*types.Day, error) {
	fullGroupSchedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	weekNum, weekdayNum := getWeekAndWeekdayNumbersBy(daysAfterCurr)
	// returns weekdayNum = -1 when the day of the week is Sunday
	if weekdayNum == -1 {
		return &types.Day{}, nil
	}
	return &fullGroupSchedule.Weeks[weekNum].Days[weekdayNum], nil
}

// GetFullGroupSchedule returns the full group's schedule.
func GetFullGroupSchedule(groupName string) (*types.Schedule, error) {
	scheduleURL, err := getGroupURL(groupName)
	if err != nil {
		return nil, err
	}

	doc, err := getDocFromURL(scheduleURL)
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

// getGroupURL returns the url to the group's schedule on UlSTU site.
func getGroupURL(groupName string) (string, error) {
	var groupURL string

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
							groupURL = fmt.Sprintf(groupScheduleURLTemplate, schedulePartNum, href)
							return false
						}
					}
				} else if foundGroupName == groupName {
					href, _ := s.Find("a").Attr("href")
					groupURL = fmt.Sprintf(groupScheduleURLTemplate, schedulePartNum, href)
					return false
				}
			}
			return true
		})

		if groupURL != "" {
			return groupURL, nil
		}
	}
	return "", nil
}

// getWeekAndWeekdayNumbersBy increases the current time by daysDelta days and returns the numbers of the school week and day of the week.
func getWeekAndWeekdayNumbersBy(daysDelta int) (int, int) {
	// getting the current time and adding daysDelta days to it
	currTimeWithDelta := time.Now().AddDate(0, 0, daysDelta)

	weekdayNum := int(currTimeWithDelta.Weekday()) - 1

	_, currWeekNumWithDelta := currTimeWithDelta.ISOWeek()
	weekNum := (currWeekNumWithDelta + 1) % 2

	return weekNum, weekdayNum
}

// getDateStrBy increases the current time by daysDelta days and returns the string representation of the new date.
func getDateStrBy(daysDelta int) string {
	timeWithDelta := time.Now().AddDate(0, 0, daysDelta)
	return timeWithDelta.Format("02.01.2006")
}
