package schedule

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/ulstu-schedule/parser/types"
	"golang.org/x/text/encoding/charmap"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	imgWidth  = 1722
	imgHeight = 1104

	fontPath = "assets/Arial.ttf"

	defaultScheduleFontSize = 19

	cellWidth  = 200
	cellHeight = 150
)

// getDocFromURL returns goquery document representation of the page with the schedule.
func getDocFromURL(URL string) (*goquery.Document, error) {
	response, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, &types.StatusCodeError{StatusCode: response.StatusCode, StatusText: http.StatusText(response.StatusCode)}
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
	default:
		return types.Laboratory
	}
}

// getWeekAndWeekDayNumbersByWeekDay returns the numbers of the selected day of the week in the current week and the current week number.
func getWeekAndWeekDayNumbersByWeekDay(weekday string) (int, int) {
	currWeekNum, _ := getWeekAndWeekDayNumbers(0)
	weekdayNum := convertWeekdayToWeekDayIdx(weekday)
	return currWeekNum, weekdayNum
}

// getWeekAndWeekDayNumbers increases the current time by daysDelta days and returns the numbers of the school week and day of the week.
func getWeekAndWeekDayNumbers(additionalDays int) (int, int) {
	// getting the current time and adding additionalDays days to it
	currTimeWithDelta := time.Now().AddDate(0, 0, additionalDays)
	return getWeekAndWeekDayNumbersByTime(currTimeWithDelta)
}

// getDateStr increases the current time by daysDelta days and returns the string representation of the new date.
func getDateStr(additionalDays int) string {
	timeWithDelta := time.Now().AddDate(0, 0, additionalDays)
	return timeWithDelta.Format("02.01.2006")
}

// convertWeekDayIdxToWeekDay converts the week day number to its string representation.
func convertWeekDayIdxToWeekDay(weekDayIdx int) string {
	switch weekDayIdx {
	case 0:
		return "Понедельник"
	case 1:
		return "Вторник"
	case 2:
		return "Среда"
	case 3:
		return "Четверг"
	case 4:
		return "Пятница"
	case 5:
		return "Суббота"
	default:
		return "Воскресенье"
	}
}

// convertWeekdayToWeekDayIdx converts the string representation of the day of the week to its index in the array.
func convertWeekdayToWeekDayIdx(weekday string) int {
	switch strings.ToLower(weekday) {
	case "понедельник":
		return 0
	case "вторник":
		return 1
	case "среда":
		return 2
	case "четверг":
		return 3
	case "пятница":
		return 4
	case "суббота":
		return 5
	default:
		return -1
	}
}

// getWeekAndWeekDayNumbersByDate returns the number of the school week (0 or 1) and the number of the day of the
// school week (-1, 0, ..., 6) by the string representation of the date.
func getWeekAndWeekDayNumbersByDate(date string) (int, int, error) {
	dateTime, err := getDateTime(date)
	if err != nil {
		return 0, 0, err
	}
	weekNum, weekDayNum := getWeekAndWeekDayNumbersByTime(dateTime)
	return weekNum, weekDayNum, nil
}

// getWeekAndWeekDayNumbersByTime returns the number of the school week (0 or 1) and the number of the day of the
// school week (-1, 0, ..., 6) by time.
func getWeekAndWeekDayNumbersByTime(time time.Time) (int, int) {
	weekDayNum := int(time.Weekday()) - 1
	_, currWeekNumWithDelta := time.ISOWeek()
	weekNum := (currWeekNumWithDelta + 1) % 2
	return weekNum, weekDayNum
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
func getDayMonthYearByDate(date string) (int, int, int, error) {
	year := time.Now().Year()
	dateWithYear := fmt.Sprintf("%s.%d", date, year)
	if isDateExist(dateWithYear) {
		dateArray := strings.Split(date, ".")
		day, _ := strconv.Atoi(dateArray[0])
		month, _ := strconv.Atoi(dateArray[1])
		return day, month, year, nil
	} else {
		return 0, 0, 0, &types.IncorrectDateError{Date: date}
	}
}

// isDateExist checks if the date matches the format "dd.mm" and exists.
func isDateExist(date string) bool {
	_, err := time.Parse("02.01.2006", date)
	if err != nil {
		return false
	}
	return true
}

// isWeeklyScheduleEmpty returns true if the weekly schedule is empty, otherwise - false.
func isWeeklyScheduleEmpty(week types.Week) bool {
	for _, d := range week.Days {
		for _, l := range d.Lessons {
			if l.SubLessons != nil {
				return false
			}
		}
	}
	return true
}

// getRandInt returns a non-negative pseudo-random int.
func getRandInt() int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Int()
}

// setDefaultSettings sets the default drawing settings.
func setDefaultSettings(dc *gg.Context) {
	dc.Stroke()
	dc.SetRGB255(0, 0, 0)
	_ = dc.LoadFontFace(fontPath, defaultScheduleFontSize)
}

// highlightRow highlights the row in the table in blue.
func highlightRow(row int, dc *gg.Context) {
	dc.DrawRectangle(4, float64(row-cellHeight), imgWidth-4, cellHeight)
	dc.SetRGBA255(25, 89, 209, 30)
	dc.Fill()

	setDefaultSettings(dc)
}

// setFontSize sets the font size depending on the number of lesson parts (lines) in the table cell.
func setFontSize(lessonPartsNum int, dc *gg.Context) {
	switch {
	case lessonPartsNum == 6:
		_ = dc.LoadFontFace(fontPath, 16.5)
	case lessonPartsNum == 7:
		_ = dc.LoadFontFace(fontPath, 16)
	case lessonPartsNum == 8:
		_ = dc.LoadFontFace(fontPath, 15)
	case lessonPartsNum == 9:
		_ = dc.LoadFontFace(fontPath, 14)
	case lessonPartsNum == 10:
		_ = dc.LoadFontFace(fontPath, 13.5)
	default:
		_ = dc.LoadFontFace(fontPath, 12.5)
	}
}