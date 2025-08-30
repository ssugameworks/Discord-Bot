package bot

import (
	"discord-bot/constants"
	"discord-bot/errors"
	"discord-bot/interfaces"
	"discord-bot/models"
	"discord-bot/scoring"
	"discord-bot/utils"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	storage            interfaces.StorageRepository
	scoreboardManager  *ScoreboardManager
	client             interfaces.APIClient
	competitionHandler *CompetitionHandler
}

func NewCommandHandler(storage interfaces.StorageRepository, apiClient interfaces.APIClient, scoreboardManager *ScoreboardManager) *CommandHandler {
	ch := &CommandHandler{
		storage:           storage,
		scoreboardManager: scoreboardManager,
		client:            apiClient,
	}
	ch.competitionHandler = NewCompetitionHandler(ch)
	return ch
}

func (ch *CommandHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if ch.shouldIgnoreMessage(s, m) {
		return
	}

	command, params, isDM := ch.parseMessage(m)
	if command == "" {
		return
	}

	ch.routeCommand(s, m, command, params, isDM)
}

// shouldIgnoreMessage 메시지를 무시해야 하는지 확인합니다
func (ch *CommandHandler) shouldIgnoreMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// 봇 자신의 메시지는 무시
	if m.Author.ID == s.State.User.ID {
		return true
	}

	// DM 디버깅 로그
	if m.GuildID == "" {
		fmt.Printf(constants.DMReceivedTemplate, m.Content, m.Author.Username)
	}

	return false
}

// parseMessage 메시지를 파싱하여 명령어와 매개변수를 추출합니다
func (ch *CommandHandler) parseMessage(m *discordgo.MessageCreate) (command string, params []string, isDM bool) {
	content := strings.TrimSpace(m.Content)
	if !strings.HasPrefix(content, constants.CommandPrefix) {
		return "", nil, false
	}

	args := strings.Fields(content)
	if len(args) == 0 {
		return "", nil, false
	}

	command = args[0][constants.CommandPrefixLength:]
	params = args[1:]
	isDM = m.GuildID == ""

	return command, params, isDM
}

// routeCommand 명령어를 해당 핸들러로 라우팅합니다
func (ch *CommandHandler) routeCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, params []string, isDM bool) {
	switch command {
	case "help", "도움말":
		ch.handleHelp(s, m)
	case "register", "등록":
		ch.handleRegister(s, m, params)
	case "scoreboard", "스코어보드":
		ch.handleScoreboardCommand(s, m, isDM)
	case "competition", "대회":
		ch.competitionHandler.HandleCompetition(s, m, params)
	case "participants", "참가자":
		ch.handleParticipants(s, m)
	case "remove", "삭제":
		ch.handleRemoveParticipant(s, m, params)
	case "ping":
		ch.handlePing(s, m)
	}
}

// handleScoreboardCommand 스코어보드 명령어를 처리합니다 (DM 체크 포함)
func (ch *CommandHandler) handleScoreboardCommand(s *discordgo.Session, m *discordgo.MessageCreate, isDM bool) {
	if isDM {
		s.ChannelMessageSend(m.ChannelID, "❌ 스코어보드는 서버에서만 확인할 수 있습니다.")
		return
	}
	ch.handleScoreboard(s, m)
}

// handlePing ping 명령어를 처리합니다
func (ch *CommandHandler) handlePing(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Pong! 🏓")
}

func (ch *CommandHandler) handleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	helpText := `🤖 **알고리즘 경진대회 봇 명령어**

**참가자 명령어:**
• ` + "`!등록 <이름> <백준ID>`" + ` - 대회 등록 신청
• ` + "`!스코어보드`" + ` - 현재 스코어보드 확인
• ` + "`!참가자`" + ` - 참가자 목록 확인

**관리자 명령어:**
• ` + "`!대회 create <대회명> <시작일> <종료일>`" + ` - 대회 생성 (YYYY-MM-DD 형식)
• ` + "`!대회 status`" + ` - 대회 상태 확인
• ` + "`!대회 blackout <on/off>`" + ` - 스코어보드 공개/비공개 설정
• ` + "`!대회 update <필드> <값>`" + ` - 대회 정보 수정 (name, start, end)
• ` + "`!삭제 <백준ID>`" + ` - 참가자 삭제

**기타:**
• ` + "`!ping`" + ` - 봇 응답 확인
• ` + "`!도움말`" + ` - 도움말 표시`

	s.ChannelMessageSend(m.ChannelID, helpText)
}

