package character

import "github.com/eveisesi/neo"

type Service interface {
	neo.CharacterRespository
}

type service struct {
	neo.CharacterRespository
}

func NewService(character neo.CharacterRespository) Service {
	return &service{
		character,
	}
}
