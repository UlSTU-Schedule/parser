package schedule

import (
	_ "embed"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/ulstu-schedule/parser/types"
)

const (
	teacherPattern     = `([А-Яа-яё]+ [А-Я] [А-Я])|(АДП П.П.)|([Прpеeпоoдаaватели]{13} [каaфеeдры]{7})`
	roomPattern        = `(\d.*[-_].+)|(\d)|(([А-Я]+-)+\d+)`
	subgroupPattern    = `\d п/г`
	practicePattern    = ` Предприятие`
	startDayWeekPatter = `(\d{2})|(\d)`

	headingTableGroupFontSize = 42
)

var groupScheduleURLs = [4]string{
	strings.Replace("https://lk.ulstu.ru/timetable/shared/schedule/"+url.QueryEscape("Часть 1 - МФ, РТФ, ЭФ (очная, очно-заочная формы обучения), ИФМИ, группы исскуственного интелекта (магистр)"), "+", "%20", -1),
	strings.Replace("https://lk.ulstu.ru/timetable/shared/schedule/"+url.QueryEscape("Часть 2 – ФИСТ, ГФ"), "+", "%20", -1),
	strings.Replace("https://lk.ulstu.ru/timetable/shared/schedule/"+url.QueryEscape("Часть 3 – ИАТУ, СФ, ИЭФ (очная, очно-заочная, заочная формы обучения), ЗВФ ИННО (очно-заочная, заочная формы обучения)"), "+", "%20", -1),
	strings.Replace("https://lk.ulstu.ru/timetable/shared/schedule/"+url.QueryEscape("Часть 4 – КЭИ"), "+", "%20", -1),
}

var (
	findTeacherAndRoom = regexp.MustCompile(fmt.Sprintf(`^%s %s$`, teacherPattern, roomPattern))
	findPractice       = regexp.MustCompile(fmt.Sprintf(`^%s$`, practicePattern))
	findTeacher        = regexp.MustCompile(teacherPattern)
	findRoom           = regexp.MustCompile(roomPattern)
	findSubGroup       = regexp.MustCompile(subgroupPattern)
	findStartDayWeek   = regexp.MustCompile(startDayWeekPatter)

	roomReplacer       = strings.NewReplacer(".", "", "_", "-", " - ", "-", " -", "-", "- ", "-")
	afterSpecCharAdder = strings.NewReplacer(",", ", ", ".", ". ", "- ", " - ", " -", " - ", "&#34;", "'")
)

