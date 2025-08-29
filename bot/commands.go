package bot

import (
	"discord-bot/api"
	"discord-bot/constants"
	"discord-bot/errors"
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
	if len(params) < 2 {
		err := errors.NewValidationError("REGISTER_INVALID_PARAMS",
			"Invalid register parameters",
			"사용법: `!등록 <이름> <백준ID>`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	name := params[0]
	baekjoonID := params[1]

	userInfo, err := ch.client.GetUserInfo(baekjoonID)
	if err != nil {
		botErr := errors.NewAPIError("BAEKJOON_USER_NOT_FOUND",
			fmt.Sprintf("Baekjoon user '%s' not found", baekjoonID), err)
		botErr.UserMsg = fmt.Sprintf("백준 사용자 '%s'를 찾을 수 없습니다.", baekjoonID)
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	err = ch.storage.AddParticipant(name, baekjoonID, userInfo.Tier, userInfo.Rating)
	if err != nil {
		botErr := errors.NewDuplicateError("PARTICIPANT_ALREADY_EXISTS",
			fmt.Sprintf("Participant with Baekjoon ID '%s' already exists", baekjoonID),
			fmt.Sprintf("백준 ID '%s'로 이미 등록된 참가자가 있습니다.", baekjoonID))
		errors.HandleDiscordError(s, m.ChannelID, botErr)
		return
	}

	tierName := getTierName(userInfo.Tier)
	colorCode := constants.GetTierANSIColor(userInfo.Tier)
	
	response := fmt.Sprintf("```ansi\n%s%s(%s)%s님 성공적으로 등록되었습니다!\n```", 
		colorCode, name, tierName, constants.ANSIReset)

	s.ChannelMessageSend(m.ChannelID, response)
}

func (ch *CommandHandler) handleScoreboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	isAdmin := ch.isAdmin(s, m)
	embed, err := ch.scoreboardManager.GenerateScoreboard(isAdmin)
	if err != nil {
		botErr := errors.NewSystemError("SCOREBOARD_GENERATION_FAILED",
			"Failed to generate scoreboard", err)
		botErr.UserMsg = "스코어보드 생성에 실패했습니다."
		errors.HandleDiscordError(s, m.ChannelID, botErr)
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

	for i, p := range participants {
		tierName := getTierName(p.StartTier)
		colorCode := constants.GetTierANSIColor(p.StartTier)
		sb.WriteString(fmt.Sprintf("%s%d. %s (%s) - %s%s\n",
			colorCode, i+1, p.Name, p.BaekjoonID, tierName, constants.ANSIReset))
	}

	sb.WriteString("```")
	s.ChannelMessageSend(m.ChannelID, sb.String())
}

func (ch *CommandHandler) handleRemoveParticipant(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {
	// 관리자 권한 확인
	if !ch.isAdmin(s, m) {
		s.ChannelMessageSend(m.ChannelID, "❌ 이 명령어는 관리자만 사용할 수 있습니다.")
		return
	}

	// 파라미터 확인
	if len(params) < 1 {
		err := errors.NewValidationError("REMOVE_INVALID_PARAMS",
			"Invalid remove parameters",
			"사용법: `!삭제 <백준ID>`")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	baekjoonID := params[0]

	// 백준 ID 유효성 검사
	if !utils.IsValidBaekjoonID(baekjoonID) {
		err := errors.NewValidationError("REMOVE_INVALID_BAEKJOON_ID",
			"Invalid Baekjoon ID format",
			"유효하지 않은 백준 ID 형식입니다.")
		errors.HandleDiscordError(s, m.ChannelID, err)
		return
	}

	// 참가자 삭제
	err := ch.storage.RemoveParticipant(baekjoonID)
	if err != nil {
		botErr := errors.NewNotFoundError("PARTICIPANT_NOT_FOUND",
			fmt.Sprintf("Participant with Baekjoon ID '%s' not found", baekjoonID),
			fmt.Sprintf("백준 ID '%s'로 등록된 참가자를 찾을 수 없습니다.", baekjoonID))
		errors.HandleDiscordError(s, m.ChannelID, botErr)
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
