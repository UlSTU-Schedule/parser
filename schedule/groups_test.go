package schedule

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/ulstu-schedule/parser/mock"
	"os"
	"regexp"
	"testing"
)

func TestParseCurrWeekGroupScheduleImg(t *testing.T) {
	groupSchedule := mock.TestGroupSchedule(t)

	pathImg, err := ParseCurrWeekGroupScheduleImg(&groupSchedule.Weeks[0], "АТсд-11")
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(pathImg))

	if pathImg != "" {
		defer func(path string) {
			err = os.Remove(path)
			assert.NoError(t, err)
		}(pathImg)
	}
}

func TestParseNextWeekGroupScheduleImg(t *testing.T) {
	groupSchedule := mock.TestGroupSchedule(t)

	pathImg, err := ParseNextWeekGroupScheduleImg(&groupSchedule.Weeks[1], "АТсд-11")
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(pathImg))

	if pathImg != "" {
		defer func(path string) {
			err = os.Remove(path)
			assert.NoError(t, err)
		}(pathImg)
	}
}

func TestConvertDayGroupScheduleToText(t *testing.T) {
	t.Run("current day", func(t *testing.T) {
		groupSchedule := mock.TestGroupSchedule(t)

		result := ConvertDayGroupScheduleToText(&groupSchedule.Weeks[0].Days[0], "АТсд-11", 0)

		findStart := regexp.MustCompile(fmt.Sprintf(`Расписание %s на сегодня`, "АТсд-11"))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
	t.Run("next day", func(t *testing.T) {
		groupSchedule := mock.TestGroupSchedule(t)

		result := ConvertDayGroupScheduleToText(&groupSchedule.Weeks[0].Days[0], "АТсд-11", 1)

		findStart := regexp.MustCompile(fmt.Sprintf(`Расписание %s на завтра`, "АТсд-11"))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
	t.Run("other day", func(t *testing.T) {
		groupSchedule := mock.TestGroupSchedule(t)

		result := ConvertDayGroupScheduleToText(&groupSchedule.Weeks[0].Days[0], "АТсд-11", 2)

		dayStr := getDateStr(2)
		findStart := regexp.MustCompile(fmt.Sprintf(`Расписание %s на %s`, "АТсд-11", dayStr))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
}
