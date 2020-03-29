package universe

import "github.com/eveisesi/neo"

type Service interface {
	neo.UniverseRepository
}

type service struct {
	neo.UniverseRepository
}

func NewService(killmail neo.UniverseRepository) Service {
	return &service{
		killmail,
	}
}