// GetGroups returns all available group names from UlSTU site.
func GetGroups() []string {
	// there cannot be more than 400 groups
	groups := make([]string, 0, 400)

	for _, scheduleURL := range groupScheduleURLs {
		doc, err := getDocFromURL(scheduleURL + "/raspisan.html")
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

	return ParseDaySchedule(schedule, groupName, daysAfterCurr)
}

// GetCurrWeekGroupScheduleImg returns the path to the image with the week schedule based on the current school week.
func GetCurrWeekGroupScheduleImg(groupName string) (string, error) {
	currWeekDate, _ := getWeekDateAndWeekDay(0)
	return GetWeekGroupScheduleImg(groupName, currWeekDate, true)
}

// GetNextWeekGroupScheduleImg returns the path to the image with the week schedule based on the next school week.
func GetNextWeekGroupScheduleImg(groupName string) (string, error) {
	nextWeekDate, _ := getWeekDateAndWeekDay(7)
	return GetWeekGroupScheduleImg(groupName, nextWeekDate, false)
}

// GetNextWeekGroupSchedule returns *types.Week received from the UlSTU site based on the next school week.
func GetNextWeekGroupSchedule(groupName string) (*types.Week, error) {
	currWeekDate, _ := getWeekDateAndWeekDay(7)
	return GetWeekGroupSchedule(groupName, currWeekDate)
}

// GetCurrWeekGroupSchedule returns *types.Week received from the UlSTU site based on the current school week.
func GetCurrWeekGroupSchedule(groupName string) (*types.Week, error) {
	currWeekDate, _ := getWeekDateAndWeekDay(7)
	return GetWeekGroupSchedule(groupName, currWeekDate)
}

// GetWeekGroupSchedule returns *types.Week received from the UlSTU site based on the selected school week.
func GetWeekGroupSchedule(groupName string, weekDate time.Time) (*types.Week, error) {
	schedule, err := GetFullGroupSchedule(groupName)
	if err != nil {
		return nil, err
	}

	return parseWeekSchedule(schedule, groupName, weekDate)
}

// GetWeekGroupScheduleImg returns the path to the image with the week schedule based on the selected school week.
func GetWeekGroupScheduleImg(groupName string, weekDate time.Time, isCurrWeek bool) (string, error) {
	schedule, err := GetWeekGroupSchedule(groupName, weekDate)
	if err != nil {
		return "", err
	}
	return getImgByWeekGroupSchedule(schedule, groupName, isCurrWeek)
}

// GetFullGroupSchedule returns the full group's schedule.
func GetFullGroupSchedule(groupName string) (*types.Schedule, error) {
	groupScheduleURL, err := getGroupScheduleURL(groupName)
	if err != nil {
		return nil, err
	}

	return getFullSchedule(groupName, groupScheduleURL, types.Group)
}

// ParseCurrWeekGroupScheduleImg returns the path to the image with the week schedule based on the current school week.
func ParseCurrWeekGroupScheduleImg(schedule *types.Week, groupName string) (string, error) {
	return getImgByWeekGroupSchedule(schedule, groupName, true)
}

// ParseNextWeekGroupScheduleImg returns the path to the image with the week schedule based on the next school week.
func ParseNextWeekGroupScheduleImg(schedule *types.Week, groupName string) (string, error) {
	return getImgByWeekGroupSchedule(schedule, groupName, false)
}

// ConvertDayGroupScheduleToText converts the information that *types.Day contains into text.
func ConvertDayGroupScheduleToText(daySchedule *types.Day, groupName string, daysAfterCurr int) string {
	sb := &strings.Builder{}

	dateStr := getDateStr(daysAfterCurr)
	_, weekDayNum := getWeekDateAndWeekDay(daysAfterCurr)

	switch daysAfterCurr {
	case 0:
		_, _ = fmt.Fprintf(sb, "Расписание %s на сегодня (%s, %s, %d-ая учебная неделя):\n\n", groupName,
			weekDays[weekDayNum], dateStr, daySchedule.WeekNumber)
	case 1:
		_, _ = fmt.Fprintf(sb, "Расписание %s на завтра (%s, %s, %d-ая учебная неделя):\n\n", groupName,
			weekDays[weekDayNum], dateStr, daySchedule.WeekNumber)
	default:
		_, _ = fmt.Fprintf(sb, "Расписание %s на %s (%s, %d-ая учебная неделя):\n\n", groupName,
			dateStr, weekDays[weekDayNum], daySchedule.WeekNumber)
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

// GetImgByWeekGroupSchedule returns the path to the image with the week schedule based on the week schedule of
// the group, the name of the group and the selected school week.
func getImgByWeekGroupSchedule(schedule *types.Week, groupName string, isCurrWeek bool) (string, error) {
	return GetImgByWeekSchedule(schedule, groupName, isCurrWeek, headingTableGroupFontSize, drawGroupLessonForWeekSchedule)
}

// putLessonInTableCell draws information about the lesson in the corresponding cell of the week schedule table.
func drawGroupLessonForWeekSchedule(lesson *types.Lesson, cellX, cellY float64, dc *gg.Context) {
	subLessons := lesson.SubLessons
	// the number of lines into which the information about the lesson is divided
	lessonPartsNum := 0
	// information about the lesson to be placed in a table cell
	lessonBuilder := &strings.Builder{}
	// a slice for the schedule of subgroups and different classes for one group
	subLessonsStr := make([]string, len(subLessons))

	for subLessonIdx := range subLessons {
		if subLessons[subLessonIdx].Practice == "" && (strings.Contains(subLessons[subLessonIdx].Name, subLessons[subLessonIdx].Teacher) && subLessons[subLessonIdx].Teacher != "" ||
			strings.Contains(subLessons[subLessonIdx].Name, subLessons[subLessonIdx].Room)) {
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
					subLessonsStr[subLessonIdx] = subLessons[subLessonIdx].StringGroup()
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
				subLessonsStr[subLessonIdx] = subLessons[subLessonIdx].StringGroup()
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

// parseGroupLesson returns *types.Lesson received from the HTML table cell.
func parseGroupLesson(groupName string, lessonIdx int, reFindTeacherAndRoom, reFindTeacher, reFindRoom *regexp.Regexp,
	reFindSubGroup *regexp.Regexp, s *goquery.Selection) *types.Lesson {
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
			subLessonName      string
			subGroupLessonMain string
		)
		// <br/> separates the name of the lesson with the teacher and the audience number
		splitLessonInfoHTML := strings.Split(tableCellHTML, " <br/>")
		// if <br/> doesn't separate anything, so we do not take it into account
		for j := 0; j < len(splitLessonInfoHTML)-1; j++ {
			// if the row contains teacher and room
			if j != 0 && reFindTeacher.MatchString(splitLessonInfoHTML[j]) && reFindRoom.MatchString(splitLessonInfoHTML[j]) {
				subGroupLesson := reFindSubGroup.FindString(splitLessonInfoHTML[j])

				lessonWithoutSubGroup := strings.ReplaceAll(splitLessonInfoHTML[j], subGroupLesson, "")

				room := reFindRoom.FindString(lessonWithoutSubGroup)

				if subGroupLesson == "" {
					subGroupLesson = subGroupLessonMain
					subGroupLessonMain = ""
				}

				lesson.SubLessons = append(lesson.SubLessons, types.SubLesson{
					Duration: types.Duration(lessonIdx),
					Type:     subLessonType,
					Group:    groupName,
					Name:     subLessonName,
					Teacher:  reFindTeacher.FindString(lessonWithoutSubGroup),
					Room:     roomReplacer.Replace(room), // remove extra characters from the room
					SubGroup: subGroupLesson,
				})
			} else if findPractice.MatchString(splitLessonInfoHTML[j]) {
				lesson.SubLessons = append(lesson.SubLessons, types.SubLesson{
					Type:     subLessonType,
					Name:     subLessonName,
					Practice: splitLessonInfoHTML[j],
				})
			} else {
				if j == 0 {
					subLessonName, subGroupLessonMain, subLessonType = getLessonInfo(splitLessonInfoHTML[j], reFindSubGroup)
				} else {
					subLessonName, subGroupLessonMain = getLessonInfoWithoutType(splitLessonInfoHTML[j], reFindSubGroup)
				}
			}
		}
	}
	return lesson
}

func getLessonInfo(lessonInfo string, reFindSubGroup *regexp.Regexp) (string, string, types.LessonType) {
	subLessonTypeAndName := strings.Split(lessonInfo, ".")
	subLessonType := determineLessonType(subLessonTypeAndName[0])
	subGroupLesson := reFindSubGroup.FindString(lessonInfo)

	// add spaces next to special characters (so that there are more hyphenation options when drawing)
	subLessonNameReplaced := afterSpecCharAdder.Replace(strings.Join(subLessonTypeAndName[1:], "."))

	subLessonName := getSubLessonName(subLessonNameReplaced, subGroupLesson)

	return subLessonName, subGroupLesson, subLessonType
}

func getLessonInfoWithoutType(lessonInfo string, reFindSubGroup *regexp.Regexp) (string, string) {
	subGroupLesson := reFindSubGroup.FindString(lessonInfo)

	// add spaces next to special characters (so that there are more hyphenation options when drawing)
	subLessonNameReplaced := afterSpecCharAdder.Replace(lessonInfo)

	subLessonName := getSubLessonName(subLessonNameReplaced, subGroupLesson)

	return subLessonName, subGroupLesson
}

func getSubLessonName(name string, subGroupLesson string) string {
	if subGroupLesson != "" {
		return strings.ReplaceAll(name, "- "+subGroupLesson, "")
	}

	return name
}

// getGroupScheduleURL returns the url to the group's schedule on UlSTU site.
func getGroupScheduleURL(groupName string) (string, error) {
	groupURL := ""

	for _, scheduleURL := range groupScheduleURLs {
		doc, err := getDocFromURL(scheduleURL + "/raspisan.html")
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
							groupURL = scheduleURL + "/" + href
							return false
						}
					}
				} else if foundGroupName == groupName {
					href, _ := s.Find("a").Attr("href")
					groupURL = scheduleURL + "/" + href
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
