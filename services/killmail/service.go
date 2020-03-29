package killmail

import (
	"github.com/eveisesi/neo"
)

type Service interface {
	neo.KillmailRespository
}

type service struct {
	neo.KillmailRespository
}

func NewService(killmail neo.KillmailRespository) Service {
	return &service{
		killmail,
	}
}
