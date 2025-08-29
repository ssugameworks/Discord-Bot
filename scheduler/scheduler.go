package scheduler

import (
	"discord-bot/bot"
	"discord-bot/config"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Scheduler struct {
	session           *discordgo.Session
	config            *config.Config
	scoreboardManager *bot.ScoreboardManager
	ticker            *time.Ticker
	stopChan          chan bool
}

func NewScheduler(session *discordgo.Session, config *config.Config, scoreboardManager *bot.ScoreboardManager) *Scheduler {
	return &Scheduler{
		session:           session,
		config:            config,
		scoreboardManager: scoreboardManager,
		stopChan:          make(chan bool),
	}
}

func (s *Scheduler) StartDailyScoreboard() {
	s.ticker = time.NewTicker(24 * time.Hour)

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

	log.Println("일일 스코어보드 스케줄러가 시작되었습니다.")
}

func (s *Scheduler) StartCustomSchedule(hour, minute int) {
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	if nextRun.Before(now) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	duration := nextRun.Sub(now)

	go func() {
		time.Sleep(duration)
		s.sendDailyScoreboard()

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.sendDailyScoreboard()
			case <-s.stopChan:
				return
			}
		}
	}()

	log.Printf("일일 스코어보드 스케줄러가 매일 %02d:%02d에 실행되도록 설정되었습니다.", hour, minute)
}

func (s *Scheduler) sendDailyScoreboard() {
	if s.config.Discord.ChannelID == "" {
		log.Println("DISCORD_CHANNEL_ID가 설정되지 않았습니다.")
		return
	}

	err := s.scoreboardManager.SendDailyScoreboard(s.session, s.config.Discord.ChannelID)
	if err != nil {
		log.Printf("일일 스코어보드 전송 실패: %v", err)
		return
	}

	log.Println("일일 스코어보드를 성공적으로 전송했습니다.")
}

func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	select {
	case s.stopChan <- true:
	default:
		// channel is full or no receiver, skip
	}

	log.Println("스케줄러가 중지되었습니다.")
}
