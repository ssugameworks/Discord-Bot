package utils

import (
	"github.com/bwmarrin/discordgo"
)

// CommandContext 명령어 처리를 위한 컨텍스트 정보를 담고 있습니다
type CommandContext struct {
	Session     *discordgo.Session
	Message     *discordgo.MessageCreate
	ErrorHelper *ErrorHandlerFactory
}

// NewCommandContext 새로운 CommandContext를 생성합니다
func NewCommandContext(s *discordgo.Session, m *discordgo.MessageCreate) *CommandContext {
	return &CommandContext{
		Session:     s,
		Message:     m,
		ErrorHelper: NewErrorHandlerFactory(s, m.ChannelID),
	}
}

// ChannelID 채널 ID를 반환합니다
func (ctx *CommandContext) ChannelID() string {
	return ctx.Message.ChannelID
}

// AuthorID 메시지 작성자 ID를 반환합니다
func (ctx *CommandContext) AuthorID() string {
	return ctx.Message.Author.ID
}

// GuildID 길드 ID를 반환합니다
func (ctx *CommandContext) GuildID() string {
	return ctx.Message.GuildID
}

// IsDM DM인지 확인합니다
func (ctx *CommandContext) IsDM() bool {
	return ctx.Message.GuildID == ""
}

// SendMessage 채널에 메시지를 전송합니다
func (ctx *CommandContext) SendMessage(content string) error {
	_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID(), content)
	return err
}

// SendEmbed 채널에 embed 메시지를 전송합니다
func (ctx *CommandContext) SendEmbed(embed *discordgo.MessageEmbed) error {
	_, err := ctx.Session.ChannelMessageSendEmbed(ctx.ChannelID(), embed)
	return err
}
