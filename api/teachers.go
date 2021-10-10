package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"strings"
)

const teacherScheduleURL = "https://old.ulstu.ru/schedule/teachers/%s"

func GetTeacherFullSchedule(teacherName string)(*types.TeacherSchedule, error){
	scheduleURL, err := getTeacherURL(teacherName)
	if err != nil {
		return nil, err
	}

	doc, err := getDocFromURL(scheduleURL)
	if err != nil {
		return nil, err
	}
	schedule := types.TeacherSchedule{}
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// first week
		if 22 <= i && i <= 79 && 2 <= i%10 && i%10 <= 9 {
			convertLessonSchedule(&schedule, s, 0, i/10-2, i%10-2, teacherName)
		}
		// second week
		if 113 <= i && i <= 170 && (i%10 == 0 || i%10 >= 3) {
			var indexDay, indexLesson int
			if i % 10 == 0 {
				indexLesson = 7
			} else{
				indexLesson = i % 10 - 3
			}
			if i == 170 {
				indexDay = 5
			} else{
				indexDay = (i / 10) % 10 - 1
			}
			convertLessonSchedule(&schedule, s, 1, indexDay, indexLesson, teacherName)
		}
	})
	return &schedule, nil
}

func convertLessonSchedule(schedule *types.TeacherSchedule, s *goquery.Selection, indexWeek int, indexDay int,
	indexLesson int, teacherName string) {
	font := s.Find("font")
	fontHTML, _ := font.Html()
	var lesson types.TeacherLesson
	if fontHTML != "" {
		formattedStr := strings.Replace(fontHTML, "<br/>", "sep", -1)
		for index, elem := range strings.Split(formattedStr, "sep") {
			if elem != "" && elem != "_" {
				if index == 0 {
					lesson.Groups = strings.Split(elem, ",")
					break
				}
				if index == 1 {
					formattedElem := strings.Split(elem, ".")
					lesson.Name = strings.TrimSpace(formattedElem[len(formattedElem) - 1])
					switch formattedElem[0] {
					case "лек":
						lesson.Type = 0
					case "Лаб":
						lesson.Type = 1
					case "пр":
						lesson.Type = 2
					}
					break
				}
				if index == 2 {
					lesson.Room = elem
					break
				}
			}
		}
		if lesson.Name != "" {
			lesson.Teacher = teacherName
		}
		schedule.Weeks[indexWeek].Days[indexDay].Lessons[indexLesson] = lesson
	}
}

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