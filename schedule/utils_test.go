package schedule

import (
	"github.com/stretchr/testify/assert"
	"github.com/ulstu-schedule/parser/mock"
	"testing"
	"time"
)

func TestIsInTimeRange(t *testing.T) {
	t.Run("correct", func(t *testing.T) {
		weekDate, _ := time.Parse("2006-01-02", "2024-03-13")
		startDate, _ := time.Parse("2006-01-02", "2024-03-12")
		endDate, _ := time.Parse("2006-01-02", "2024-03-15")

		isInRange := isInTimeRange(weekDate, startDate, endDate)

		assert.EqualValues(t, true, isInRange)
	})
	t.Run("weekDate < start", func(t *testing.T) {
		weekDate, _ := time.Parse("2006-01-02", "2024-03-10")
		startDate, _ := time.Parse("2006-01-02", "2024-03-12")
		endDate, _ := time.Parse("2006-01-02", "2024-03-15")

		isInRange := isInTimeRange(weekDate, startDate, endDate)

		assert.EqualValues(t, false, isInRange)
	})
	t.Run("weekDate > end", func(t *testing.T) {
		weekDate, _ := time.Parse("2006-01-02", "2024-03-19")
		startDate, _ := time.Parse("2006-01-02", "2024-03-12")
		endDate, _ := time.Parse("2006-01-02", "2024-03-15")

		isInRange := isInTimeRange(weekDate, startDate, endDate)

		assert.EqualValues(t, false, isInRange)
	})
}

func TestGetScheduleWeekNumDyDate(t *testing.T) {
	t.Run("correct. Get 12 week num", func(t *testing.T) {
		weekDate, _ := time.Parse("2006-01-02", "2024-04-24")

		groupSchedule := mock.TestGroupSchedule(t)
		weekNum := getScheduleWeekNumDyDate(groupSchedule, weekDate)

		assert.EqualValues(t, 1, weekNum)
	})
	t.Run("correct. Get 11 week num", func(t *testing.T) {
		weekDate, _ := time.Parse("2006-01-02", "2024-04-17")

		groupSchedule := mock.TestGroupSchedule(t)
		weekNum := getScheduleWeekNumDyDate(groupSchedule, weekDate)

		assert.EqualValues(t, 0, weekNum)
	})
	t.Run("invalid date", func(t *testing.T) {
		weekDate, _ := time.Parse("2006-01-02", "2024-04-10")

		groupSchedule := mock.TestGroupSchedule(t)
		weekNum := getScheduleWeekNumDyDate(groupSchedule, weekDate)

		assert.EqualValues(t, 0, weekNum)
	})
}
