package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"golang.org/x/text/encoding/charmap"
	"net/http"
	"strings"
)

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
