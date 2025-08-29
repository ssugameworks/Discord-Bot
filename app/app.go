package app

import (
	"discord-bot/bot"
	"discord-bot/config"
	"discord-bot/scheduler"
	"discord-bot/storage"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type Application struct {
	config            *config.Config
	session           *discordgo.Session
	storage           *storage.Storage
	commandHandler    *bot.CommandHandler
	scoreboardManager *bot.ScoreboardManager
	scheduler         *scheduler.Scheduler
}

func New() (*Application, error) {
	app := &Application{}

	if err := app.loadConfig(); err != nil {
		return nil, err
	}

	if err := app.initializeStorage(); err != nil {
		return nil, err
	}

	if err := app.initializeDiscord(); err != nil {
		return nil, err
	}

	app.setupHandlers()
	app.initializeScheduler()

	return app, nil
}

func (app *Application) loadConfig() error {
	app.config = config.Load()
	if err := app.config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}

func (app *Application) initializeStorage() error {
	app.storage = storage.NewStorage()
	return nil
}

func (app *Application) initializeDiscord() error {
	session, err := discordgo.New("Bot " + app.config.Discord.Token)
	if err != nil {
		return fmt.Errorf("디스코드 세션 생성 실패: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuilds | discordgo.IntentsDirectMessages
	app.session = session
	return nil
}

func (app *Application) setupHandlers() {
	app.commandHandler = bot.NewCommandHandler(app.storage)
	app.scoreboardManager = bot.NewScoreboardManager(app.storage)

	app.session.AddHandler(app.commandHandler.HandleMessage)
	app.session.AddHandler(app.handleReady)
}

func (app *Application) initializeScheduler() {
	app.scheduler = scheduler.NewScheduler(app.session, app.config, app.scoreboardManager)
}

func (app *Application) Start() error {
	if err := app.session.Open(); err != nil {
		return fmt.Errorf("웹소켓 연결 실패: %w", err)
	}

	if app.config.Schedule.Enabled {
		app.scheduler.StartCustomSchedule(
			app.config.Schedule.ScoreboardHour,
			app.config.Schedule.ScoreboardMinute,
		)
		log.Printf("매일 %02d:%02d에 자동으로 스코어보드가 띄워집니다.",
			app.config.Schedule.ScoreboardHour, app.config.Schedule.ScoreboardMinute)
	} else {
		log.Println("DISCORD_CHANNEL_ID가 설정되지 않았습니다. 스코어보드가 비활성화되었습니다.")
	}

	app.printStartupMessage()
	return nil
}

func (app *Application) printStartupMessage() {
	fmt.Println("디스코드 봇이 실행되었습니다!")
	fmt.Println("📋 사용 가능한 명령어: !help")
	if app.config.Schedule.Enabled {
		fmt.Printf("⏰ 매일 %02d:%02d에 자동으로 스코어보드가 전송됩니다.\n",
			app.config.Schedule.ScoreboardHour, app.config.Schedule.ScoreboardMinute)
	}
}

func (app *Application) Run() error {
	if err := app.Start(); err != nil {
		return err
	}

	// 종료 신호 대기
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGKILL)
	<-sc

	return app.Stop()
}

func (app *Application) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	// TODO: Welcome message
}

func (app *Application) Stop() error {
	fmt.Println("🔄 봇을 종료하는 중...")

	if app.scheduler != nil {
		app.scheduler.Stop()
	}

	if app.session != nil {
		app.session.Close()
	}

	fmt.Println("봇이 정상적으로 종료되었습니다.")
	return nil
}
