package alliance

import "github.com/eveisesi/neo"

type Service interface {
	neo.AllianceRespository
}

type service struct {
	neo.AllianceRespository
}

func NewService(alliance neo.AllianceRespository) Service {
	return &service{
		alliance,
	}
}
