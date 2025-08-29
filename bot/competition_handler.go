package bot

import (
	"discord-bot/constants"
	"discord-bot/errors"
	"discord-bot/models"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// CompetitionHandler는 대회 관련 명령어를 처리합니다
type CompetitionHandler struct {
	commandHandler *CommandHandler
}

// NewCompetitionHandler는 새로운 CompetitionHandler 인스턴스를 생성합니다
func NewCompetitionHandler(ch *CommandHandler) *CompetitionHandler {
	return &CompetitionHandler{
		commandHandler: ch,
	}
}

// HandleCompetition은 대회 관련 명령어를 처리합니다
func (ch *CompetitionHandler) HandleCompetition(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	// DM이 아닌 경우에만 관리자 권한 확인
	if m.GuildID != "" && !ch.commandHandler.isAdmin(s, m) {
		err := errors.NewValidationError("INSUFFICIENT_PERMISSIONS",
			"User does not have administrator permissions",
			"❌  관리자 권한이 필요합니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	if len(params) == 0 {
		err := errors.NewValidationError("COMPETITION_INVALID_PARAMS",
			"Invalid competition parameters",
			"사용법: `!대회 <create|status|blackout|update>`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	subCommand := params[0]
	switch subCommand {
	case "create":
		ch.handleCompetitionCreate(s, m, params[1:])
	case "status":
		ch.handleCompetitionStatus(s, m)
	case "blackout":
		ch.handleCompetitionBlackout(s, m, params[1:])
	case "update":
		ch.handleCompetitionUpdate(s, m, params[1:])
	default:
		err := errors.NewValidationError("COMPETITION_UNKNOWN_COMMAND",
			fmt.Sprintf("Unknown competition command: %s", subCommand),
			"알 수 없는 명령어입니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
	}
}

func (ch *CompetitionHandler) handleCompetitionCreate(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	if len(params) < 3 {
		err := errors.NewValidationError("COMPETITION_CREATE_INVALID_PARAMS",
			"Invalid competition create parameters",
			"사용법: `!대회 create <대회명> <시작일> <종료일>`\n예시: `!대회 create 2024알고리즘대회 2024-01-01 2024-01-21`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	name := params[0]
	startDateStr := params[1]
	endDateStr := params[2]

	startDate, err := time.Parse(constants.DateFormat, startDateStr)
	if err != nil {
		botErr := errors.NewValidationError("INVALID_START_DATE",
			fmt.Sprintf("Invalid start date format: %s", startDateStr),
			"시작일 형식이 올바르지 않습니다. (YYYY-MM-DD)")
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	endDate, err := time.Parse(constants.DateFormat, endDateStr)
	if err != nil {
		botErr := errors.NewValidationError("INVALID_END_DATE",
			fmt.Sprintf("Invalid end date format: %s", endDateStr),
			"종료일 형식이 올바르지 않습니다. (YYYY-MM-DD)")
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	if endDate.Before(startDate) {
		err := errors.NewValidationError("INVALID_DATE_RANGE",
			"End date is before start date",
			"종료일이 시작일보다 빠릅니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	err = ch.commandHandler.storage.CreateCompetition(name, startDate, endDate)
	if err != nil {
		botErr := errors.NewSystemError("COMPETITION_CREATE_FAILED",
			"Failed to create competition", err)
		botErr.UserMsg = "대회 생성에 실패했습니다."
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	blackoutStart := endDate.AddDate(0, 0, -constants.BlackoutDays)
	response := fmt.Sprintf("🏆 **대회가 생성되었습니다!**\n"+
		"📝 대회명: %s\n"+
		"📅 기간: %s ~ %s\n"+
		"🔒 블랙아웃: %s ~ %s\n"+
		"✅ 상태: active",
		name,
		startDate.Format(constants.DateFormat),
		endDate.Format(constants.DateFormat),
		blackoutStart.Format(constants.DateFormat),
		endDate.Format(constants.DateFormat))

	errors.SendDiscordSuccess(s, m.ChannelID, response)
}

func (ch *CompetitionHandler) handleCompetitionStatus(s *discordgo.Session, m *discordgo.MessageCreate) {
	competition := ch.commandHandler.storage.GetCompetition()
	if competition == nil {
		err := errors.NewNotFoundError("NO_ACTIVE_COMPETITION",
			"No active competition found",
			"활성화된 대회가 없습니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	now := time.Now()
	status := "진행 중"
	if now.Before(competition.StartDate) {
		status = "시작 전"
	} else if now.After(competition.EndDate) {
		status = "종료됨"
	}

	blackoutStatus := "공개"
	if ch.commandHandler.storage.IsBlackoutPeriod() {
		blackoutStatus = "비공개 (블랙아웃)"
	}

	response := fmt.Sprintf("🏆 **%s** 대회가 진행 중입니다!\n"+
		"📅 **기간:** %s ~ %s\n"+
		"📊 **상태:** %s\n"+
		"🔒 **스코어보드:** %s\n"+
		"👥 **참가자 수:** %d명",
		competition.Name,
		competition.StartDate.Format(constants.DateFormat),
		competition.EndDate.Format(constants.DateFormat),
		status,
		blackoutStatus,
		len(ch.commandHandler.storage.GetParticipants()))

	s.ChannelMessageSend(m.ChannelID, response)
}

func (ch *CompetitionHandler) handleCompetitionBlackout(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	if len(params) == 0 {
		err := errors.NewValidationError("BLACKOUT_INVALID_PARAMS",
			"Invalid blackout parameters",
			"사용법: `!대회 blackout <on|off>`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	setting := strings.ToLower(params[0])
	var visible bool

	switch setting {
	case "on":
		visible = false
	case "off":
		visible = true
	default:
		err := errors.NewValidationError("BLACKOUT_INVALID_SETTING",
			fmt.Sprintf("Invalid blackout setting: %s", setting),
			"올바른 설정값을 입력하세요: `on` 또는 `off`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	err := ch.commandHandler.storage.SetScoreboardVisibility(visible)
	if err != nil {
		botErr := errors.NewSystemError("BLACKOUT_SETTING_FAILED",
			"Failed to set scoreboard visibility", err)
		botErr.UserMsg = "설정 변경에 실패했습니다."
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	status := "공개"
	if !visible {
		status = "비공개"
	}

	message := fmt.Sprintf("스코어보드가 **%s**로 설정되었습니다.", status)
	errors.SendDiscordSuccess(s, m.ChannelID, message)
}

func (ch *CompetitionHandler) handleCompetitionUpdate(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	if len(params) < 2 {
		err := errors.NewValidationError("COMPETITION_UPDATE_INVALID_PARAMS",
			"Invalid competition update parameters",
			"사용법: `!대회 update <필드> <값>`\n필드: name, start, end\n예시: `!대회 update name 새로운대회명`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	field := strings.ToLower(params[0])
	value := strings.Join(params[1:], " ")

	competition := ch.commandHandler.storage.GetCompetition()
	if competition == nil {
		err := errors.NewNotFoundError("NO_ACTIVE_COMPETITION",
			"No active competition found",
			"수정할 대회가 없습니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	switch field {
	case "name":
		ch.handleUpdateName(s, m, value, competition.Name)
	case "start":
		ch.handleUpdateStartDate(s, m, value, competition)
	case "end":
		ch.handleUpdateEndDate(s, m, value, competition)
	default:
		err := errors.NewValidationError("INVALID_UPDATE_FIELD",
			fmt.Sprintf("Invalid field: %s", field),
			"올바르지 않은 필드입니다. 사용 가능한 필드: name, start, end")
		errors.HandleDiscordError(s, m.ChannelID, err)
	}
}

func (ch *CompetitionHandler) handleUpdateName(s *discordgo.Session, m *discordgo.MessageCreate, newName, oldName string) {
	if newName == "" {
		err := errors.NewValidationError("EMPTY_COMPETITION_NAME",
			"Competition name cannot be empty",
			"대회명이 비어있습니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	err := ch.commandHandler.storage.UpdateCompetitionName(newName)
	if err != nil {
		botErr := errors.NewSystemError("COMPETITION_UPDATE_FAILED",
			"Failed to update competition name", err)
		botErr.UserMsg = "대회명 수정에 실패했습니다."
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	message := fmt.Sprintf("대회명이 **%s**에서 **%s**로 변경되었습니다.", oldName, newName)
	errors.SendDiscordSuccess(s, m.ChannelID, message)
}

func (ch *CompetitionHandler) handleUpdateStartDate(s *discordgo.Session, m *discordgo.MessageCreate, dateStr string, competition *models.Competition) {
	startDate, err := time.Parse(constants.DateFormat, dateStr)
	if err != nil {
		botErr := errors.NewValidationError("INVALID_START_DATE",
			fmt.Sprintf("Invalid start date format: %s", dateStr),
			"시작일 형식이 올바르지 않습니다. (YYYY-MM-DD)")
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	if startDate.After(competition.EndDate) {
		err := errors.NewValidationError("INVALID_DATE_RANGE",
			"Start date cannot be after end date",
			"시작일이 종료일보다 늦을 수 없습니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	oldDate := competition.StartDate
	err = ch.commandHandler.storage.UpdateCompetitionStartDate(startDate)
	if err != nil {
		botErr := errors.NewSystemError("COMPETITION_UPDATE_FAILED",
			"Failed to update competition start date", err)
		botErr.UserMsg = "시작일 수정에 실패했습니다."
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	message := fmt.Sprintf("시작일이 **%s**에서 **%s**로 변경되었습니다.",
		oldDate.Format(constants.DateFormat), startDate.Format(constants.DateFormat))
	errors.SendDiscordSuccess(s, m.ChannelID, message)
}

func (ch *CompetitionHandler) handleUpdateEndDate(s *discordgo.Session, m *discordgo.MessageCreate, dateStr string, competition *models.Competition) {
	endDate, err := time.Parse(constants.DateFormat, dateStr)
	if err != nil {
		botErr := errors.NewValidationError("INVALID_END_DATE",
			fmt.Sprintf("Invalid end date format: %s", dateStr),
			"종료일 형식이 올바르지 않습니다. (YYYY-MM-DD)")
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	if endDate.Before(competition.StartDate) {
		err := errors.NewValidationError("INVALID_DATE_RANGE",
			"End date cannot be before start date",
			"종료일이 시작일보다 빠를 수 없습니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	oldDate := competition.EndDate
	err = ch.commandHandler.storage.UpdateCompetitionEndDate(endDate)
	if err != nil {
		botErr := errors.NewSystemError("COMPETITION_UPDATE_FAILED",
			"Failed to update competition end date", err)
		botErr.UserMsg = "종료일 수정에 실패했습니다."
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	message := fmt.Sprintf("종료일이 **%s**에서 **%s**로 변경되었습니다.",
		oldDate.Format(constants.DateFormat), endDate.Format(constants.DateFormat))
	errors.SendDiscordSuccess(s, m.ChannelID, message)
}