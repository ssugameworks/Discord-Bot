package scoring

import (
	"discord-bot/api"
	"discord-bot/constants"
	"discord-bot/models"
	"math"
)

type ScoreCalculator struct {
	client      *api.SolvedACClient
	tierManager *models.TierManager
}

func NewScoreCalculator() *ScoreCalculator {
	return &ScoreCalculator{
		client:      api.NewSolvedACClient(),
		tierManager: models.NewTierManager(),
	}
}

func (sc *ScoreCalculator) CalculateScore(handle string, startTier int, startProblemIDs []int) (float64, error) {
	top100, err := sc.client.GetUserTop100(handle)
	if err != nil {
		return 0, err
	}

	// 시작 시점 문제 ID들을 맵으로 변환
	startProblemsMap := make(map[int]bool)
	for _, id := range startProblemIDs {
		startProblemsMap[id] = true
	}

	totalScore := 0.0

	for _, problem := range top100.Items {
		// 참가 시점에 이미 해결한 문제는 제외
		if startProblemsMap[problem.ProblemID] {
			continue
		}

		problemTier := problem.Level
		points := sc.tierManager.GetTierPoints(problemTier)
		if points == 0 {
			continue
		}

		weight := sc.getWeight(problemTier, startTier)
		score := float64(points) * weight
		totalScore += score
	}

	return math.Round(totalScore), nil
}

func (sc *ScoreCalculator) getWeight(problemTier, startTier int) float64 {
	if problemTier > startTier {
		return constants.ChallengeMultiplier
	} else if problemTier == startTier {
		return constants.BaseMultiplier
	} else {
		return constants.PenaltyMultiplier
	}
}

// GetTierName returns the display name for a tier using the global tier manager
func GetTierName(tier int) string {
	tm := models.NewTierManager()
	return tm.GetTierName(tier)
}