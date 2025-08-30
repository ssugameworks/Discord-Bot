package interfaces

// ScoreCalculator 점수 계산을 위한 인터페이스입니다
type ScoreCalculator interface {
	CalculateScore(handle string, startTier int, startProblemIDs []int) (float64, error)
}
