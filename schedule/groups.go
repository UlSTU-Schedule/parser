package schedule

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/ulstu-schedule/parser/types"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	groupScheduleURLTemplate = "https://old.ulstu.ru/schedule/students/part%d/%s"
	findTeacherRegexp        = `([А-Яа-яё]+ [А-Я] [А-Я])|([Прpеeпоoдаaватели]{13} [каaфеeдры]{7}|)`
	findRoomRegexp           = `(\d.*[-_].+)|(\d)`
	tableImgGroupPath        = "assets/weekly_schedule_group_template.png"
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

	if isWeeklyScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, errors.New("the schedule for the selected week is empty or not loaded yet")
	}
	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDailyGroupScheduleByWeekDay returns a text representation of the daily schedule based on the selected day of
// the current week.
func GetTextDailyGroupScheduleByWeekDay(groupName, weekDay string) (string, error) {
	schedule, err := GetDailyGroupScheduleByWeekDay(groupName, weekDay)
	if err != nil {
		return "", err
	}

	weekDayNumNow := int(time.Now().Weekday()) - 1
	weekDayNum := convertWeekdayToWeekDayIdx(weekDay)
	if weekDayNum == -1 && weekDayNumNow != -1 {
		weekDayNum = 6
	}
	return convertDailyGroupScheduleToText(groupName, schedule, weekDayNum-weekDayNumNow), nil
}

// GetDailyGroupScheduleByWeekDay returns *types.Day received from the full schedule based on the selected day of the
// current week.
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

	if isWeeklyScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, errors.New("the schedule for the selected week is empty or not loaded yet")
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

