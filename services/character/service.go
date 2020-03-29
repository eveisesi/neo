package character

import "github.com/eveisesi/neo"

type Service interface {
	killboard.CharacterRespository
}

type service struct {
	killboard.CharacterRespository
}

func NewService(character killboard.CharacterRespository) Service {
	return &service{
		character,
	}
}
