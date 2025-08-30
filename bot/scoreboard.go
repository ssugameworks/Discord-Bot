package bot

import (
	"discord-bot/api"
	"discord-bot/constants"
	"discord-bot/models"
	"discord-bot/scoring"
	"discord-bot/storage"
	"discord-bot/utils"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type ScoreboardManager struct {
	storage    *storage.Storage
	calculator *scoring.ScoreCalculator
	client     *api.SolvedACClient
}

func NewScoreboardManager(storage *storage.Storage) *ScoreboardManager {
	return &ScoreboardManager{
		storage:    storage,
		calculator: scoring.NewScoreCalculator(),
		client:     api.NewSolvedACClient(),
	}
}

func (sm *ScoreboardManager) GenerateScoreboard(isAdmin bool) (*discordgo.MessageEmbed, error) {
	competition := sm.storage.GetCompetition()
	if competition == nil || !competition.IsActive {
		return nil, fmt.Errorf("no active competition")
	}

	// 블랙아웃 체크
	if embed := sm.checkBlackoutPeriod(competition, isAdmin); embed != nil {
		return embed, nil
	}

	// 참가자 체크
	participants := sm.storage.GetParticipants()
	if embed := sm.checkEmptyParticipants(competition, participants); embed != nil {
		return embed, nil
	}

	// 점수 데이터 수집
	scores, err := sm.collectScoreData(participants)
	if err != nil {
		return nil, err
	}

	// 정렬 및 포맷팅
	sm.sortScores(scores)
	return sm.formatScoreboard(competition, scores, isAdmin), nil
}

// checkBlackoutPeriod 블랙아웃 기간인지 확인하고 해당 embed 반환
func (sm *ScoreboardManager) checkBlackoutPeriod(competition *models.Competition, isAdmin bool) *discordgo.MessageEmbed {
	if sm.storage.IsBlackoutPeriod() && competition.ShowScoreboard && !isAdmin {
		tm := models.NewTierManager()
		return &discordgo.MessageEmbed{
			Title:       "🔒 스코어보드 비공개",
			Description: "마지막 3일간 스코어보드가 비공개됩니다",
			Color:       tm.GetTierColor(0), // Unranked color
		}
	}
	return nil
}

// checkEmptyParticipants 참가자가 없는지 확인하고 해당 embed 반환
func (sm *ScoreboardManager) checkEmptyParticipants(competition *models.Competition, participants []models.Participant) *discordgo.MessageEmbed {
	if len(participants) == 0 {
		tm := models.NewTierManager()
		return &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("🏆 %s 스코어보드", competition.Name),
			Description: "참가자가 없습니다.",
			Color:       tm.GetTierColor(0), // Unranked color
		}
	}
	return nil
}

// collectScoreData 참가자들의 점수 데이터를 수집합니다
func (sm *ScoreboardManager) collectScoreData(participants []models.Participant) ([]models.ScoreData, error) {
	scores := make([]models.ScoreData, 0, len(participants))

	for _, participant := range participants {
		scoreData, err := sm.calculateParticipantScore(participant)
		if err != nil {
			// 개별 참가자 에러는 로그만 남기고 계속 진행
			continue
		}
		scores = append(scores, scoreData)
	}

	return scores, nil
}

// calculateParticipantScore 개별 참가자의 점수를 계산합니다
func (sm *ScoreboardManager) calculateParticipantScore(participant models.Participant) (models.ScoreData, error) {
	userInfo, err := sm.client.GetUserInfo(participant.BaekjoonID)
	if err != nil {
		return models.ScoreData{}, err
	}

	score, err := sm.calculator.CalculateScore(participant.BaekjoonID, participant.StartTier, participant.StartProblemIDs)
	if err != nil {
		return models.ScoreData{}, err
	}

	top100, err := sm.client.GetUserTop100(participant.BaekjoonID)
	if err != nil {
		return models.ScoreData{}, err
	}

	// 새로 푼 문제 수 계산 (현재 - 시작시점)
	newProblemCount := top100.Count - participant.StartProblemCount
	if newProblemCount < 0 {
		newProblemCount = 0
	}

	return models.ScoreData{
		ParticipantID: participant.ID,
		Name:          participant.Name,
		BaekjoonID:    participant.BaekjoonID,
		Score:         score,
		CurrentTier:   userInfo.Tier,
		CurrentRating: userInfo.Rating,
		ProblemCount:  newProblemCount,
	}, nil
}

// sortScores 점수 데이터를 정렬합니다
func (sm *ScoreboardManager) sortScores(scores []models.ScoreData) {
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
}

func (sm *ScoreboardManager) formatScoreboard(competition *models.Competition, scores []models.ScoreData, isAdmin bool) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏆 %s 스코어보드", competition.Name),
		Description: fmt.Sprintf("%s ~ %s",
			competition.StartDate.Format(constants.DateFormat),
			competition.EndDate.Format(constants.DateFormat)),
		Color: constants.ColorTierGold,
	}

	if len(scores) == 0 {
		embed.Description += "\n\n아직 점수가 계산된 참가자가 없습니다."
		return embed
	}

	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%-4s %-*s %6s\n",
		"순위", constants.MaxUsernameLength, "이름", "점수"))
	sb.WriteString("──────────────────────────────\n")

	for i, score := range scores {
		rank := i + 1
		sb.WriteString(fmt.Sprintf("%-4d %-*s %6.0f\n",
			rank,
			constants.MaxUsernameLength,
			utils.TruncateString(score.Name, constants.MaxUsernameLength),
			score.Score))
	}

	sb.WriteString("```")

	embed.Description += "\n\n" + sb.String()

	// 블랙아웃 경고 추가
	now := time.Now()
	if now.Before(competition.BlackoutStartDate) {
		daysLeft := int(competition.BlackoutStartDate.Sub(now).Hours() / 24)
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("⚠️ %d일 후 스코어보드가 비공개됩니다", daysLeft),
		}
	}

	return embed
}

func (sm *ScoreboardManager) SendDailyScoreboard(session *discordgo.Session, channelID string) error {
	embed, err := sm.GenerateScoreboard(false) // 자동 스코어보드는 관리자 권한 없음
	if err != nil {
		return err
	}

	_, err = session.ChannelMessageSendEmbed(channelID, embed)
	return err
}
