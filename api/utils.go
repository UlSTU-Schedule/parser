package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"golang.org/x/text/encoding/charmap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var lessonsTime = [8]string{"08:30-09:50", "10:00-11:20", "11:30-12:50", "13:30-14:50", "15:00-16:20", "16:30-17:50", "18:00-19:20", "19:30-20:50"}

// getDocFromURL returns goquery document representation of the page with the schedule.
func getDocFromURL(URL string) (*goquery.Document, error) {
	response, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %s %d", response.Status, response.StatusCode)
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

// getLessonTypeStr returns a string representation of types.LessonType.
func getLessonTypeStr(lessonType types.LessonType) string {
	switch lessonType {
	case types.Lecture:
		return "Лек."
	case types.Practice:
		return "Пр."
	default:
		return "Лаб."
	}
}

// getWeekAndWeekDayNumbersByWeekDay returns the numbers of the selected day of the week in the current week and the current week number.
func getWeekAndWeekDayNumbersByWeekDay(weekday string) (int, int) {
	currWeekNum, _ := getWeekAndWeekdayNumbersBy(0)
	weekdayNum := convertWeekdayToIndex(weekday)
	return currWeekNum, weekdayNum
}

// (TODO: getWeekAndWeekdayNumbersBy -> getWeekAndWeekDayNumbers, daysDelta -> additionalDays)
// getWeekAndWeekdayNumbersBy increases the current time by daysDelta days and returns the numbers of the school week and day of the week.
func getWeekAndWeekdayNumbersBy(daysDelta int) (int, int) {
	// getting the current time and adding daysDelta days to it
	currTimeWithDelta := time.Now().AddDate(0, 0, daysDelta)

	// TODO: replace 73-78 to return getWeekAndWeekDayNumbersByTime(currTimeWithDelta)
	weekdayNum := int(currTimeWithDelta.Weekday()) - 1

	_, currWeekNumWithDelta := currTimeWithDelta.ISOWeek()
	weekNum := (currWeekNumWithDelta + 1) % 2

	return weekNum, weekdayNum
}

// (TODO: getDateStrBy -> getDateStr, daysDelta -> additionalDays)
// getDateStrBy increases the current time by daysDelta days and returns the string representation of the new date.
func getDateStrBy(daysDelta int) string {
	timeWithDelta := time.Now().AddDate(0, 0, daysDelta)
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

// TODO: convertWeekdayToIndex -> convertWeekdayToWeekDayIdx
// convertWeekdayToIndex converts the string representation of the day of the week to its index in the array.
func convertWeekdayToIndex(weekday string) int {
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
	if isDateValid(fmt.Sprintf("%s.%d", date, year)) {
		dateArray := strings.Split(date, ".")
		day, _ := strconv.Atoi(dateArray[0])
		month, _ := strconv.Atoi(dateArray[1])
		return day, month, year, nil
	} else {
		return 0, 0, 0, fmt.Errorf("incorrect date: %s", date)
	}
}

// isDateValid checks if the date is correct. The date is correct if it matches the format dd.mm and exists in the
// current year.
func isDateValid(date string) bool {
	dateTime, err := time.Parse("02.01.2006", date)
	if err != nil {
		return false
	}

	// if the entered date exists
	if dateTime.Format("02.01.2006") == date {
		now := time.Now()

		fallSemStartTime := time.Date(now.Year(), time.Month(9), 1, 0, 0, 0, 0, time.Local)
		fallSemEndTime := time.Date(now.Year(), time.Month(12), 30, 0, 0, 0, 0, time.Local)
		// if the entered date and the current date correspond to the fall semester
		if now.After(fallSemStartTime) && now.Before(fallSemEndTime) &&
			dateTime.After(fallSemStartTime) && dateTime.Before(fallSemEndTime) {
			return true
		}

		springSemStartTime := time.Date(now.Year(), time.Month(2), 1, 0, 0, 0, 0, time.Local)
		springSemEndTime := time.Date(now.Year(), time.Month(6), 30, 0, 0, 0, 0, time.Local)
		// if the entered date and the current date correspond to the spring semester
		if now.After(springSemStartTime) && now.Before(springSemEndTime) &&
			dateTime.After(springSemStartTime) && dateTime.Before(springSemEndTime) {
			return true
		}
	}
	return false
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
