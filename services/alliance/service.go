package alliance

import "github.com/ddouglas/killboard"

type Service interface {
	killboard.AllianceRespository
}

type service struct {
	killboard.AllianceRespository
}

func NewService(alliance killboard.AllianceRespository) Service {
	return &service{
		alliance,
	}
}
