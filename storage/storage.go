package storage

import (
	"discord-bot/api"
	"discord-bot/constants"
	"discord-bot/models"
	"discord-bot/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Storage는 참가자와 대회 데이터를 관리하는 저장소입니다
type Storage struct {
	participants []models.Participant
	competition  *models.Competition
	apiClient    *api.SolvedACClient
}

// NewStorage는 새로운 Storage 인스턴스를 생성하고 데이터를 로드합니다
func NewStorage() *Storage {
	utils.Info("Initializing storage system")
	s := &Storage{
		apiClient: api.NewSolvedACClient(),
	}
	s.loadData()
	utils.Info("Storage system initialized successfully")
	return s
}

func (s *Storage) loadData() {
	s.loadParticipants()
	s.loadCompetition()
}

// loadParticipants는 참가자 데이터를 파일에서 로드합니다
func (s *Storage) loadParticipants() {
	utils.Debug("Loading participants from file: %s", constants.ParticipantsFileName)
	file, err := os.Open(constants.ParticipantsFileName)
	if err != nil {
		if os.IsNotExist(err) {
			utils.Warn("Participants file not found, starting with empty list")
			s.participants = []models.Participant{}
			return
		}
		utils.Error("Failed to open participants file: %v", err)
		// 파일이 존재하지만 열 수 없는 경우 빈 슬라이스로 초기화하지 않음
		s.participants = []models.Participant{}
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		utils.Error("Failed to read participants file: %v", err)
		// 읽기 실패 시에도 기존 데이터 보존을 위해 빈 슬라이스로 초기화하지 않음
		s.participants = []models.Participant{}
		return
	}

	// 빈 파일 처리
	if len(data) == 0 {
		utils.Info("Empty participants file, starting with empty list")
		s.participants = []models.Participant{}
		return
	}

	if err := json.Unmarshal(data, &s.participants); err != nil {
		utils.Error("Failed to parse participants data: %v", err)
		// JSON 파싱 실패 시 백업 파일 생성
		backupFile := constants.ParticipantsFileName + ".corrupted"
		os.WriteFile(backupFile, data, constants.FilePermission)
		utils.Warn("Corrupted participants file backed up as %s", backupFile)
		s.participants = []models.Participant{}
		return
	}

	utils.Info("Loaded %d participants", len(s.participants))
}

// loadCompetition는 대회 데이터를 파일에서 로드합니다
func (s *Storage) loadCompetition() {
	utils.Debug("Loading competition from file: %s", constants.CompetitionFileName)
	file, err := os.Open(constants.CompetitionFileName)
	if err != nil {
		if os.IsNotExist(err) {
			utils.Warn("Competition file not found")
		} else {
			utils.Error("Failed to open competition file: %v", err)
		}
		s.competition = nil
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		utils.Error("Failed to read competition file: %v", err)
		s.competition = nil
		return
	}

	// 빈 파일 처리
	if len(data) == 0 {
		utils.Info("Empty competition file")
		s.competition = nil
		return
	}

	if err := json.Unmarshal(data, &s.competition); err != nil {
		utils.Error("Failed to parse competition data: %v", err)
		// JSON 파싱 실패 시 백업 파일 생성
		backupFile := constants.CompetitionFileName + ".corrupted"
		os.WriteFile(backupFile, data, constants.FilePermission)
		utils.Warn("Corrupted competition file backed up as %s", backupFile)
		s.competition = nil
		return
	}

	utils.Info("Loaded competition: %s", s.competition.Name)
}

// SaveParticipants는 참가자 데이터를 파일에 저장합니다
func (s *Storage) SaveParticipants() error {
	utils.Debug("Saving participants to file: %s", constants.ParticipantsFileName)
	data, err := json.MarshalIndent(s.participants, "", "  ")
	if err != nil {
		utils.Error("Failed to marshal participants data: %v", err)
		return err
	}

	err = os.WriteFile(constants.ParticipantsFileName, data, constants.FilePermission)
	if err != nil {
		utils.Error("Failed to save participants file: %v", err)
		return err
	}

	utils.Info("Successfully saved %d participants", len(s.participants))
	return nil
}

// SaveCompetition는 대회 데이터를 파일에 저장합니다
func (s *Storage) SaveCompetition() error {
	if s.competition == nil {
		utils.Debug("No competition to save")
		return nil
	}

	utils.Debug("Saving competition to file: %s", constants.CompetitionFileName)
	data, err := json.MarshalIndent(s.competition, "", "  ")
	if err != nil {
		utils.Error("Failed to marshal competition data: %v", err)
		return err
	}

	err = os.WriteFile(constants.CompetitionFileName, data, constants.FilePermission)
	if err != nil {
		utils.Error("Failed to save competition file: %v", err)
		return err
	}

	utils.Info("Successfully saved competition: %s", s.competition.Name)
	return nil
}

// AddParticipant는 새로운 참가자를 추가합니다
func (s *Storage) AddParticipant(name, baekjoonID string, startTier, startRating int) error {
	// 입력값 검증
	if !utils.IsValidUsername(name) {
		return fmt.Errorf("invalid username: %s", name)
	}
	if !utils.IsValidBaekjoonID(baekjoonID) {
		return fmt.Errorf("invalid baekjoon ID: %s", baekjoonID)
	}

	// 중복 확인
	for _, p := range s.participants {
		if p.BaekjoonID == baekjoonID {
			utils.Warn("Attempt to add duplicate participant: %s", baekjoonID)
			return fmt.Errorf("participant with Baekjoon ID %s already exists", baekjoonID)
		}
	}

	// 참가 시점의 해결한 문제들 가져오기
	startProblemIDs := []int{}
	startProblemCount := 0
	
	top100, err := s.apiClient.GetUserTop100(baekjoonID)
	if err == nil {
		for _, problem := range top100.Items {
			startProblemIDs = append(startProblemIDs, problem.ProblemID)
		}
		startProblemCount = len(startProblemIDs)
		utils.Info("Loaded %d starting problems for participant %s", startProblemCount, baekjoonID)
	} else {
		utils.Warn("Failed to load starting problems for participant %s: %v", baekjoonID, err)
	}

	participant := models.Participant{
		ID:                len(s.participants) + 1,
		Name:              utils.SanitizeString(name),
		BaekjoonID:        baekjoonID,
		StartTier:         startTier,
		StartRating:       startRating,
		CreatedAt:         time.Now(),
		StartProblemIDs:   startProblemIDs,
		StartProblemCount: startProblemCount,
	}

	s.participants = append(s.participants, participant)
	utils.Info("Added new participant: %s (%s)", name, baekjoonID)
	return s.SaveParticipants()
}

func (s *Storage) GetParticipants() []models.Participant {
	return s.participants
}

func (s *Storage) CreateCompetition(name string, startDate, endDate time.Time) error {
	blackoutStart := endDate.AddDate(0, 0, -constants.BlackoutDays)
	
	s.competition = &models.Competition{
		ID:                1,
		Name:              name,
		StartDate:         startDate,
		EndDate:           endDate,
		BlackoutStartDate: blackoutStart,
		IsActive:          true,
		ShowScoreboard:    true,
		Participants:      s.participants,
	}

	return s.SaveCompetition()
}

func (s *Storage) GetCompetition() *models.Competition {
	return s.competition
}

func (s *Storage) SetScoreboardVisibility(visible bool) error {
	if s.competition == nil {
		return fmt.Errorf("no active competition")
	}

	s.competition.ShowScoreboard = visible
	return s.SaveCompetition()
}

func (s *Storage) IsBlackoutPeriod() bool {
	if s.competition == nil {
		return false
	}

	now := time.Now()
	return now.After(s.competition.BlackoutStartDate) && now.Before(s.competition.EndDate)
}

// UpdateCompetitionName은 대회명을 업데이트합니다
func (s *Storage) UpdateCompetitionName(name string) error {
	if s.competition == nil {
		return fmt.Errorf("no active competition")
	}

	s.competition.Name = name
	return s.SaveCompetition()
}

// UpdateCompetitionStartDate는 대회 시작일을 업데이트합니다
func (s *Storage) UpdateCompetitionStartDate(startDate time.Time) error {
	if s.competition == nil {
		return fmt.Errorf("no active competition")
	}

	s.competition.StartDate = startDate
	return s.SaveCompetition()
}

// UpdateCompetitionEndDate는 대회 종료일을 업데이트하고 블랙아웃 기간도 자동으로 재설정합니다
func (s *Storage) UpdateCompetitionEndDate(endDate time.Time) error {
	if s.competition == nil {
		return fmt.Errorf("no active competition")
	}

	s.competition.EndDate = endDate
	// 블랙아웃 기간도 자동으로 재설정 (종료일 3일 전부터)
	s.competition.BlackoutStartDate = endDate.AddDate(0, 0, -constants.BlackoutDays)
	return s.SaveCompetition()
}