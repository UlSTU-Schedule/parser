package schedule

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/ulstu-schedule/parser/types"
	"golang.org/x/text/encoding/charmap"
	"image"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	imgWidth  = 1722
	imgHeight = 1104

	defaultScheduleFontSize = 19

	cellWidth  = 200
	cellHeight = 150
)

//go:embed assets/Arial.ttf
var font []byte

// weekDays represents string values of the days of week.
var weekDays = [7]string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}

// getDocFromURL returns goquery document representation of the page with the schedule.
func getDocFromURL(URL string) (*goquery.Document, error) {
	response, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Printf("error occured while closing response body: %s", err.Error())
		}
	}(response.Body)

	if response.StatusCode > 299 {
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

// GetWeekAndWeekDayNumbersByDate returns the number of the school week (0 or 1) and the number of the day of the
// school week (0, ..., 6) by the string representation of the date.
func GetWeekAndWeekDayNumbersByDate(date string) (weekNum int, weekDayNum int, err error) {
	dateTime, err := getDateTime(date)
	if err != nil {
		return
	}
	weekNum, weekDayNum = getWeekAndWeekDayNumbersByTime(dateTime)
	return
}

// GetWeekAndWeekDayNumbersByWeekDay returns the numbers of the selected day of the week in the current week and the current week number.
func GetWeekAndWeekDayNumbersByWeekDay(weekDay string) (int, int) {
	currWeekNum, _ := GetWeekAndWeekDayNumbers(0)
	weekDayNum := convertWeekDayToWeekDayIdx(weekDay)
	return currWeekNum, weekDayNum
}

// GetWeekAndWeekDayNumbers increases the current time by daysDelta days and returns the numbers of the school week and day of the week.
func GetWeekAndWeekDayNumbers(additionalDays int) (int, int) {
	// getting the current time and adding additionalDays days to it
	currTimeWithDelta := time.Now().AddDate(0, 0, additionalDays)
	return getWeekAndWeekDayNumbersByTime(currTimeWithDelta)
}

// getDateStr increases the current time by daysDelta days and returns the string representation of the new date.
func getDateStr(additionalDays int) string {
	timeWithDelta := time.Now().AddDate(0, 0, additionalDays)
	return timeWithDelta.Format("02.01.2006")
}

// convertWeekDayToWeekDayIdx converts the string representation of the day of the week to its index in the array.
func convertWeekDayToWeekDayIdx(weekDay string) int {
	switch strings.ToLower(weekDay) {
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
		return 6
	}
}

// getWeekAndWeekDayNumbersByTime returns the number of the school week (0 or 1) and the number of the day of the
// school week (0, ..., 6) by time.
func getWeekAndWeekDayNumbersByTime(time time.Time) (weekNum int, weekDayNum int) {
	weekDayNum = int(time.Weekday()) - 1
	if weekDayNum == -1 {
		weekDayNum = 6
	}
	_, currWeekNumWithDelta := time.ISOWeek()
	weekNum = (currWeekNumWithDelta + 1) % 2
	return
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
