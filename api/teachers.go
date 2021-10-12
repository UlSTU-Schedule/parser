package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"strings"
)

const teacherScheduleURL = "https://old.ulstu.ru/schedule/teachers/%s"

// GetTeacherFullSchedule - получение полного расписания для учителя
func GetTeacherFullSchedule(teacherName string) (*types.TeacherSchedule, error) {
	scheduleURL, err := getTeacherURL(teacherName)
	if err != nil {
		return nil, err
	}

	doc, err := getDocFromURL(scheduleURL)
	if err != nil {
		return nil, err
	}

	schedule := new(types.TeacherSchedule)
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// first week
		if 22 <= i && i <= 79 && 2 <= i%10 && i%10 <= 9 {
			convertLessonSchedule(schedule, s, 0, i/10-2, i%10-2, teacherName)
		}
		// second week
		if 113 <= i && i <= 170 && (i%10 == 0 || i%10 >= 3) {
			var indexDay, indexLesson int
			if i%10 == 0 {
				indexLesson = 7
			} else {
				indexLesson = i%10 - 3
			}
			if i == 170 {
				indexDay = 5
			} else {
				indexDay = (i/10)%10 - 1
			}
			convertLessonSchedule(schedule, s, 1, indexDay, indexLesson, teacherName)
		}
	})
	return schedule, nil
}

// convertLessonSchedule - конвертация html в объект расписания
func convertLessonSchedule(schedule *types.TeacherSchedule, s *goquery.Selection, indexWeek int, indexDay int,
	indexLesson int, teacherName string) {
	font := s.Find("font")
	fontHTML, _ := font.Html()
	lesson := new(types.TeacherLesson)
	if !strings.HasPrefix(fontHTML, "_") && fontHTML != "" {
		splitHTML := strings.Split(fontHTML, "<br/>")
		groups := strings.Split(splitHTML[0], ",")
		lesson.Groups = make([]types.TeacherGroupLesson, 0, len(groups))
		for _, group := range groups {
			teacherGroupLesson := new(types.TeacherGroupLesson)
			teacherGroupLesson.Teacher = teacherName
			teacherGroupLesson.Group = group
			teacherGroupLesson.Room = splitHTML[2]

			formattedElem := strings.Split(splitHTML[1], ".")
			teacherGroupLesson.Name = strings.TrimSpace(formattedElem[len(formattedElem)-1])
			switch formattedElem[0] {
			case "лек":
				teacherGroupLesson.Type = 0
			case "Лаб":
				teacherGroupLesson.Type = 1
			case "пр":
				teacherGroupLesson.Type = 2
			}
			lesson.Groups = append(lesson.Groups, *teacherGroupLesson)
		}
		schedule.Weeks[indexWeek].Days[indexDay].Lessons[indexLesson] = *lesson
	}
}

// getTeacherURL - получение ссылки на расписания преподавателя
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
