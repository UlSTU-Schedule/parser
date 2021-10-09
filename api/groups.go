package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ulstu-schedule/parser/types"
	"strings"
)

const groupScheduleURL = "https://old.ulstu.ru/schedule/students/part%d/%s"

// GetGroupFullSchedule ...
func GetGroupFullSchedule(groupName string) (*types.GroupSchedule, error) {
	scheduleURL, err := getGroupURL(groupName)
	if err != nil {
		return nil, err
	}

	doc, err := getDocFromURL(scheduleURL)
	if err != nil {
		return nil, err
	}

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// first week
		if 22 <= i && i <= 79 && 2 <= i%10 && i%10 <= 9 {
			font := s.Find("font")
			fontHTML, _ := font.Html()
			if fontHTML != "" {
				for _, elem := range strings.Split(strings.Replace(fontHTML, "<br/>", "sep", -1), " sep") {
					if elem != "" {
						fmt.Println(elem)
					}
				}
			}
		}

		// second week
		if 113 <= i && i <= 170 && (i%10 == 0 || i%10 >= 3) {
			t := s.Find("font").Text()
			if t != "" {
				fmt.Println(i, t)
			}
		}
	})
	return nil, nil
}

// getGroupURL ...
func getGroupURL(groupName string) (string, error) {
	var groupURL string

	for schedulePartNum := 1; schedulePartNum < 4; schedulePartNum++ {
		doc, err := getDocFromURL(fmt.Sprintf(groupScheduleURL, schedulePartNum, "raspisan.html"))
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
							groupURL = fmt.Sprintf(groupScheduleURL, schedulePartNum, href)
							return false
						}
					}
				} else if foundGroupName == groupName {
					href, _ := s.Find("a").Attr("href")
					groupURL = fmt.Sprintf(groupScheduleURL, schedulePartNum, href)
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
