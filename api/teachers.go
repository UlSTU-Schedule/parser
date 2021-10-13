package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"strings"
)

const teacherScheduleURL = "https://old.ulstu.ru/schedule/teachers/%s"

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
		lessonName := strings.TrimSpace(lessonTypeAndName[1])
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
