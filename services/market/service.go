package market

import "github.com/eveisesi/neo"

type Service interface {
	neo.MarketRepository
}

type service struct {
	neo.MarketRepository
}

func NewService(market neo.MarketRepository) Service {
	return &service{
		market,
	}
}
