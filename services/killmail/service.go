package killmail

import (
	"github.com/ddouglas/killboard"
)

type Service interface {
	killboard.KillmailRespository
}

type service struct {
	killboard.KillmailRespository
}

func NewService(killmail killboard.KillmailRespository) Service {
	return &service{
		killmail,
	}
}
