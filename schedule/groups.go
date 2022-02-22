package schedule

import (
	_ "embed"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/ulstu-schedule/parser/types"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	teacherPattern = `([А-Яа-яё]+ [А-Я] [А-Я])|(АДП П.П.)|([Прpеeпоoдаaватели]{13} [каaфеeдры]{7}|)`
	roomPattern    = `(\d.*[-_].+)|(\d)`

	headingTableGroupFontSize = 42
)

var groupScheduleURLs = [4]string{
	strings.Replace("https://lk.ulstu.ru/timetable/shared/schedule/"+url.QueryEscape("Часть 1 – РТФ, ЭФ, ИЭФ, ИФМИ"), "+", "%20", -1),
	strings.Replace("https://lk.ulstu.ru/timetable/shared/schedule/"+url.QueryEscape("Часть 2 – ФИСТ, СФ, ГФ"), "+", "%20", -1),
	strings.Replace("https://lk.ulstu.ru/timetable/shared/schedule/"+url.QueryEscape("Часть 3 – МФ, КЭИ"), "+", "%20", -1),
	strings.Replace("https://lk.ulstu.ru/timetable/shared/schedule/"+url.QueryEscape("Часть 4 – ИАТУ, очно-заочная, заочная формы обучения"), "+", "%20", -1),
}

var (
	findTeacherAndRoom = regexp.MustCompile(fmt.Sprintf(`^%s %s$`, teacherPattern, roomPattern))
	findTeacher        = regexp.MustCompile(teacherPattern)
	findRoom           = regexp.MustCompile(roomPattern)

	roomReplacer       = strings.NewReplacer(".", "", "_", "-", " - ", "-", " -", "-", "- ", "-")
	afterSpecCharAdder = strings.NewReplacer(",", ", ", ".", ". ", "- ", " - ", " -", " - ", "&#34;", "'")
)

//go:embed assets/week_schedule_group_template.png
var weekScheduleGroupTemp []byte

// GetTextDayGroupScheduleByDate returns a text representation of the day schedule based on the the string
// representation of the date.
func GetTextDayGroupScheduleByDate(groupName, inputDate string) (string, error) {
	schedule, err := GetDayGroupScheduleByDate(groupName, inputDate)
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

	return ConvertDayGroupScheduleToText(schedule, groupName, diffBetweenInputAndCurrDates), nil
}

// GetDayGroupScheduleByDate returns *types.Day received from the UlSTU site based on the string representation
// of the date.
func GetDayGroupScheduleByDate(groupName, date string) (*types.Day, error) {
	schedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	return ParseDayGroupScheduleByDate(schedule, groupName, date)
}

// ParseDayGroupScheduleByDate returns *types.Day received from types.Schedule based on the string representation
// of the date.
func ParseDayGroupScheduleByDate(schedule *types.Schedule, groupName string, date string) (*types.Day, error) {
	weekNum, weekDayNum, err := GetWeekAndWeekDayNumbersByDate(date)
	if err != nil {
		return nil, err
	}

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: groupName, WeekNum: weekNum, WeekDayNum: weekDayNum}
	}
	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDayGroupScheduleByWeekDay returns a text representation of the day schedule based on the selected day of
// the current week.
func GetTextDayGroupScheduleByWeekDay(groupName, weekDay string) (string, error) {
	schedule, err := GetDayGroupScheduleByWeekDay(groupName, weekDay)
	if err != nil {
		return "", err
	}

	weekDayNumNow := int(time.Now().Weekday()) - 1
	if weekDayNumNow == -1 {
		weekDayNumNow = 6
	}
	weekDayNum := convertWeekDayToWeekDayIdx(weekDay)

	return ConvertDayGroupScheduleToText(schedule, groupName, weekDayNum-weekDayNumNow), nil
}

// GetDayGroupScheduleByWeekDay returns *types.Day received from the UlSTU site based on the selected day of the
// current week.
func GetDayGroupScheduleByWeekDay(groupName, weekDay string) (*types.Day, error) {
	schedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	return ParseDayGroupScheduleByWeekDay(schedule, groupName, weekDay)
}

// ParseDayGroupScheduleByWeekDay returns *types.Day received from types.Schedule based on the selected day of the
// current week.
func ParseDayGroupScheduleByWeekDay(schedule *types.Schedule, groupName string, weekDay string) (*types.Day, error) {
	weekNum, weekDayNum := GetWeekAndWeekDayNumbersByWeekDay(weekDay)

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: groupName, WeekNum: weekNum, WeekDayNum: weekDayNum}
	}

	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// GetTextDayGroupSchedule returns a text representation of the day schedule.
