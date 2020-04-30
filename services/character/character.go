package character

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

var rcharacter = "character:%d"

func (s *service) Character(ctx context.Context, id uint64) (*neo.Character, error) {
	var character = new(neo.Character)
	var key = fmt.Sprintf(rcharacter, id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {

		err = json.Unmarshal(result, character)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal character from redis")
		}
		return character, nil
	}

	character, err = s.CharacterRespository.Character(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "unable to query database for character")
	}

	if err == nil {
		bSlice, err := json.Marshal(character)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal character for cache")
		}

		_, err = s.redis.Set(key, bSlice, time.Minute*60).Result()

		return character, errors.Wrap(err, "failed to cache character in redis")
	}

	// Character is not cached, the DB doesn't have this character, lets check ESI
	character, m := s.esi.GetCharactersCharacterID(id, null.NewString("", false))
	if m.IsError() {
		return nil, m.Msg
	}

	// ESI has the character. Lets insert it into the db, and cache it is redis
	_, err = s.CharacterRespository.CreateCharacter(ctx, character)
	if err != nil {
		return character, errors.Wrap(err, "unable to insert character into db")
	}

	byteSlice, err := json.Marshal(character)
	if err != nil {
		return character, errors.Wrap(err, "unable to marshal character for cache")
	}

	_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()

	return character, errors.Wrap(err, "failed to cache solar character in redis")
}

func (s *service) AlliancesByAllianceIDs(ctx context.Context, ids []uint64) ([]*neo.Character, error) {

	var characters = make([]*neo.Character, 0)
	for _, id := range ids {
		key := fmt.Sprintf(rcharacter, id)
		result, err := s.redis.Get(key).Bytes()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			return nil, errors.Wrap(err, "encountered error querying redis")
		}

		if len(result) > 0 {

			var character = new(neo.Character)
			err = json.Unmarshal(result, character)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal character bytes into struct")
			}

			characters = append(characters, character)

		}
	}

	if len(ids) == len(characters) {
		return characters, nil
	}

	var missing []uint64
	for _, id := range ids {
		found := false
		for _, character := range characters {
			if character.ID == id {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return characters, nil
	}

	dbTypes, err := s.CharacterRespository.CharactersByCharacterIDs(ctx, missing)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, character := range dbTypes {
		key := fmt.Sprintf(rcharacter, character.ID)

		byteSlice, err := json.Marshal(character)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal character to slice of bytes")
		}

		_, err = s.redis.Set(key, byteSlice, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache character in redis")
		}

		characters = append(characters, character)
	}

	return characters, nil

}