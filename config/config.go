package config

import (
	"discord-bot/constants"
	"os"
	"strconv"
	"strings"
)

// Config 애플리케이션의 전체 설정을 관리합니다
type Config struct {
	Discord  DiscordConfig
	Schedule ScheduleConfig
	Logging  LoggingConfig
	Features FeatureFlags
}

type DiscordConfig struct {
	Token     string
	ChannelID string
}

type ScheduleConfig struct {
	ScoreboardHour   int
	ScoreboardMinute int
	Enabled          bool
}

type LoggingConfig struct {
	Level     string
	DebugMode bool
}

type FeatureFlags struct {
	EnableAutoScoreboard bool
	EnableDetailedErrors bool
}

// Load는 환경변수에서 설정을 로드합니다
func Load() *Config {
	return &Config{
		Discord: DiscordConfig{
			Token:     getEnv(constants.EnvDiscordToken, ""),
			ChannelID: getEnv(constants.EnvChannelID, ""),
		},
		Schedule: ScheduleConfig{
			ScoreboardHour:   getEnvInt("SCOREBOARD_HOUR", constants.DailyScoreboardHour),
			ScoreboardMinute: getEnvInt("SCOREBOARD_MINUTE", constants.DailyScoreboardMinute),
			Enabled:          getEnv(constants.EnvChannelID, "") != "",
		},
		Logging: LoggingConfig{
			Level:     getEnv(constants.EnvLogLevel, constants.LogLevelInfo),
			DebugMode: getEnvBool(constants.EnvDebugMode, false),
		},
		Features: FeatureFlags{
			EnableAutoScoreboard: getEnvBool("ENABLE_AUTO_SCOREBOARD", true),
			EnableDetailedErrors: getEnvBool("ENABLE_DETAILED_ERRORS", false),
		},
	}
}

// Validate 설정의 유효성을 검사합니다
func (c *Config) Validate() error {
	if c.Discord.Token == "" {
		return &ConfigError{
			Field:   "Discord.Token",
			Message: "Discord bot token is required",
		}
	}
	return nil
}

// IsDebugMode 디버그 모드 여부를 반환합니다
func (c *Config) IsDebugMode() bool {
	return c.Logging.DebugMode || strings.ToUpper(c.Logging.Level) == constants.LogLevelDebug
}

// ConfigError 설정 관련 오류를 나타냅니다
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error in " + e.Field + ": " + e.Message
}

// 헬퍼 함수들
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
