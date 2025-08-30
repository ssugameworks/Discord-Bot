package errors

import (
	"discord-bot/constants"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// ErrorType 오류의 종류를 나타냅니다
type ErrorType int

const (
	TypeValidation ErrorType = iota
	TypeAPI
	TypeNotFound
	TypeDuplicate
	TypePermission
	TypeCompetition
	TypeSystem
)

// AppError 애플리케이션에서 발생하는 구조화된 오류를 표현합니다
type AppError struct {
	Type     ErrorType
	Code     string
	Message  string
	UserMsg  string
	Internal error
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// GetUserMessage 사용자에게 표시할 메시지를 반환합니다
func (e *AppError) GetUserMessage() string {
	if e.UserMsg != "" {
		return e.UserMsg
	}
	return e.Message
}

// 오류 생성 함수들

// NewValidationError 입력값 검증 오류를 생성합니다
func NewValidationError(code, message, userMsg string) *AppError {
	return &AppError{
		Type:    TypeValidation,
		Code:    code,
		Message: message,
		UserMsg: userMsg,
	}
}

// NewAPIError 외부 API 연동 오류를 생성합니다
func NewAPIError(code, message string, err error) *AppError {
	return &AppError{
		Type:     TypeAPI,
		Code:     code,
		Message:  message,
		UserMsg:  "외부 서비스 연결에 문제가 발생했습니다. 잠시 후 다시 시도해주세요.",
		Internal: err,
	}
}

// NewNotFoundError 리소스를 찾을 수 없는 오류를 생성합니다
func NewNotFoundError(code, message, userMsg string) *AppError {
	return &AppError{
		Type:    TypeNotFound,
		Code:    code,
		Message: message,
		UserMsg: userMsg,
	}
}

// NewDuplicateError 중복 리소스 오류를 생성합니다
func NewDuplicateError(code, message, userMsg string) *AppError {
	return &AppError{
		Type:    TypeDuplicate,
		Code:    code,
		Message: message,
		UserMsg: userMsg,
	}
}

// NewPermissionError 권한 관련 오류를 생성합니다
func NewPermissionError(code, message, userMsg string) *AppError {
	return &AppError{
		Type:    TypePermission,
		Code:    code,
		Message: message,
		UserMsg: userMsg,
	}
}

// NewCompetitionError 대회 관련 오류를 생성합니다
func NewCompetitionError(code, message, userMsg string) *AppError {
	return &AppError{
		Type:    TypeCompetition,
		Code:    code,
		Message: message,
		UserMsg: userMsg,
	}
}

// NewSystemError 시스템 내부 오류를 생성합니다
func NewSystemError(code, message string, err error) *AppError {
	return &AppError{
		Type:     TypeSystem,
		Code:     code,
		Message:  message,
		UserMsg:  "시스템 오류가 발생했습니다. 관리자에게 문의해주세요.",
		Internal: err,
	}
}

// Discord 메시지 관련 헬퍼 함수들

// HandleDiscordError 오류를 처리하고 Discord 채널에 메시지를 전송합니다
func HandleDiscordError(s *discordgo.Session, channelID string, err error) {
	if appErr, ok := err.(*AppError); ok {
		// 로그에 상세 정보 기록
		if appErr.Internal != nil {
			fmt.Printf("ERROR: %s - %s: %v\n", appErr.Code, appErr.Message, appErr.Internal)
		} else {
			fmt.Printf("ERROR: %s - %s\n", appErr.Code, appErr.Message)
		}

		s.ChannelMessageSend(channelID, constants.EmojiError+" "+appErr.GetUserMessage())
	} else {
		// 예상치 못한 오류 로깅
		fmt.Printf("UNEXPECTED ERROR: %v\n", err)
		s.ChannelMessageSend(channelID, constants.EmojiError+" 예상치 못한 오류가 발생했습니다.")
	}
}

// SendDiscordSuccess 성공 메시지를 Discord 채널에 전송합니다
func SendDiscordSuccess(s *discordgo.Session, channelID, message string) {
	s.ChannelMessageSend(channelID, constants.EmojiSuccess+" "+message)
}

// SendDiscordInfo 정보 메시지를 Discord 채널에 전송합니다
func SendDiscordInfo(s *discordgo.Session, channelID, message string) {
	s.ChannelMessageSend(channelID, constants.EmojiInfo+" "+message)
}

// SendDiscordWarning 경고 메시지를 Discord 채널에 전송합니다
func SendDiscordWarning(s *discordgo.Session, channelID, message string) {
	s.ChannelMessageSend(channelID, constants.EmojiWarning+" "+message)
}