// GetDailyGroupSchedule returns *types.Day received from the full schedule regarding how many days have passed
// relative to the current time.
func GetDailyGroupSchedule(groupName string, daysAfterCurr int) (*types.Day, error) {
	schedule, err := GetFullGroupSchedule(groupName)
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

// convertDailyGroupScheduleToText converts the information that *types.Day contains into text.
func convertDailyGroupScheduleToText(groupName string, dailySchedule *types.Day, daysAfterCurr int) string {
	result := ""

	dateStr := getDateStr(daysAfterCurr)
	weekNum, weekDayNum := getWeekAndWeekDayNumbers(daysAfterCurr)
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

			result += fmt.Sprintf("%d-ая пара (%s): ", lessonNum+1, dailySchedule.Lessons[lessonNum].SubLessons[0].Duration)

			if len(dailySchedule.Lessons[lessonNum].SubLessons) == 1 {
				formattedLesson := strings.Replace(dailySchedule.Lessons[lessonNum].SubLessons[0].Name, ",",
					", ", -1)

				formattedRoom := strings.Replace(dailySchedule.Lessons[lessonNum].SubLessons[0].Room, " ", "",
					-1)
				formattedRoom = strings.Replace(formattedRoom, ".", "", -1)

				result += fmt.Sprintf("%s %s, %s, аудитория %s", dailySchedule.Lessons[lessonNum].SubLessons[0].Type,
					formattedLesson, dailySchedule.Lessons[lessonNum].SubLessons[0].Teacher, formattedRoom)
			} else {
				var subgroupsLessonsOnCurrLesson string
				for _, subgroupLesson := range dailySchedule.Lessons[lessonNum].SubLessons {
					if strings.Contains(subgroupLesson.Name, subgroupLesson.Teacher) ||
						strings.Contains(subgroupLesson.Name, subgroupLesson.Room) {
						continue
					}

					formattedLesson := strings.Replace(subgroupLesson.Name, ",", ", ", -1)

					formattedRoom := strings.Replace(subgroupLesson.Room, " ", "", -1)
					formattedRoom = strings.Replace(formattedRoom, ".", "", -1)

					subgroupLessonInfo := fmt.Sprintf("%s %s, %s, аудитория %s; ", subgroupLesson.Type,
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

// GetCurrWeekGroupScheduleImg returns the path to the image with the weekly schedule based on the current school week.
func GetCurrWeekGroupScheduleImg(groupName string) (string, error) {
	currWeekNum, _ := getWeekAndWeekDayNumbers(0)
	return GetWeeklyGroupScheduleImg(groupName, currWeekNum)
}

// GetNextWeekGroupScheduleImg returns the path to the image with the weekly schedule based on the next school week.
func GetNextWeekGroupScheduleImg(groupName string) (string, error) {
	nextWeekNum, _ := getWeekAndWeekDayNumbers(7)
	return GetWeeklyGroupScheduleImg(groupName, nextWeekNum)
}

// GetCurrWeekGroupSchedule returns *types.Week received from the full schedule based on the current school week.
func GetCurrWeekGroupSchedule(groupName string) (*types.Week, error) {
	currWeekNum, _ := getWeekAndWeekDayNumbers(0)
	return GetWeeklyGroupSchedule(groupName, currWeekNum)
}

// GetNextWeekGroupSchedule returns *types.Week received from the full schedule based on the next school week.
func GetNextWeekGroupSchedule(groupName string) (*types.Week, error) {
	nextWeekNum, _ := getWeekAndWeekDayNumbers(7)
	return GetWeeklyGroupSchedule(groupName, nextWeekNum)
}

// GetWeeklyGroupScheduleImg returns the path to the image with the weekly schedule based on the selected school week.
func GetWeeklyGroupScheduleImg(groupName string, weekNum int) (string, error) {
	schedule, err := GetWeeklyGroupSchedule(groupName, weekNum)
	if err != nil {
		return "", err
	}

	// loads an image of an empty table that will be filled in pairs
	tableImg, _ := gg.LoadPNG(tableImgGroupPath)
	dc := gg.NewContextForImage(tableImg)

	// writes the group name and the school week number
	_ = dc.LoadFontFace(fontPath, headingTableFontSize)
	dc.SetRGB255(25, 89, 209)
	dc.DrawString(groupName, 575, 60)
	dc.DrawString(fmt.Sprintf("%d-ая", weekNum+1), imgWidth-105, 60)

	setDefaultSettings(dc)

	currWeekNum, currWeekDayNum := getWeekAndWeekDayNumbers(0)

	// goes through the row of each day of the week in the weekly schedule table
	for row := 352; row < imgHeight; row += cellHeight {
		// corresponds to the index of the day of the week and starts with 0
		rowNum := row/cellHeight - 2

		// coordinates of the beginning of the current row
		x, y := float64(130), float64(row)

		// if the row corresponds to the current day of the week, which is not Sunday
		if rowNum == currWeekDayNum && currWeekDayNum != -1 && weekNum == currWeekNum {
			highlightRow(row, dc)
		}

		for _, lessons := range schedule.Days[rowNum].Lessons {
			if len(lessons.SubLessons) > 0 {
				putLessonInTableCell(lessons.SubLessons, x, y, dc)
			}
			// moves on to the next lesson
			x += cellWidth
		}
	}

	weeklySchedulePath := fmt.Sprintf("assets/weekly_schedule%d.png", getRandInt())
	return weeklySchedulePath, dc.SavePNG(weeklySchedulePath)
}

// putLessonInTableCell draws information about the lesson in the corresponding cell of the weekly schedule table.
func putLessonInTableCell(subLessons []types.SubLesson, cellX, cellY float64, dc *gg.Context) {
	// the number of lines into which the information about the lesson is divided
	lessonPartsNum := 0
	// information about the lesson to be placed in a table cell
	lessonStr := ""
	// a slice for the schedule of subgroups and different classes for one group
	subLessonsStr := make([]string, len(subLessons))

	for subLessonIdx := range subLessons {
		if strings.Contains(subLessons[subLessonIdx].Name, subLessons[subLessonIdx].Teacher) ||
			strings.Contains(subLessons[subLessonIdx].Name, subLessons[subLessonIdx].Room) {
			continue
		}

		fLessonName := strings.Replace(subLessons[subLessonIdx].Name, ".", ". ", -1)
		fLessonName = strings.Replace(fLessonName, "-", " – ", -1)
		fLessonName = strings.Replace(fLessonName, ",", ", ", -1)

		fRoom := strings.Replace(subLessons[subLessonIdx].Room, " ", "", -1)
		fRoom = strings.Replace(fRoom, ".", "", -1)

		subLessonInfo := fmt.Sprintf("%s %s, %s, аудитория %s", subLessons[subLessonIdx].Type, fLessonName,
			subLessons[subLessonIdx].Teacher, fRoom)

		// removes duplicate names of the sublessons
		if subLessonIdx > 0 {
			if strings.Replace(subLessonsStr[0], "- ", "", -1) == subLessonInfo {
				continue
			}

			if strings.Contains(subLessonsStr[0], fLessonName) {
				isTeacherInInfo := false
				for idx := 0; idx < subLessonIdx; idx++ {
					if strings.Contains(subLessonsStr[idx], subLessons[subLessonIdx].Teacher) {
						isTeacherInInfo = true
						break
					}
				}
				if !isTeacherInInfo {
					subLessonsStr[subLessonIdx] = fmt.Sprintf("%s, аудитория %s",
						subLessons[subLessonIdx].Teacher, fRoom)
				}
			} else {
				subLessonsStr[subLessonIdx] = subLessonInfo
			}
		} else {
			subLessonsStr[subLessonIdx] = subLessonInfo
		}

		// divides the information about the lesson (consists of sublessons) into parts so that it fits into the cell
		lessonParts := dc.WordWrap(strings.Join(subLessonsStr, " "), cellWidth-20)
		lessonPartsNum = len(lessonParts)

		// measures how wide the information about the lesson
		lessonPartsWidth, _ := dc.MeasureMultilineString(strings.Join(lessonParts, "\n"),
			1.7)

		if lessonPartsWidth >= cellWidth {
			fLessonName = formatLessonNameToFitIntoCell(fLessonName)
			// removes duplicate names of the sublessons
			if subLessonIdx > 0 && strings.Contains(subLessonsStr[0], fLessonName) {
				subLessonsStr[subLessonIdx] = fmt.Sprintf("%s, аудитория %s",
					subLessons[subLessonIdx].Teacher, fRoom)
			} else {
				subLessonsStr[subLessonIdx] = fmt.Sprintf("%s %s, %s, аудитория %s", subLessons[subLessonIdx].Type,
					fLessonName, subLessons[subLessonIdx].Teacher, fRoom)
			}
		}
	}

	hasFontChanged := false
	// reduces the font size if there are more parts with the lesson schedule than fit in the cell
	if lessonPartsNum > 5 {
		setFontSize(lessonPartsNum, dc)
		hasFontChanged = true
	}

	for subLessonIdx, subLessonStr := range subLessonsStr {
		if len(subLessonStr) > 0 {
			lessonStr += subLessonStr
			if subLessonIdx != len(subLessonsStr)-1 {
				lessonStr += "; "
			}
		}
	}

	dc.DrawStringWrapped(lessonStr, cellX, cellY-143, 0, 0, cellWidth-20, 1.7, 1)

	if hasFontChanged {
		_ = dc.LoadFontFace(fontPath, defaultScheduleFontSize)
	}
}

// formatLessonNameToFitIntoCell changes the name of the lesson so that it fits the width of the table cell, and
// returns the changed value.
func formatLessonNameToFitIntoCell(lessonName string) string {
	lessonNameParts := strings.Split(lessonName, " ")
	for idx := range lessonNameParts {
		line := lessonNameParts[idx]
		runeCntInLine := utf8.RuneCountInString(line)
		if runeCntInLine > 18 {
			newCurrLine := ""
			for runeCnt := 0; len(line) > 0; runeCnt++ {
				r, size := utf8.DecodeRuneInString(line)
				newCurrLine += fmt.Sprintf("%c", r)
				if runeCnt == runeCntInLine/2 {
					newCurrLine += "- "
				}
				line = line[size:]
			}
			lessonNameParts[idx] = newCurrLine
		}
	}
	return strings.Join(lessonNameParts, " ")
}

// GetWeeklyGroupSchedule returns *types.Week received from the full schedule based on the selected school week.
func GetWeeklyGroupSchedule(groupName string, weekNum int) (*types.Week, error) {
	if weekNum < 0 || weekNum > 1 {
		return nil, errors.New("incorrect value of the week number: it can be either zero or one")
	}

	schedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	if isWeeklyScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, errors.New("the schedule for the selected week is empty or not loaded yet")
	}
	return &schedule.Weeks[weekNum], nil
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
			groupSchedule.Weeks[0].Days[dayIdx].Lessons[lessonIdx] = *getGroupLessonFromTableCell(groupName, lessonIdx,
				reFindTeacherAndRoom, reFindTeacher, reFindRoom, s)
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
			groupSchedule.Weeks[1].Days[dayIdx].Lessons[lessonIdx] = *getGroupLessonFromTableCell(groupName, lessonIdx,
				reFindTeacherAndRoom, reFindTeacher, reFindRoom, s)
		}
	})
	return groupSchedule, nil
}

// getGroupLessonFromTableCell returns *types.Lesson received from the HTML table cell.
func getGroupLessonFromTableCell(groupName string, lessonIdx int, reFindTeacherAndRoom, reFindTeacher, reFindRoom *regexp.Regexp,
	s *goquery.Selection) *types.Lesson {
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
					Duration: types.Duration(lessonIdx),
					Type:     subLessonType,
					Group:    groupName,
					Name:     subLessonName,
					Teacher:  reFindTeacher.FindString(splitLessonInfoHTML[j]),
					Room:     reFindRoom.FindString(splitLessonInfoHTML[j]),
				}
				lesson.SubLessons = append(lesson.SubLessons, subLesson)
			} else {
				if j == 0 {
					subLessonTypeAndName := strings.Split(splitLessonInfoHTML[j], ".")
					subLessonType = determineLessonType(subLessonTypeAndName[0])
					subLessonName = strings.Join(subLessonTypeAndName[1:], ". ")
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
