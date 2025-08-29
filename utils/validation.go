package utils

import (
	"discord-bot/constants"
	"regexp"
	"strings"
	"time"
)

// 문자열 유효성 검사
func IsValidUsername(username string) bool {
	if len(username) == 0 || len(username) > 50 {
		return false
	}
	// 특수문자 제한 (기본적인 문자, 숫자, 한글, 공백만 허용)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9가-힣\s]+$`, username)
	return matched
}

func IsValidBaekjoonID(id string) bool {
	if len(id) == 0 || len(id) > 20 {
		return false
	}
	// 백준 ID는 영문, 숫자, 언더스코어만 허용
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, id)
	return matched
}

// 날짜 유효성 검사
func IsValidDateString(dateStr string) bool {
	_, err := time.Parse(constants.DateFormat, dateStr)
	return err == nil
}

func IsValidDateRange(startDate, endDate time.Time) bool {
	return !endDate.Before(startDate)
}

// 문자열 처리
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= len(constants.TruncateIndicator) {
		return constants.TruncateIndicator[:maxLen]
	}
	return s[:maxLen-len(constants.TruncateIndicator)] + constants.TruncateIndicator
}

func SanitizeString(s string) string {
	// Discord 메시지에서 문제가 될 수 있는 특수문자 제거/변경
	s = strings.ReplaceAll(s, "`", "'")
	s = strings.ReplaceAll(s, "@", "(at)")
	return strings.TrimSpace(s)
}

// 슬라이스 유틸리티
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 안전한 정수 변환
func SafeIntAdd(a, b int) int {
	const maxInt = int(^uint(0) >> 1)
	if a > 0 && b > maxInt-a {
		return maxInt
	}
	if a < 0 && b < -maxInt-a {
		return -maxInt
	}
	return a + b
}
