package killmail

import (
	"github.com/ddouglas/neo"
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
