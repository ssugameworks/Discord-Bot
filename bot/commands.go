package bot

import (
	"discord-bot/api"
	"discord-bot/constants"
	"discord-bot/errors"
	"discord-bot/models"
	"discord-bot/scoring"
	"discord-bot/storage"
	"discord-bot/utils"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	storage            *storage.Storage
	scoreboardManager  *ScoreboardManager
	client             *api.SolvedACClient
	competitionHandler *CompetitionHandler
}

func NewCommandHandler(storage *storage.Storage) *CommandHandler {
	ch := &CommandHandler{
		storage:           storage,
		scoreboardManager: NewScoreboardManager(storage),
		client:            api.NewSolvedACClient(),
	}
	ch.competitionHandler = NewCompetitionHandler(ch)
	return ch
}

func (ch *CommandHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// DM 디버깅 로그
	if m.GuildID == "" {
		fmt.Printf(constants.DMReceivedTemplate, m.Content, m.Author.Username)
	}

	content := strings.TrimSpace(m.Content)
	if !strings.HasPrefix(content, constants.CommandPrefix) {
		return
	}

	args := strings.Fields(content)
	if len(args) == 0 {
		return
	}

	command := args[0][constants.CommandPrefixLength:]
	params := args[1:]

	// DM 처리 확인
	isDM := m.GuildID == ""

	switch command {
	case "help", "도움말":
		ch.handleHelp(s, m)
	case "register", "등록":
		ch.handleRegister(s, m, params)
	case "scoreboard", "스코어보드":
		if isDM {
			s.ChannelMessageSend(m.ChannelID, "❌ 스코어보드는 서버에서만 확인할 수 있습니다.")
			return
		}
		ch.handleScoreboard(s, m)
	case "competition", "대회":
		ch.competitionHandler.HandleCompetition(s, m, params)
	case "participants", "참가자":
		ch.handleParticipants(s, m)
	case "remove", "삭제":
		ch.handleRemoveParticipant(s, m, params)
	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong! 🏓")
	}
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
