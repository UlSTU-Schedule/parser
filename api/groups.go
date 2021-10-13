package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"regexp"
	"strings"
)

const (
	groupScheduleURLTemplate = "https://old.ulstu.ru/schedule/students/part%d/%s"
	teachersRegexp           = `([А-Яа-яё]+ [А-Я] [А-Я])|([Прpеeпоoдаaватели]{13} [каaфеeдры]{7}|)`
	roomsRegexp              = `(\d.*[-_].+)|(\d)`
)

// GetFullGroupSchedule returns the full group's schedule
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

	reFindTeacherAndRoom := regexp.MustCompile(fmt.Sprintf(`^%s %s$`, teachersRegexp, roomsRegexp))
	reFindTeacher := regexp.MustCompile(teachersRegexp)
	reFindRoom := regexp.MustCompile(roomsRegexp)

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// first week lessons
		if 22 <= i && i <= 79 && 2 <= i%10 && i%10 <= 9 {
			dayIdx := i/10 - 2
			lessonIdx := i%10 - 2
			groupSchedule.Weeks[0].Days[dayIdx].Lessons[lessonIdx] = *getGroupLessonFromTableCell(groupName, reFindTeacherAndRoom, reFindTeacher, reFindRoom, s)
		}
		// second week lessons
		if 113 <= i && i <= 170 && (i%10 == 0 || i%10 >= 3) {
			var lessonIdx, dayIdx int
			if i%10 == 0 {
				lessonIdx = 7
				dayIdx = i/10 - 12
			} else {
				lessonIdx = i%10 - 3
				dayIdx = i/10 - 11
			}
			groupSchedule.Weeks[1].Days[dayIdx].Lessons[lessonIdx] = *getGroupLessonFromTableCell(groupName, reFindTeacherAndRoom, reFindTeacher, reFindRoom, s)
		}
	})
	return groupSchedule, nil
}

// getGroupLessonFromTableCell returns *types.Lesson received from the HTML table cell
func getGroupLessonFromTableCell(groupName string, reFindTeacherAndRoom *regexp.Regexp, reFindTeacher *regexp.Regexp, reFindRoom *regexp.Regexp, s *goquery.Selection) *types.Lesson {
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

// getGroupURL returns the url to the group's schedule on UlSTU site
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
