package utils

import (
	"discord-bot/constants"
	"fmt"
	"time"
)

// FormatDateRange는 날짜 범위를 포맷팅합니다
func FormatDateRange(start, end time.Time) string {
	return fmt.Sprintf("%s ~ %s",
		start.Format(constants.DateFormat),
		end.Format(constants.DateFormat))
}

// FormatDate는 단일 날짜를 포맷팅합니다
func FormatDate(date time.Time) string {
	return date.Format(constants.DateFormat)
}

// FormatDateTime은 날짜와 시간을 포맷팅합니다
func FormatDateTime(dateTime time.Time) string {
	return dateTime.Format(constants.DateTimeFormat)
}

// FormatTime은 시간만 포맷팅합니다
func FormatTime(time time.Time) string {
	return time.Format(constants.TimeFormat)
}