package schedule

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/text/encoding/charmap"

	"github.com/ulstu-schedule/parser/types"
)

const (
	imgWidth  = 1722
	imgHeight = 1104

	defaultScheduleFontSize = 19

	cellWidth  = 200
	cellHeight = 150

	maxScheduleWeekCount = 2
	lengthScheduleTable  = 91
)

//go:embed assets/week_schedule_teacher_template.png
var weekScheduleTemp []byte

//go:embed assets/Arial.ttf
var font []byte

// weekDays represents string values of the days of week.
var weekDays = [7]string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}

// ParseCurrWeekSchedule returns *types.Week received from *types.Schedule based on the current school week.
func ParseCurrWeekSchedule(schedule *types.Schedule, name string) (*types.Week, error) {
	currWeekDate, _ := getWeekDateAndWeekDay(0)
	return parseWeekSchedule(schedule, name, currWeekDate)
}

// ParseNextWeekSchedule returns *types.Week received from *types.Schedule based on the next school week.
func ParseNextWeekSchedule(schedule *types.Schedule, groupName string) (*types.Week, error) {
	nextWeekDate, _ := getWeekDateAndWeekDay(7)
	return parseWeekSchedule(schedule, groupName, nextWeekDate)
}

// ParseDaySchedule returns *types.Day received from types.Schedule regarding how many days have passed
// relative to the current time.
func ParseDaySchedule(schedule *types.Schedule, name string, daysAfterCurr int) (*types.Day, error) {
	weekDate, weekDayNum := getWeekDateAndWeekDay(daysAfterCurr)

	weekNum := getScheduleWeekNumDyDate(schedule, weekDate)

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: name, WeekNum: weekNum, WeekDayNum: weekDayNum}
	}

	return &schedule.Weeks[weekNum].Days[weekDayNum], nil
}

// getDocFromURL returns goquery document representation of the page with the schedule.
func getDocFromURL(URL string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Printf("error occured while closing response body: %s", err.Error())
		}
	}(response.Body)

	if response.StatusCode >= 300 {
		return nil, &types.StatusCodeError{
			StatusCode: response.StatusCode,
			StatusText: http.StatusText(response.StatusCode),
		}
	}

	// convert from windows-1251 to utf-8
	decoder := charmap.Windows1251.NewDecoder()
	reader := decoder.Reader(response.Body)

	return goquery.NewDocumentFromReader(reader)
}

// determineLessonType returns types.LessonType representation of a string.
func determineLessonType(lessonType string) types.LessonType {
	lessonType = strings.ToLower(lessonType)
	switch lessonType {
	case "лек":
		return types.Lecture
	case "пр":
		return types.Practice
	case "лаб":
		return types.Laboratory
	default:
		return types.Unknown
	}
}

// getWeekDateAndWeekDay get week date and week day
func getWeekDateAndWeekDay(additionalDays int) (time.Time, int) {
	// getting the current time and adding additionalDays days to it
	currTimeWithDelta := time.Now().AddDate(0, 0, additionalDays)
	return getWeekDateAndWeekDayByTime(currTimeWithDelta)
}

// getDateStr increases the current time by daysDelta days and returns the string representation of the new date.
func getDateStr(additionalDays int) string {
	timeWithDelta := time.Now().AddDate(0, 0, additionalDays)
	return timeWithDelta.Format("02.01.2006")
}

// getWeekDateAndWeekDayByTime
func getWeekDateAndWeekDayByTime(weekDate time.Time) (time.Time, int) {
	weekDayNum := int(weekDate.Weekday()) - 1
	if weekDayNum == -1 {
		weekDayNum = 6
	}
	return weekDate, weekDayNum
}

// getDateTime returns time.Time object from the string representation of the date.
func getDateTime(date string) (time.Time, error) {
	day, month, year, err := getDayMonthYearByDate(date)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), nil
}

// getDayMonthYearByDate returns the day, month, and year extracted from the string representation of the date (Date
// format: dd.mm). The year is considered equal to the current one.
func getDayMonthYearByDate(date string) (day int, month int, year int, err error) {
	year = time.Now().Year()
	dateWithYear := fmt.Sprintf("%s.%d", date, year)
	if isDateExist(dateWithYear) {
		dateArray := strings.Split(date, ".")
		day, _ = strconv.Atoi(dateArray[0])
		month, _ = strconv.Atoi(dateArray[1])
	} else {
		err = &types.IncorrectDateError{Date: date}
	}
	return
}

// isDateExist checks if the date matches the format "dd.mm" and exists.
func isDateExist(date string) bool {
	if _, err := time.Parse("02.01.2006", date); err != nil {
		return false
	}
	return true
}

