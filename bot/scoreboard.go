package bot

import (
	"discord-bot/api"
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

func (sm *ScoreboardManager) GenerateScoreboard(isAdmin bool) (string, error) {
	competition := sm.storage.GetCompetition()
	if competition == nil || !competition.IsActive {
		return "", fmt.Errorf("no active competition")
	}

	if sm.storage.IsBlackoutPeriod() && competition.ShowScoreboard && !isAdmin {
		return "🔒 **스코어보드 비공개 기간입니다** (마지막 3일)", nil
	}

	participants := sm.storage.GetParticipants()
	if len(participants) == 0 {
		return "참가자가 없습니다.", nil
	}

	scores := make([]models.ScoreData, 0, len(participants))

	for _, participant := range participants {
		userInfo, err := sm.client.GetUserInfo(participant.BaekjoonID)
		if err != nil {
			continue
		}

		score, err := sm.calculator.CalculateScore(participant.BaekjoonID, participant.StartTier, participant.StartProblemIDs)
		if err != nil {
			continue
		}

		top100, err := sm.client.GetUserTop100(participant.BaekjoonID)
		if err != nil {
			continue
		}

		// 새로 푼 문제 수 계산 (현재 - 시작시점)
		newProblemCount := top100.Count - participant.StartProblemCount
		if newProblemCount < 0 {
			newProblemCount = 0
		}

		scoreData := models.ScoreData{
			ParticipantID: participant.ID,
			Name:          participant.Name,
			BaekjoonID:    participant.BaekjoonID,
			Score:         score,
			CurrentTier:   userInfo.Tier,
			CurrentRating: userInfo.Rating,
			ProblemCount:  newProblemCount,
		}

		scores = append(scores, scoreData)
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return sm.formatScoreboard(competition, scores, isAdmin), nil
}

func (sm *ScoreboardManager) formatScoreboard(competition *models.Competition, scores []models.ScoreData, isAdmin bool) string {
	var sb strings.Builder

	sb.WriteString("🏆 **알고리즘 경진대회 스코어보드**\n")
	sb.WriteString(fmt.Sprintf("📅 **%s**\n", competition.Name))
	sb.WriteString(fmt.Sprintf("⏰ %s ~ %s\n\n",
		competition.StartDate.Format("2006-01-02"),
		competition.EndDate.Format("2006-01-02")))

	if len(scores) == 0 {
		sb.WriteString("아직 점수가 계산된 참가자가 없습니다.\n")
		return sb.String()
	}

	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%-4s %-15s %6s\n",
		"순위", "이름", "점수"))
	sb.WriteString("──────────────────────────────\n")

	for i, score := range scores {
		rank := i + 1
		sb.WriteString(fmt.Sprintf("%-4d %-15s %6.0f\n",
			rank,
			utils.TruncateString(score.Name, 15),
			score.Score))
	}

	sb.WriteString("```\n")

	now := time.Now()
	if now.Before(competition.BlackoutStartDate) {
		daysLeft := int(competition.BlackoutStartDate.Sub(now).Hours() / 24)
		sb.WriteString(fmt.Sprintf("\n⚠️ **%d일 후 스코어보드가 비공개됩니다**", daysLeft))
	}

	return sb.String()
}

func (sm *ScoreboardManager) SendDailyScoreboard(session *discordgo.Session, channelID string) error {
	scoreboard, err := sm.GenerateScoreboard(false) // 자동 스코어보드는 관리자 권한 없음
	if err != nil {
		return err
	}

	_, err = session.ChannelMessageSend(channelID, scoreboard)
	return err
}

