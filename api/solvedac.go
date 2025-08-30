package api

import (
	"discord-bot/constants"
	"discord-bot/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// solved.ac API와 통신하는 클라이언트입니다
type SolvedACClient struct {
	client  *http.Client
	baseURL string
}

// solved.ac 사용자 정보를 나타냅니다
type UserInfo struct {
	Handle          string `json:"handle"`
	Bio             string `json:"bio"`
	Rating          int    `json:"rating"`
	Tier            int    `json:"tier"`
	Class           int    `json:"class"`
	ClassDecoration string `json:"classDecoration"`
	ProfileImageURL string `json:"profileImageUrl"`
	SolvedCount     int    `json:"solvedCount"`
	Verified        bool   `json:"verified"`
	Rank            int    `json:"rank"`
}

// solved.ac 문제 정보를 나타냅니다
type ProblemInfo struct {
	ProblemID         int     `json:"problemId"`
	Level             int     `json:"level"`
	TitleKo           string  `json:"titleKo"`
	AcceptedUserCount int     `json:"acceptedUserCount"`
	AverageTries      float64 `json:"averageTries"`
}

// 사용자의 TOP 100 문제 응답을 나타냅니다
type Top100Response struct {
	Count int           `json:"count"`
	Items []ProblemInfo `json:"items"`
}

// 새로운 SolvedACClient 인스턴스를 생성합니다
func NewSolvedACClient() *SolvedACClient {
	utils.Debug("Creating new SolvedAC API client")
	return &SolvedACClient{
		client: &http.Client{
			Timeout: constants.APITimeout,
		},
		baseURL: constants.SolvedACBaseURL,
	}
}

// 지정된 핸들의 사용자 정보를 가져옵니다
func (c *SolvedACClient) GetUserInfo(handle string) (*UserInfo, error) {
	if !utils.IsValidBaekjoonID(handle) {
		return nil, fmt.Errorf("잘못된 핸들 형식: %s", handle)
	}

	url := fmt.Sprintf("%s/user/show?handle=%s", c.baseURL, handle)
	return c.getUserInfoWithRetry(url, handle)
}

// 재시도 로직을 포함한 사용자 정보 조회
func (c *SolvedACClient) getUserInfoWithRetry(url, handle string) (*UserInfo, error) {
	var lastErr error

	for attempt := 0; attempt < constants.MaxRetries; attempt++ {
		if attempt > 0 {
			utils.Debug("Retrying user info fetch for %s (attempt %d/%d)", handle, attempt+1, constants.MaxRetries)
			time.Sleep(constants.RetryDelay * time.Duration(attempt))
		}

		utils.Debug("Fetching user info from: %s", url)

		resp, err := c.client.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("사용자 정보 조회 실패: %w", err)
			utils.Warn("Attempt %d failed for user %s: %v", attempt+1, handle, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("요청 한도 초과")
			utils.Warn("Rate limited for user %s, attempt %d", handle, attempt+1)
			time.Sleep(constants.RetryDelay * 2)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API가 상태 코드 %d를 반환했습니다", resp.StatusCode)
			utils.Warn("API returned non-200 status for user %s: %d", handle, resp.StatusCode)
			if resp.StatusCode >= 500 {
				continue // 서버 에러는 재시도
			}
			break // 클라이언트 에러는 즉시 반환
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("응답 읽기 실패: %w", err)
			utils.Error("Failed to read response body for user %s: %v", handle, err)
			continue
		}

		var userInfo UserInfo
		if err := json.Unmarshal(body, &userInfo); err != nil {
			lastErr = fmt.Errorf("사용자 정보 파싱 실패: %w", err)
			utils.Error("Failed to parse user info for %s: %v", handle, err)
			continue
		}

		utils.Debug("Successfully fetched user info for %s (tier: %d, rating: %d)",
			handle, userInfo.Tier, userInfo.Rating)
		return &userInfo, nil
	}

	utils.Error("Failed to fetch user info for %s after %d attempts: %v", handle, constants.MaxRetries, lastErr)
	return nil, lastErr
}

// 지정된 사용자의 TOP 100 문제를 가져옵니다
func (c *SolvedACClient) GetUserTop100(handle string) (*Top100Response, error) {
	if !utils.IsValidBaekjoonID(handle) {
		return nil, fmt.Errorf("잘못된 핸들 형식: %s", handle)
	}

	url := fmt.Sprintf("%s/user/top_100?handle=%s", c.baseURL, handle)
	return c.getUserTop100WithRetry(url, handle)
}

// 재시도 로직을 포함한 TOP 100 조회
func (c *SolvedACClient) getUserTop100WithRetry(url, handle string) (*Top100Response, error) {
	var lastErr error

	for attempt := 0; attempt < constants.MaxRetries; attempt++ {
		if attempt > 0 {
			utils.Debug("Retrying top 100 fetch for %s (attempt %d/%d)", handle, attempt+1, constants.MaxRetries)
			time.Sleep(constants.RetryDelay * time.Duration(attempt))
		}

		utils.Debug("Fetching top 100 problems from: %s", url)

		resp, err := c.client.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("TOP 100 조회 실패: %w", err)
			utils.Warn("Attempt %d failed for top 100 %s: %v", attempt+1, handle, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("요청 한도 초과")
			utils.Warn("Rate limited for top 100 %s, attempt %d", handle, attempt+1)
			time.Sleep(constants.RetryDelay * 2)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API가 상태 코드 %d를 반환했습니다", resp.StatusCode)
			utils.Warn("API returned non-200 status for top 100 %s: %d", handle, resp.StatusCode)
			if resp.StatusCode >= 500 {
				continue // 서버 에러는 재시도
			}
			break // 클라이언트 에러는 즉시 반환
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("응답 읽기 실패: %w", err)
			utils.Error("Failed to read top 100 response body for %s: %v", handle, err)
			continue
		}

		var top100 Top100Response
		if err := json.Unmarshal(body, &top100); err != nil {
			lastErr = fmt.Errorf("TOP 100 파싱 실패: %w", err)
			utils.Error("Failed to parse top 100 for %s: %v", handle, err)
			continue
		}

		utils.Debug("Successfully fetched %d top problems for %s", top100.Count, handle)
		return &top100, nil
	}

	utils.Error("Failed to fetch top 100 for %s after %d attempts: %v", handle, constants.MaxRetries, lastErr)
	return nil, lastErr
}