// IsWeekScheduleEmpty returns true if the week schedule is empty, otherwise - false.
func IsWeekScheduleEmpty(week types.Week) bool {
	for _, d := range week.Days {
		for _, l := range d.Lessons {
			if l.SubLessons != nil {
				return false
			}
		}
	}
	return true
}

// IsFullScheduleEmpty returns true if the full schedule is empty, otherwise - false.
func IsFullScheduleEmpty(s *types.Schedule) bool {
	wg := &sync.WaitGroup{}
	inResultsEmptyCheck := make(chan bool, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(weekNum int, wg *sync.WaitGroup, out chan bool) {
			defer wg.Done()
			out <- IsWeekScheduleEmpty(s.Weeks[weekNum])
		}(i, wg, inResultsEmptyCheck)
	}

	wg.Wait()

	return <-inResultsEmptyCheck && <-inResultsEmptyCheck
}

// getRandInt returns a non-negative pseudo-random int.
func getRandInt() int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Int()
}

// highlightRow highlights the row in the table in blue.
func highlightRow(row int, dc *gg.Context) {
	dc.DrawRectangle(4, float64(row-cellHeight), imgWidth-4, cellHeight)
	dc.SetRGBA255(25, 89, 209, 30)
	dc.Fill()

	setDefaultSettings(dc)
}

// setDefaultSettings sets the default drawing settings.
func setDefaultSettings(dc *gg.Context) {
	dc.Stroke()
	dc.SetRGB255(0, 0, 0)

	setFont(defaultScheduleFontSize, dc)
}

// setFontSize sets the font's size depending on the number of lesson parts (lines) in the table cell.
func setFontSize(lessonPartsNum int, dc *gg.Context) {
	switch {
	case lessonPartsNum == 6:
		setFont(16.5, dc)
	case lessonPartsNum == 7:
		setFont(16, dc)
	case lessonPartsNum == 8:
		setFont(15, dc)
	case lessonPartsNum == 9:
		setFont(14, dc)
	case lessonPartsNum == 10:
		setFont(13.5, dc)
	default:
		setFont(12.5, dc)
	}
}

// getWeekScheduleTmplImg returns image.Image based on a byte slice of the embedding png templates.
func getWeekScheduleTmplImg(embeddingTmpl []byte) image.Image {
	weekScheduleTmpl, _, _ := image.Decode(bytes.NewReader(embeddingTmpl))
	return weekScheduleTmpl
}

func setFont(fontSize float64, dc *gg.Context) {
	fnt, _ := truetype.Parse(font)
	face := truetype.NewFace(fnt, &truetype.Options{
		Size: fontSize,
	})
	dc.SetFontFace(face)
}

// getFullSchedule returns the full  schedule.
func getFullSchedule(name string, url string, typeSchedule types.ScheduleType) (*types.Schedule, error) {
	// this group is not on the website with a schedule (the name does not exist or the schedule has not been loaded yet)
	if url == "" {
		return nil, &types.UnavailableScheduleError{Name: name, WeekNum: -1, WeekDayNum: -1}
	}

	doc, err := getDocFromURL(url)
	if err != nil {
		return nil, err
	}

	schedule := &types.Schedule{}

	pSelection := doc.Find("p")
	tablesSchedule := doc.Find("table")

	tablesSchedule.Each(func(tableIdx int, tableS *goquery.Selection) {
		if tableIdx == maxScheduleWeekCount {
			return

		}
		pTableSelection := tableS.Find("p")

		weekNumStr := strings.Split(pSelection.Get(tableIdx*lengthScheduleTable).LastChild.LastChild.Data, ": ")[1]
		weekNumDisplay, _ := strconv.Atoi(strings.Split(weekNumStr, "-")[0])
		schedule.Weeks[tableIdx].Number = weekNumDisplay

		pTableSelection.Each(func(pIdx int, pS *goquery.Selection) {
			iMod10 := pIdx % 10
			iDiv10 := pIdx / 10
			// first week lessons. Условие для дней 21 <= i && i <= 79 && 1 <= iMod10 && iMod10 <= 9
			if 20 <= pIdx && pIdx <= 79 && 0 <= iMod10 && iMod10 <= 8 {
				dayIdx := iDiv10 - 2
				lessonIdx := iMod10 - 1

				if pIdx == 20 {
					dateStartWeek, dateEndWeek := parseDateScheduleWeek(findStartDayWeek, pS)
					schedule.Weeks[tableIdx].DateStart = dateStartWeek
					schedule.Weeks[tableIdx].DateEnd = dateEndWeek
					return
				}

				if pIdx != 20 && iMod10 == 0 {
					return
				}

				schedule.Weeks[tableIdx].Days[dayIdx].WeekNumber = weekNumDisplay
				if typeSchedule == types.Group {
					schedule.Weeks[tableIdx].Days[dayIdx].Lessons[lessonIdx] = *parseGroupLesson(name, lessonIdx,
						findTeacherAndRoom, findTeacher, findRoom, findSubGroup, pS)
				}
				if typeSchedule == types.Teacher {
					schedule.Weeks[tableIdx].Days[dayIdx].Lessons[lessonIdx] = *parseTeacherLesson(name, lessonIdx, pS)
				}
			}
		})
	})

	if IsFullScheduleEmpty(schedule) {
		return nil, &types.UnavailableScheduleError{Name: name, WeekNum: -1, WeekDayNum: -1}
	}

	return schedule, nil
}

