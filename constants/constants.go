package constants

import "time"

// 파일 관련 상수
const (
	ParticipantsFileName = "participants.json"
	CompetitionFileName  = "competition.json"
	FilePermission       = 0644
)

// API 관련 상수
const (
	SolvedACBaseURL = "https://solved.ac/api/v3"
	APITimeout      = 30 * time.Second
	MaxRetries      = 3
	RetryDelay      = 1 * time.Second
)

// 점수 계산 상수
const (
	ChallengeMultiplier = 1.4
	BaseMultiplier      = 1.0
	PenaltyMultiplier   = 0.5
)

// 대회 관련 상수
const (
	BlackoutDays          = 3
	DailyScoreboardHour   = 9
	DailyScoreboardMinute = 0
)

// Discord 관련 상수
const (
	CommandPrefix = "!"
)

// 이모지 상수
const (
	EmojiSuccess  = "✅"
	EmojiError    = "❌"
	EmojiInfo     = "ℹ️"
	EmojiWarning  = "⚠️"
	EmojiTrophy   = "🏆"
	EmojiUser     = "👤"
	EmojiTarget   = "🎯"
	EmojiMedal    = "🏅"
	EmojiStats    = "📊"
	EmojiCalendar = "📅"
	EmojiClock    = "⏰"
	EmojiLock     = "🔒"
	EmojiPeople   = "👥"
)

// 날짜 형식
const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
)

// 로그 관련 상수
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

// 문자열 크기 제한
const (
	MaxUsernameLength = 15
	TruncateIndicator = "..."
)

// 메시지 템플릿
const (
	DMReceivedTemplate  = "DM 수신: %s from %s\n"
	CommandPrefixLength = 1 // "!" 길이
)

// 티어별 색상
const (
	ColorTierBronze   = 0xA25B1F // 브론즈 - 갈색
	ColorTierSilver   = 0x495E78 // 실버 - 은색
	ColorTierGold     = 0xE09E37 // 골드 - 금색
	ColorTierPlatinum = 0x6DDFA8 // 플래티넘 - 다크터쿼이즈
	ColorTierDiamond  = 0x50B1F6 // 다이아몬드 - 닷저블루
	ColorTierRuby     = 0xEA3364 // 루비 - 루비색
	ColorTierMaster   = 0x8A2BE2 // 마스터 - 블루바이올렛
	ColorTierDefault  = 0x36393F // 기본/언랭크 - 디스코드 다크그레이
)

// ANSI 색상 코드 (티어별)
const (
	ANSIReset     = "\x1b[0m"
	ANSIBronze    = "\x1b[1;33m" // 밝은 노란색 (브론즈)
	ANSISilver    = "\x1b[1;37m" // 밝은 흰색 (실버)
	ANSIGold      = "\x1b[1;33m" // 밝은 노란색 (골드)
	ANSIPlatinum  = "\x1b[1;36m" // 밝은 청록색 (플래티넘)
	ANSIDiamond   = "\x1b[1;34m" // 밝은 파란색 (다이아몬드)
	ANSIRuby      = "\x1b[1;31m" // 밝은 빨간색 (루비)
	ANSIMaster    = "\x1b[1;35m" // 밝은 자주색 (마스터)
	ANSIDefault   = "\x1b[0m"    // 기본색
)

// GetTierANSIColor returns ANSI color code for a given tier
func GetTierANSIColor(tier int) string {
	switch {
	case tier >= 1 && tier <= 5:
		return ANSIBronze
	case tier >= 6 && tier <= 10:
		return ANSISilver
	case tier >= 11 && tier <= 15:
		return ANSIGold
	case tier >= 16 && tier <= 20:
		return ANSIPlatinum
	case tier >= 21 && tier <= 25:
		return ANSIDiamond
	case tier >= 26 && tier <= 30:
		return ANSIRuby
	case tier >= 31:
		return ANSIMaster
	default:
		return ANSIDefault
	}
}

// 환경 변수 키
const (
	EnvDiscordToken = "DISCORD_BOT_TOKEN"
	EnvChannelID    = "DISCORD_CHANNEL_ID"
	EnvLogLevel     = "LOG_LEVEL"
	EnvDebugMode    = "DEBUG_MODE"
)
