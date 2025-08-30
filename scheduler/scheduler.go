package scheduler

import (
	"discord-bot/bot"
	"discord-bot/config"
	"discord-bot/constants"
	"discord-bot/utils"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Scheduler struct {
	session           *discordgo.Session
	config            *config.Config
	scoreboardManager *bot.ScoreboardManager
	ticker            *time.Ticker
	customTicker      *time.Ticker
	stopChan          chan bool
	customStopChan    chan bool
}

func NewScheduler(session *discordgo.Session, config *config.Config, scoreboardManager *bot.ScoreboardManager) *Scheduler {
	return &Scheduler{
		session:           session,
		config:            config,
		scoreboardManager: scoreboardManager,
		stopChan:          make(chan bool),
		customStopChan:    make(chan bool),
	}
}

func (s *Scheduler) StartDailyScoreboard() {
	s.ticker = time.NewTicker(constants.SchedulerInterval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.sendDailyScoreboard()
			case <-s.stopChan:
				return
			}
		}
	}()

	utils.Info("일일 스코어보드 스케줄러가 시작되었습니다")
}

func (s *Scheduler) StartCustomSchedule(hour, minute int) {
	// 기존 커스텀 스케줄러가 있다면 정리
	s.stopCustomScheduler()

	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	if nextRun.Before(now) {
		nextRun = nextRun.Add(constants.SchedulerInterval)
	}

	duration := nextRun.Sub(now)

	go func() {
		// 첫 실행까지 대기, 중단 신호 체크
		select {
		case <-time.After(duration):
			s.sendDailyScoreboard()
		case <-s.customStopChan:
			return
		}

		// 정기적 실행 시작
		s.customTicker = time.NewTicker(constants.SchedulerInterval)
		defer s.customTicker.Stop()

		for {
			select {
			case <-s.customTicker.C:
				s.sendDailyScoreboard()
			case <-s.customStopChan:
				return
			}
		}
	}()

	utils.Info("일일 스코어보드 스케줄러가 매일 %02d:%02d에 실행되도록 설정되었습니다", hour, minute)
}

func (s *Scheduler) sendDailyScoreboard() {
	if s.config.Discord.ChannelID == "" {
		utils.Error("채널 ID가 설정되지 않아 스코어보드를 전송할 수 없습니다")
		return
	}

	err := s.scoreboardManager.SendDailyScoreboard(s.session, s.config.Discord.ChannelID)
	if err != nil {
		utils.Error("일일 스코어보드 전송 실패: %v", err)
		return
	}

	utils.Info("일일 스코어보드를 성공적으로 전송했습니다")
}

func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	s.stopCustomScheduler()

	select {
	case s.stopChan <- true:
	default:
		// channel is full or no receiver, skip
	}

	utils.Info("스케줄러가 중지되었습니다")
}

func (s *Scheduler) stopCustomScheduler() {
	if s.customTicker != nil {
		s.customTicker.Stop()
		s.customTicker = nil
	}

	select {
	case s.customStopChan <- true:
	default:
		// channel is full or no receiver, skip
	}
}
