package schedule

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/ulstu-schedule/parser/mock"
	"os"
	"regexp"
	"testing"
)

func TestParseCurrWeekTeacherScheduleImg(t *testing.T) {
	teacherSchedule := mock.TestTeacherSchedule(t)

	pathImg, err := ParseCurrWeekTeacherScheduleImg(&teacherSchedule.Weeks[0], "Зенкина С М")
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(pathImg))

	if pathImg != "" {
		defer func(path string) {
			err = os.Remove(path)
			assert.NoError(t, err)
		}(pathImg)
	}
}

func TestParseNextWeekTeacherScheduleImg(t *testing.T) {
	teacherSchedule := mock.TestTeacherSchedule(t)

	pathImg, err := ParseNextWeekTeacherScheduleImg(&teacherSchedule.Weeks[1], "Зенкина С М")
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(pathImg))

	if pathImg != "" {
		defer func(path string) {
			err = os.Remove(path)
			assert.NoError(t, err)
		}(pathImg)
	}
}

func TestConvertDayTeacherScheduleToText(t *testing.T) {
	t.Run("current day", func(t *testing.T) {
		teacherSchedule := mock.TestTeacherSchedule(t)

		result := ConvertDayTeacherScheduleToText("Зенкина С М", teacherSchedule.Weeks[0].Days[0], 0)

		findStart := regexp.MustCompile(fmt.Sprintf(`%s проводит следующие пары сегодня`, "Зенкина С М"))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
	t.Run("next day", func(t *testing.T) {
		teacherSchedule := mock.TestTeacherSchedule(t)

		result := ConvertDayTeacherScheduleToText("Зенкина С М", teacherSchedule.Weeks[0].Days[0], 1)

		findStart := regexp.MustCompile(fmt.Sprintf(`%s проводит следующие пары завтра`, "Зенкина С М"))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
	t.Run("other day", func(t *testing.T) {
		teacherSchedule := mock.TestTeacherSchedule(t)

		result := ConvertDayTeacherScheduleToText("Зенкина С М", teacherSchedule.Weeks[0].Days[0], 2)

		dayStr := getDateStr(2)
		findStart := regexp.MustCompile(fmt.Sprintf(`%s проводит следующие пары %s`, "Зенкина С М", dayStr))

		assert.EqualValues(t, true, findStart.MatchString(result))
	})
}
