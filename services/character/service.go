package character

import "github.com/ddouglas/neo"

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
