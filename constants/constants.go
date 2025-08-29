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
	SolvedACBaseURL    = "https://solved.ac/api/v3"
	APITimeout         = 30 * time.Second
	MaxRetries         = 3
	RetryDelay         = 1 * time.Second
)

// 점수 계산 상수
const (
	ChallengeMultiplier = 1.4
	BaseMultiplier      = 1.0 
	PenaltyMultiplier   = 0.5
	Top100ProblemCount  = 100
)

// 대회 관련 상수
const (
	BlackoutDays       = 3
	DefaultCompetitionID = 1
	DailyScoreboardHour = 9
	DailyScoreboardMinute = 0
)

// Discord 관련 상수
const (
	CommandPrefix = "!"
	MaxMessageLength = 2000
	EmbedColorSuccess = 0x00ff00
	EmbedColorError   = 0xff0000
	EmbedColorInfo    = 0x0099ff
)

// 이모지 상수
const (
	EmojiSuccess = "✅"
	EmojiError   = "❌"
	EmojiInfo    = "ℹ️"
	EmojiWarning = "⚠️"
	EmojiTrophy  = "🏆"
	EmojiUser    = "👤"
	EmojiTarget  = "🎯"
	EmojiMedal   = "🏅"
	EmojiStats   = "📊"
	EmojiCalendar = "📅"
	EmojiClock    = "⏰"
	EmojiLock     = "🔒"
	EmojiPeople   = "👥"
)

// 날짜 형식
const (
	DateFormat = "2006-01-02"
	TimeFormat = "15:04:05"
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
	MaxProblemTitleLength = 30
	TruncateIndicator = "..."
)

// 환경 변수 키
const (
	EnvDiscordToken   = "DISCORD_BOT_TOKEN"
	EnvChannelID      = "DISCORD_CHANNEL_ID"
	EnvLogLevel       = "LOG_LEVEL"
	EnvDebugMode      = "DEBUG_MODE"
)