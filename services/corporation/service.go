package corporation

import "github.com/eveisesi/neo"

type Service interface {
	neo.CorporationRespository
}

type service struct {
	neo.CorporationRespository
}

func NewService(corporation neo.CorporationRespository) Service {
	return &service{
		corporation,
	}
}
