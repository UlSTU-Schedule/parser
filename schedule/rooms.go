package schedule

import (
	"fmt"
	"github.com/ulstu-schedule/parser/types"
	"strings"
)

// ConvertDayRoomScheduleToText converts the information that types.Day contains into text.
func ConvertDayRoomScheduleToText(roomName string, daySchedule types.Day, daysAfterCurr int) string {
	result := strings.Builder{}

	dateStr := getDateStr(daysAfterCurr)
	weekNum, weekDayNum := GetWeekAndWeekDayNumbers(daysAfterCurr)

	switch daysAfterCurr {
	case 0:
		_, _ = fmt.Fprintf(&result, "Расписание кабинента %s на сегодня (%s, %s, %d-ая учебная неделя):\n\n",
			roomName, weekDays[weekDayNum], dateStr, weekNum+1)
	case 1:
		_, _ = fmt.Fprintf(&result, "Расписание кабинента %s на завтра (%s, %s, %d-ая учебная неделя):\n\n",
			roomName, weekDays[weekDayNum], dateStr, weekNum+1)
	default:
		_, _ = fmt.Fprintf(&result, "Расписание кабинента %s на %s (%s, %d-ая учебная неделя):\n\n", roomName,
			dateStr, weekDays[weekDayNum], weekNum+1)
	}

	noLessons := true
	for lessonIndex := 0; lessonIndex < len(daySchedule.Lessons); lessonIndex++ {
		subLessons := daySchedule.Lessons[lessonIndex].SubLessons

		if len(subLessons) > 0 {
			noLessons = false
			result.WriteString(daySchedule.Lessons[lessonIndex].StringRoomLesson())
		}
	}

	if noLessons {
		switch daysAfterCurr {
		case 0:
			result.WriteString("Сегодня пар нет")
		case 1:
			result.WriteString("Завтра пар нет")
		default:
			_, _ = fmt.Fprintf(&result, "%s пар нет", dateStr)
		}
	}

	return result.String()
}
