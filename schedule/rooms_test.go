package schedule

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/ulstu-schedule/parser/mock"
	"regexp"
	"testing"
)

func TestConvertDayRoomScheduleToText(t *testing.T) {
	t.Run("current day", func(t *testing.T) {
		teacherSchedule := mock.TestRoomSchedule(t)

		result := ConvertDayRoomScheduleToText("Зенкина С М", teacherSchedule.Weeks[0].Days[0], 0)

		findStart := regexp.MustCompile(fmt.Sprintf(`Расписание кабинента %s на сегодня`, "Зенкина С М"))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
	t.Run("next day", func(t *testing.T) {
		teacherSchedule := mock.TestRoomSchedule(t)

		result := ConvertDayRoomScheduleToText("Зенкина С М", teacherSchedule.Weeks[0].Days[0], 1)

		findStart := regexp.MustCompile(fmt.Sprintf(`Расписание кабинента %s на завтра`, "Зенкина С М"))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
	t.Run("other day", func(t *testing.T) {
		teacherSchedule := mock.TestRoomSchedule(t)

		result := ConvertDayRoomScheduleToText("Зенкина С М", teacherSchedule.Weeks[0].Days[0], 2)

		dayStr := getDateStr(2)
		findStart := regexp.MustCompile(fmt.Sprintf(`Расписание кабинента %s на %s`, "Зенкина С М", dayStr))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
}