// GetImgByWeekSchedule return path on img of schedule
func GetImgByWeekSchedule(
	schedule *types.Week,
	name string,
	isCurrWeek bool,
	headingFontSize float64,
	drawLessonForWeekSchedule func(lesson *types.Lesson, x float64, y float64, dc *gg.Context)) (string, error) {
	// loads an template of an empty table that will be filled in pairs
	tableImg := getWeekScheduleTmplImg(weekScheduleTemp)
	dc := gg.NewContextForImage(tableImg)

	setFont(headingFontSize, dc)
	dc.SetRGB255(25, 89, 209)
	dc.DrawString(name, 515, 60)
	dc.DrawString(fmt.Sprintf("%d-ая", schedule.Number), imgWidth-105, 60)

	setDefaultSettings(dc)

	_, currWeekDayNum := getWeekDateAndWeekDay(0)

	for row := 352; row < imgHeight; row += cellHeight {
		rowNum := row/cellHeight - 2

		x, y := float64(130), float64(row)

		if rowNum == currWeekDayNum && currWeekDayNum != -1 && isCurrWeek {
			highlightRow(row, dc)
		}
		scheduleDay := schedule.Days[rowNum]

		for _, lesson := range scheduleDay.Lessons {
			if len(lesson.SubLessons) > 0 {
				drawLessonForWeekSchedule(&lesson, x, y, dc)
			}
			// moves on to the next lesson
			x += cellWidth
		}
	}

	weekSchedulePath := fmt.Sprintf("week_schedule%d.png", getRandInt())
	return weekSchedulePath, dc.SavePNG(weekSchedulePath)
}

// ParseWeekSchedule returns *types.Week received from *types.Schedule based on the selected school week.
func parseWeekSchedule(schedule *types.Schedule, name string, weekDate time.Time) (*types.Week, error) {
	weekNum := getScheduleWeekNumDyDate(schedule, weekDate)

	if IsWeekScheduleEmpty(schedule.Weeks[weekNum]) {
		return nil, &types.UnavailableScheduleError{Name: name, WeekNum: weekNum, WeekDayNum: -1}
	}

	if schedule.Weeks[0].Number < 3 && schedule.Weeks[1].Number < 3 {
		if weekNum < 0 || weekNum > 1 {
			return nil, &types.IncorrectWeekNumberError{WeekNum: weekNum}
		}
		return &schedule.Weeks[weekNum], nil
	}

	if weekNum > 1 {
		return nil, &types.UnavailableScheduleError{Name: name, WeekNum: weekNum, WeekDayNum: -1}
	}

	return &schedule.Weeks[weekNum], nil
}

func getScheduleWeekNumDyDate(schedule *types.Schedule, weekDate time.Time) int {
	var weekNumSchedule int
	for weekNum, week := range schedule.Weeks {
		isInRange := isInTimeRange(weekDate, week.DateStart, week.DateEnd)
		if isInRange {
			weekNumSchedule = weekNum
			break
		}
	}
	return weekNumSchedule
}

func isInTimeRange(weekDate time.Time, start time.Time, end time.Time) bool {
	return weekDate.After(start) && weekDate.Before(end)
}

func parseDateScheduleWeek(reFindStartDayWeek *regexp.Regexp, s *goquery.Selection) (time.Time, time.Time) {
	dayStartWeekHTML, _ := s.Find("b").Html()
	if dayStartWeekHTML != "" {
		dayStartWeek := reFindStartDayWeek.FindString(dayStartWeekHTML)

		dateStartWeek, _ := getDateStartWeek(dayStartWeek)

		dateEndWeek := dateStartWeek.AddDate(0, 0, 6)
		return dateStartWeek, dateEndWeek
	}
	return time.Now(), time.Now()
}

func getDateStartWeek(dayStartWeek string) (time.Time, error) {
	year, month, day := time.Now().Date()

	dayStartWeekNum, _ := strconv.Atoi(dayStartWeek)
	if dayStartWeekNum < 10 && day > 22 {
		month += 1
	}

	dateString := fmt.Sprintf("%d/%d/%d", year, month, dayStartWeekNum)
	return time.Parse("2006/1/2", dateString)
}