func (ch *CommandHandler) handleRegister(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)
	
	if len(params) < 2 {
		errorHandlers.Validation().HandleInvalidParams("REGISTER_INVALID_PARAMS",
			"Invalid register parameters",
			"사용법: `!등록 <이름> <백준ID>`")
		return
	}

	name := params[0]
	baekjoonID := params[1]

	userInfo, err := ch.client.GetUserInfo(baekjoonID)
	if err != nil {
		errorHandlers.API().HandleBaekjoonUserNotFound(baekjoonID, err)
		return
	}

	err = ch.storage.AddParticipant(name, baekjoonID, userInfo.Tier, userInfo.Rating)
	if err != nil {
		errorHandlers.Data().HandleParticipantAlreadyExists(baekjoonID)
		return
	}

	tierName := getTierName(userInfo.Tier)
	tm := models.NewTierManager()
	colorCode := tm.GetTierANSIColor(userInfo.Tier)
	
	response := fmt.Sprintf("```ansi\n%s%s(%s)%s님 성공적으로 등록되었습니다!\n```", 
		colorCode, name, tierName, tm.GetANSIReset())

	s.ChannelMessageSend(m.ChannelID, response)
}

func (ch *CommandHandler) handleScoreboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)
	
	isAdmin := ch.isAdmin(s, m)
	embed, err := ch.scoreboardManager.GenerateScoreboard(isAdmin)
	if err != nil {
		errorHandlers.System().HandleScoreboardGenerationFailed(err)
		return
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func (ch *CommandHandler) handleParticipants(s *discordgo.Session, m *discordgo.MessageCreate) {
	participants := ch.storage.GetParticipants()
	if len(participants) == 0 {
		errors.SendDiscordInfo(s, m.ChannelID, "참가자가 없습니다.")
		return
	}

	var sb strings.Builder
	sb.WriteString("```ansi\n")

	tm := models.NewTierManager()
	for i, p := range participants {
		tierName := getTierName(p.StartTier)
		colorCode := tm.GetTierANSIColor(p.StartTier)
		sb.WriteString(fmt.Sprintf("%s%d. %s (%s) - %s%s\n",
			colorCode, i+1, p.Name, p.BaekjoonID, tierName, tm.GetANSIReset()))
	}

	sb.WriteString("```")
	s.ChannelMessageSend(m.ChannelID, sb.String())
}

func (ch *CommandHandler) handleRemoveParticipant(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	errorHandlers := utils.NewErrorHandlerFactory(s, m.ChannelID)
	
	// 관리자 권한 확인
	if !ch.isAdmin(s, m) {
		errorHandlers.Validation().HandleInsufficientPermissions()
		return
	}

	// 파라미터 확인
	if len(params) < 1 {
		errorHandlers.Validation().HandleInvalidParams("REMOVE_INVALID_PARAMS",
			"Invalid remove parameters",
			"사용법: `!삭제 <백준ID>`")
		return
	}

	baekjoonID := params[0]

	// 백준 ID 유효성 검사
	if !utils.IsValidBaekjoonID(baekjoonID) {
		errorHandlers.Validation().HandleInvalidParams("REMOVE_INVALID_BAEKJOON_ID",
			"Invalid Baekjoon ID format",
			"유효하지 않은 백준 ID 형식입니다.")
		return
	}

	// 참가자 삭제
	err := ch.storage.RemoveParticipant(baekjoonID)
	if err != nil {
		errorHandlers.Data().HandleParticipantNotFound(baekjoonID)
		return
	}

	response := fmt.Sprintf("✅ **참가자 삭제 완료**\n🎯 백준ID: %s", baekjoonID)
	s.ChannelMessageSend(m.ChannelID, response)
}

// isAdmin는 사용자가 서버 관리자 권한을 가지고 있는지 확인합니다
func (ch *CommandHandler) isAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// 길드 정보 가져오기
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		return false
	}

	// 서버 소유자인지 확인
	if m.Author.ID == guild.OwnerID {
		return true
	}

	// 멤버 정보 가져오기
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		return false
	}

	// 멤버의 역할들을 확인
	for _, roleID := range member.Roles {
		role, err := s.State.Role(m.GuildID, roleID)
		if err != nil {
			continue
		}

		// 관리자 권한(ADMINISTRATOR) 확인
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}

	return false
}

func getTierName(tier int) string {
	return scoring.GetTierName(tier)
}
