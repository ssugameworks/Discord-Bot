package interfaces

import "discord-bot/api"

// APIClient 외부 API와의 통신을 위한 인터페이스입니다
type APIClient interface {
	GetUserInfo(handle string) (*api.UserInfo, error)
	GetUserTop100(handle string) (*api.Top100Response, error)
}
