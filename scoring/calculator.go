package scoring

import (
	"discord-bot/api"
	"discord-bot/constants"
	"math"
)


var tierPoints = map[int]int{
	1: 1,    // Bronze V
	2: 2,    // Bronze IV
	3: 3,    // Bronze III
	4: 4,    // Bronze II
	5: 5,    // Bronze I
	6: 8,    // Silver V
	7: 10,   // Silver IV
	8: 12,   // Silver III
	9: 14,   // Silver II
	10: 16,  // Silver I
	11: 18,  // Gold V
	12: 20,  // Gold IV
	13: 22,  // Gold III
	14: 23,  // Gold II
	15: 25,  // Gold I
	16: 28,  // Platinum V
	17: 30,  // Platinum IV
	18: 32,  // Platinum III
	19: 35,  // Platinum II
	20: 37,  // Platinum I
	21: 40,  // Diamond V
	22: 42,  // Diamond IV
	23: 45,  // Diamond III
	24: 47,  // Diamond II
	25: 50,  // Diamond I
	26: 55,  // Ruby V
	27: 60,  // Ruby IV
	28: 65,  // Ruby III
	29: 70,  // Ruby II
	30: 75,  // Ruby I
}

type ScoreCalculator struct {
	client *api.SolvedACClient
}

func NewScoreCalculator() *ScoreCalculator {
	return &ScoreCalculator{
		client: api.NewSolvedACClient(),
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
		points, exists := tierPoints[problemTier]
		if !exists {
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

func GetTierName(tier int) string {
	tierNames := map[int]string{
		0: "Unranked",
		1: "Bronze V", 2: "Bronze IV", 3: "Bronze III", 4: "Bronze II", 5: "Bronze I",
		6: "Silver V", 7: "Silver IV", 8: "Silver III", 9: "Silver II", 10: "Silver I",
		11: "Gold V", 12: "Gold IV", 13: "Gold III", 14: "Gold II", 15: "Gold I",
		16: "Platinum V", 17: "Platinum IV", 18: "Platinum III", 19: "Platinum II", 20: "Platinum I",
		21: "Diamond V", 22: "Diamond IV", 23: "Diamond III", 24: "Diamond II", 25: "Diamond I",
		26: "Ruby V", 27: "Ruby IV", 28: "Ruby III", 29: "Ruby II", 30: "Ruby I",
		31: "Master",
	}
	
	if name, exists := tierNames[tier]; exists {
		return name
	}
	return "Unknown"
}