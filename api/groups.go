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
	teachersRegexp           = `([А-Яа-яё]+ [А-Я] [А-Я])|([Па-яa-w] кафедры|)`
	roomsRegexp              = `(\d.*[-_].+)|(\d)`
)

// GetFullGroupSchedule ...
func GetFullGroupSchedule(groupName string) (*types.GroupSchedule, error) {
	scheduleURL, err := getGroupURL(groupName)
	if err != nil {
		return nil, err
	}

	doc, err := getDocFromURL(scheduleURL)
	if err != nil {
		return nil, err
	}

	groupSchedule := new(types.GroupSchedule)

	// find strings that consists is teachers and rooms
	re := regexp.MustCompile(fmt.Sprintf(`^%s %s$`, teachersRegexp, roomsRegexp))

	// find teacher in string
	reTeacher := regexp.MustCompile(teachersRegexp)

	// find room in string
	reRoom := regexp.MustCompile(roomsRegexp)

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// first week lessons
		if 22 <= i && i <= 79 && 2 <= i%10 && i%10 <= 9 {
			lesson := new(types.GroupLesson)

			lessonInfo := s.Find("font")
			lessonInfoHTML, _ := lessonInfo.Html()

			// if the lesson exists
			if lessonInfoHTML != "" {
				// the capacity of the slice is 4, because one subgroup cannot have more than 4 pairs at one time
				subLessons := make([]types.GroupSubLesson, 0, 4)
				// each lesson corresponds to 1 or more teachers and rooms that follow it
				lessonBeforeTeachersAndRooms := ""
				// type of lessons that are located before teachers and rooms
				var lessonType types.LessonType
				// <br/> separates the name of the lesson with the teacher and the audience number
				splitLessonInfoHTML := strings.Split(lessonInfoHTML, " <br/>")
				// if <br/> doesn't separate anything, so we do not take it into account
				for j := 0; j < len(splitLessonInfoHTML)-1; j++ {
					// if the row contains teacher and room
					if re.MatchString(splitLessonInfoHTML[j]) {

						subLesson := types.GroupSubLesson{
							Type:    lessonType,
							Group:   groupName,
							Name:    lessonBeforeTeachersAndRooms,
							Teacher: reTeacher.FindString(splitLessonInfoHTML[j]),
							Room:    reRoom.FindString(splitLessonInfoHTML[j]),
						}

						subLessons = append(subLessons, subLesson)
					} else {
						lessonBeforeTeachersAndRooms = splitLessonInfoHTML[j]
						// determine type of lesson
						switch {
						case strings.Contains(lessonBeforeTeachersAndRooms, "Лек."):
							lessonType = types.Lecture
						case strings.Contains(lessonBeforeTeachersAndRooms, "пр."):
							lessonType = types.Practice
						default:
							lessonType = types.Laboratory
						}
					}
				}
				lesson.SubLessons = subLessons
			}

			dayNum := i/10 - 2
			lessonNum := i%10 - 2
			groupSchedule.Weeks[0].Days[dayNum].Lessons[lessonNum] = *lesson
		}

		// second week lessons
		if 113 <= i && i <= 170 && (i%10 == 0 || i%10 >= 3) {
			lesson := new(types.GroupLesson)

			lessonInfo := s.Find("font")
			lessonInfoHTML, _ := lessonInfo.Html()

			// if the lesson exists
			if lessonInfoHTML != "" {
				// the capacity of the slice is 4, because one subgroup cannot have more than 4 pairs at one time
				subLessons := make([]types.GroupSubLesson, 0, 4)
				// each lesson corresponds to 1 or more teachers and rooms that follow it
				lessonBeforeTeachersAndRooms := ""
				// type of lessons that are located before teachers and rooms
				var lessonType types.LessonType
				// <br/> separates the name of the lesson with the teacher and the audience number
				splitLessonInfoHTML := strings.Split(lessonInfoHTML, " <br/>")
				// if <br/> doesn't separate anything, so we do not take it into account
				for j := 0; j < len(splitLessonInfoHTML)-1; j++ {
					// if the row contains teacher and room
					if re.MatchString(splitLessonInfoHTML[j]) {

						subLesson := types.GroupSubLesson{
							Type:    lessonType,
							Group:   groupName,
							Name:    lessonBeforeTeachersAndRooms,
							Teacher: reTeacher.FindString(splitLessonInfoHTML[j]),
							Room:    reRoom.FindString(splitLessonInfoHTML[j]),
						}

						subLessons = append(subLessons, subLesson)
					} else {
						lessonBeforeTeachersAndRooms = splitLessonInfoHTML[j]
						// determine type of lesson
						switch {
						case strings.Contains(lessonBeforeTeachersAndRooms, "Лек."):
							lessonType = types.Lecture
						case strings.Contains(lessonBeforeTeachersAndRooms, "пр."):
							lessonType = types.Practice
						default:
							lessonType = types.Laboratory
						}
					}
				}
				lesson.SubLessons = subLessons
			}

			var (
				lessonNum int
				dayNum    int
			)
			switch {
			case 3 <= i%10 && i%10 <= 9:
				lessonNum = i%10 - 3
				dayNum = i/10 - 11
			case i%10 == 0:
				lessonNum = 7
				dayNum = i/10 - 12
			}
			groupSchedule.Weeks[1].Days[dayNum].Lessons[lessonNum] = *lesson
		}
	})

	return groupSchedule, nil
}

// getGroupURL ...
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