func GetTextDayGroupSchedule(groupName string, daysAfterCurr int) (string, error) {
	schedule, err := GetDayGroupSchedule(groupName, daysAfterCurr)
	if err != nil {
		return "", err
	}

	return ConvertDayGroupScheduleToText(schedule, groupName, daysAfterCurr), nil
}

// GetDayGroupSchedule returns *types.Day received from the UlSTU site regarding how many days have passed
// relative to the current time.
func GetDayGroupSchedule(groupName string, daysAfterCurr int) (*types.Day, error) {
	schedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	return ParseDayGroupSchedule(schedule, groupName, daysAfterCurr)
}

// ParseDayGroupSchedule returns *types.Day received from types.Schedule regarding how many days have passed
// relative to the current time.
func ParseDayGroupSchedule(schedule *types.Schedule, groupName string, daysAfterCurr int) (*types.Day, error) {
	weekNum, weekDayNum := GetWeekAndWeekDayNumbers(daysAfterCurr)

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: groupName, WeekNum: weekNum, WeekDayNum: weekDayNum}
	}

	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// ConvertDayGroupScheduleToText converts the information that *types.Day contains into text.
func ConvertDayGroupScheduleToText(daySchedule *types.Day, groupName string, daysAfterCurr int) string {
	sb := &strings.Builder{}

	dateStr := getDateStr(daysAfterCurr)
	weekNum, weekDayNum := GetWeekAndWeekDayNumbers(daysAfterCurr)

	switch daysAfterCurr {
	case 0:
		_, _ = fmt.Fprintf(sb, "Расписание %s на сегодня (%s, %s, %d-ая учебная неделя):\n\n", groupName,
			weekDays[weekDayNum], dateStr, weekNum+1)
	case 1:
		_, _ = fmt.Fprintf(sb, "Расписание %s на завтра (%s, %s, %d-ая учебная неделя):\n\n", groupName,
			weekDays[weekDayNum], dateStr, weekNum+1)
	default:
		_, _ = fmt.Fprintf(sb, "Расписание %s на %s (%s, %d-ая учебная неделя):\n\n", groupName,
			dateStr, weekDays[weekDayNum], weekNum+1)
	}

	noLessons := true
	for lessonNum := 0; lessonNum < len(daySchedule.Lessons); lessonNum++ {
		if len(daySchedule.Lessons[lessonNum].SubLessons) > 0 {
			noLessons = false
			sb.WriteString(daySchedule.Lessons[lessonNum].StringGroupLesson())
		}
	}

	if noLessons {
		switch daysAfterCurr {
		case 0:
			sb.WriteString("Сегодня пар нет")
		case 1:
			sb.WriteString("Завтра пар нет")
		default:
			_, _ = fmt.Fprintf(sb, "%s пар нет", dateStr)
		}
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// GetCurrWeekGroupScheduleImg returns the path to the image with the week schedule based on the current school week.
func GetCurrWeekGroupScheduleImg(groupName string) (string, error) {
	currWeekNum, _ := GetWeekAndWeekDayNumbers(0)
	return GetWeekGroupScheduleImg(groupName, currWeekNum)
}

// ParseCurrWeekGroupScheduleImg returns the path to the image with the week schedule based on the current school week.
func ParseCurrWeekGroupScheduleImg(schedule *types.Week, groupName string) (string, error) {
	currWeekNum, _ := GetWeekAndWeekDayNumbers(0)
	return GetImgByWeekGroupSchedule(schedule, groupName, currWeekNum)
}

// GetNextWeekGroupScheduleImg returns the path to the image with the week schedule based on the next school week.
func GetNextWeekGroupScheduleImg(groupName string) (string, error) {
	nextWeekNum, _ := GetWeekAndWeekDayNumbers(7)
	return GetWeekGroupScheduleImg(groupName, nextWeekNum)
}

// ParseNextWeekGroupScheduleImg returns the path to the image with the week schedule based on the next school week.
func ParseNextWeekGroupScheduleImg(schedule *types.Week, groupName string) (string, error) {
	nextWeekNum, _ := GetWeekAndWeekDayNumbers(7)
	return GetImgByWeekGroupSchedule(schedule, groupName, nextWeekNum)
}

// GetWeekGroupScheduleImg returns the path to the image with the week schedule based on the selected school week.
func GetWeekGroupScheduleImg(groupName string, weekNum int) (string, error) {
	schedule, err := GetWeekGroupSchedule(groupName, weekNum)
	if err != nil {
		return "", err
	}

	return GetImgByWeekGroupSchedule(schedule, groupName, weekNum)
}

// GetImgByWeekGroupSchedule returns the path to the image with the week schedule based on the week schedule of
// the group, the name of the group and the selected school week.
func GetImgByWeekGroupSchedule(schedule *types.Week, groupName string, weekNum int) (string, error) {
	// loads an template of an empty table that will be filled in pairs
	tableImg := getWeekScheduleTmplImg(weekScheduleGroupTemp)
	dc := gg.NewContextForImage(tableImg)

	setFont(headingTableGroupFontSize, dc)

	// writes the group name and the school week number
	dc.SetRGB255(25, 89, 209)
	dc.DrawString(groupName, 575, 60)
	dc.DrawString(fmt.Sprintf("%d-ая", weekNum+1), imgWidth-105, 60)

	setDefaultSettings(dc)

	currWeekNum, currWeekDayNum := GetWeekAndWeekDayNumbers(0)

	// goes through the row of each day of the week in the week schedule table
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

	weekSchedulePath := fmt.Sprintf("week_schedule%d.png", getRandInt())
	return weekSchedulePath, dc.SavePNG(weekSchedulePath)
}

// putLessonInTableCell draws information about the lesson in the corresponding cell of the week schedule table.
func putLessonInTableCell(subLessons []types.SubLesson, cellX, cellY float64, dc *gg.Context) {
	// the number of lines into which the information about the lesson is divided
	lessonPartsNum := 0
	// information about the lesson to be placed in a table cell
	lessonBuilder := &strings.Builder{}
	// a slice for the schedule of subgroups and different classes for one group
	subLessonsStr := make([]string, len(subLessons))

	for subLessonIdx := range subLessons {
		if strings.Contains(subLessons[subLessonIdx].Name, subLessons[subLessonIdx].Teacher) && subLessons[subLessonIdx].Teacher != "" ||
			strings.Contains(subLessons[subLessonIdx].Name, subLessons[subLessonIdx].Room) {
			continue
		}

		// removes duplicate names of the sublessons
		if subLessonIdx > 0 {
			if strings.Contains(subLessonsStr[0], subLessons[subLessonIdx].Name) {
				isTeacherInInfo := false
				for idx := 0; idx < subLessonIdx; idx++ {
					if strings.Contains(subLessonsStr[idx], subLessons[subLessonIdx].Teacher) {
						isTeacherInInfo = true
						break
					}
				}
				if !isTeacherInInfo {
					subLessonsStr[subLessonIdx] = fmt.Sprintf("%s, аудитория %s",
						subLessons[subLessonIdx].Teacher, subLessons[subLessonIdx].Room)
				}
			} else {
				subLessonsStr[subLessonIdx] = subLessons[subLessonIdx].StringGroupSubLesson()
			}
		} else {
			subLessonsStr[subLessonIdx] = subLessons[subLessonIdx].StringGroupSubLesson()
		}

		// divides the information about the lesson (consists of sublessons) into parts so that it fits into the cell
		lessonParts := dc.WordWrap(strings.Join(subLessonsStr, " "), cellWidth-20)
		lessonPartsNum = len(lessonParts)

		// measures how wide the information about the lesson
		lessonPartsWidth, _ := dc.MeasureMultilineString(strings.Join(lessonParts, "\n"), 1.7)

		if lessonPartsWidth >= cellWidth {
			fLessonName := formatLessonNameToFitIntoCell(subLessons[subLessonIdx].Name)
			// removes duplicate names of the sublessons
			if subLessonIdx > 0 && strings.Contains(subLessonsStr[0], fLessonName) {
				subLessonsStr[subLessonIdx] = fmt.Sprintf("%s, аудитория %s",
					subLessons[subLessonIdx].Teacher, subLessons[subLessonIdx].Room)
			} else {
				subLessons[subLessonIdx].Name = fLessonName
				subLessonsStr[subLessonIdx] = subLessons[subLessonIdx].StringGroupSubLesson()
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
		if subLessonStr != "" {
			lessonBuilder.WriteString(subLessonStr)
			if subLessonIdx != len(subLessonsStr)-1 {
				lessonBuilder.WriteString("; ")
			}
		}
	}

	dc.DrawStringWrapped(lessonBuilder.String(), cellX, cellY-143, 0, 0, cellWidth-20, 1.3, 1)

	if hasFontChanged {
		setFont(defaultScheduleFontSize, dc)
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

// GetNextWeekGroupSchedule returns *types.Week received from the UlSTU site based on the next school week.
func GetNextWeekGroupSchedule(groupName string) (*types.Week, error) {
	nextWeekNum, _ := GetWeekAndWeekDayNumbers(7)
	return GetWeekGroupSchedule(groupName, nextWeekNum)
}

// ParseNextWeekGroupSchedule returns *types.Week received from *types.Schedule based on the next school week.
func ParseNextWeekGroupSchedule(schedule *types.Schedule, groupName string) (*types.Week, error) {
	nextWeekNum, _ := GetWeekAndWeekDayNumbers(7)
	return ParseWeekGroupSchedule(schedule, groupName, nextWeekNum)
}

// GetCurrWeekGroupSchedule returns *types.Week received from the UlSTU site based on the current school week.
func GetCurrWeekGroupSchedule(groupName string) (*types.Week, error) {
	currWeekNum, _ := GetWeekAndWeekDayNumbers(0)
	return GetWeekGroupSchedule(groupName, currWeekNum)
}

// ParseCurrWeekGroupSchedule returns *types.Week received from *types.Schedule based on the current school week.
func ParseCurrWeekGroupSchedule(schedule *types.Schedule, groupName string) (*types.Week, error) {
	currWeekNum, _ := GetWeekAndWeekDayNumbers(0)
	return ParseWeekGroupSchedule(schedule, groupName, currWeekNum)
}

// GetWeekGroupSchedule returns *types.Week received from the UlSTU site based on the selected school week.
func GetWeekGroupSchedule(groupName string, weekNum int) (*types.Week, error) {
	schedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	return ParseWeekGroupSchedule(schedule, groupName, weekNum)
}

// ParseWeekGroupSchedule returns *types.Week received from *types.Schedule based on the selected school week.
func ParseWeekGroupSchedule(schedule *types.Schedule, groupName string, weekNum int) (*types.Week, error) {
	if weekNum < 0 || weekNum > 1 {
		return nil, &types.IncorrectWeekNumberError{WeekNum: weekNum}
	}

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: groupName, WeekNum: weekNum, WeekDayNum: -1}
	}

	return &schedule.Weeks[weekNum], nil
}

// GetFullGroupSchedule returns the full group's schedule.
func GetFullGroupSchedule(groupName string) (*types.Schedule, error) {
	url, err := getGroupScheduleURL(groupName)
	if err != nil {
		return nil, err
	}

	// this group is not on the website with a schedule (the group does not exist or the schedule has not been loaded yet)
	if url == "" {
		return nil, &types.UnavailableScheduleError{Name: groupName, WeekNum: -1, WeekDayNum: -1}
	}

	doc, err := getDocFromURL(url)
	if err != nil {
		return nil, err
	}

	groupSchedule := &types.Schedule{}

	pSelection := doc.Find("p")

	// we have schedule of two weeks
	if pSelection.Length() == 182 {
		pSelection.Each(func(i int, s *goquery.Selection) {
			iMod10 := i % 10
			iDiv10 := i / 10

			// first week lessons
			if 22 <= i && i <= 79 && 2 <= iMod10 && iMod10 <= 9 {
				dayIdx := iDiv10 - 2
				lessonIdx := iMod10 - 2
				groupSchedule.Weeks[0].Days[dayIdx].Lessons[lessonIdx] = *parseGroupLesson(groupName, lessonIdx,
					findTeacherAndRoom, findTeacher, findRoom, s)
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
				groupSchedule.Weeks[1].Days[dayIdx].Lessons[lessonIdx] = *parseGroupLesson(groupName, lessonIdx,
					findTeacherAndRoom, findTeacher, findRoom, s)
			}
		})
	} else {
		// we have one school week schedule
		weekNumStr := pSelection.Get(0).LastChild.LastChild.Data
		weekNum, _ := strconv.Atoi(string(strings.Split(weekNumStr, ": ")[1][0]))

		pSelection.Each(func(i int, s *goquery.Selection) {
			iMod10 := i % 10
			iDiv10 := i / 10

			if 22 <= i && i <= 79 && 2 <= iMod10 && iMod10 <= 9 {
				dayIdx := iDiv10 - 2
				lessonIdx := iMod10 - 2
				groupSchedule.Weeks[weekNum-1].Days[dayIdx].Lessons[lessonIdx] = *parseGroupLesson(groupName, lessonIdx,
					findTeacherAndRoom, findTeacher, findRoom, s)
			}
		})
	}

	if IsFullScheduleEmpty(groupSchedule) {
		return nil, &types.UnavailableScheduleError{Name: groupName, WeekNum: -1, WeekDayNum: -1}
	}

	return groupSchedule, nil
}

// parseGroupLesson returns *types.Lesson received from the HTML table cell.
func parseGroupLesson(groupName string, lessonIdx int, reFindTeacherAndRoom, reFindTeacher, reFindRoom *regexp.Regexp,
	s *goquery.Selection) *types.Lesson {
	lesson := &types.Lesson{}
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
			if j != 0 && reFindTeacherAndRoom.MatchString(splitLessonInfoHTML[j]) {
				room := reFindRoom.FindString(splitLessonInfoHTML[j])

				lesson.SubLessons = append(lesson.SubLessons, types.SubLesson{
					Duration: types.Duration(lessonIdx),
					Type:     subLessonType,
					Group:    groupName,
					Name:     subLessonName,
					Teacher:  reFindTeacher.FindString(splitLessonInfoHTML[j]),
					Room:     roomReplacer.Replace(room), // remove extra characters from the room
				})
			} else {
				if j == 0 {
					subLessonTypeAndName := strings.Split(splitLessonInfoHTML[j], ".")
					subLessonType = determineLessonType(subLessonTypeAndName[0])
					// add spaces next to special characters (so that there are more hyphenation options when drawing)
					subLessonName = afterSpecCharAdder.Replace(strings.Join(subLessonTypeAndName[1:], "."))
				} else {
					subLessonName = afterSpecCharAdder.Replace(splitLessonInfoHTML[j])
				}
			}
		}
	}
	return lesson
}

// getGroupScheduleURL returns the url to the group's schedule on UlSTU site.
func getGroupScheduleURL(groupName string) (string, error) {
	groupURL := ""

	for _, url := range groupScheduleURLs {
		doc, err := getDocFromURL(url + "/raspisan.html")
		if err != nil {
			continue
		}

		doc.Find("td").EachWithBreak(func(i int, s *goquery.Selection) bool {
			foundGroupName := s.Find("font").Text()
			if foundGroupName != "" {
				if strings.Contains(foundGroupName, ", ") {
					foundGroupNames := strings.Split(foundGroupName, ", ")
					for _, foundGroupName = range foundGroupNames {
						if foundGroupName == groupName {
							href, _ := s.Find("a").Attr("href")
							groupURL = url + "/" + href
							return false
						}
					}
				} else if foundGroupName == groupName {
					href, _ := s.Find("a").Attr("href")
					groupURL = url + "/" + href
					return false
				}
			}
			return true
		})

		if groupURL != "" {
			doc, err = getDocFromURL(groupURL)
			if err != nil {
				return "", err
			}

			groupNameFromDoc := doc.Find("p").Get(0).LastChild.FirstChild.Data
			if !strings.Contains(groupNameFromDoc, groupName) {
				return "", &types.IncorrectLinkError{Name: groupName, NameFromURL: groupNameFromDoc}
			}

			return groupURL, nil
		}
	}
	return groupURL, nil
}

// GetGroups returns all available group names from UlSTU site.
func GetGroups() []string {
	// there cannot be more than 400 groups
	groups := make([]string, 0, 400)

	for _, url := range groupScheduleURLs {
		doc, err := getDocFromURL(url + "/raspisan.html")
		if err != nil {
			continue
		}

		doc.Find("td").Each(func(i int, s *goquery.Selection) {
			foundGroupName := s.Find("font").Text()
			if foundGroupName != "" && !strings.Contains(foundGroupName, "курс") {
				if strings.Contains(foundGroupName, ", ") {
					foundGroupNames := strings.Split(foundGroupName, ", ")
					for _, foundGroupName = range foundGroupNames {
						groups = append(groups, foundGroupName)
					}
				} else {
					groups = append(groups, foundGroupName)
				}
			}
		})
	}
	return groups
}
