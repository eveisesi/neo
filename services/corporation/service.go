package corporation

import "github.com/eveisesi/neo"

type Service interface {
	killboard.CorporationRespository
}

type service struct {
	killboard.CorporationRespository
}

func NewService(corporation killboard.CorporationRespository) Service {
	return &service{
		corporation,
	}
}
