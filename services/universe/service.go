package universe

import "github.com/ddouglas/neo"

type Service interface {
	killboard.UniverseRepository
}

type service struct {
	killboard.UniverseRepository
}

func NewService(killmail killboard.UniverseRepository) Service {
	return &service{
		killmail,
	}
}
