package character

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/eveisesi/neo"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/pkg/errors"
)

func (s *service) Character(ctx context.Context, id uint64) (*neo.Character, error) {
	var character = new(neo.Character)
	var key = fmt.Sprintf(neo.REDIS_CHARACTER, id)

	result, err := s.redis.WithContext(ctx).Get(key).Bytes()
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
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "unable to query database for character")
	}

	if err == nil {
		bSlice, err := json.Marshal(character)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal character for cache")
		}

		_, err = s.redis.WithContext(ctx).Set(key, bSlice, time.Minute*60).Result()

		return character, errors.Wrap(err, "failed to cache character in redis")
	}

	// Character is not cached, the DB doesn't have this character, lets check ESI
	character, m := s.esi.GetCharactersCharacterID(ctx, id, "")
	if m.IsErr() {
		return nil, m.Msg
	}

	// ESI has the character. Lets insert it into the db, and cache it is redis
	err = s.CharacterRespository.CreateCharacter(ctx, character)
	if err != nil {
		return character, errors.Wrap(err, "unable to insert character into db")
	}

	byteSlice, err := json.Marshal(character)
	if err != nil {
		return character, errors.Wrap(err, "unable to marshal character for cache")
	}

	_, err = s.redis.WithContext(ctx).Set(key, byteSlice, time.Minute*60).Result()

	return character, errors.Wrap(err, "failed to cache character in redis")
}

func (s *service) CharactersByCharacterIDs(ctx context.Context, ids []uint64) ([]*neo.Character, error) {

	var characters = make([]*neo.Character, 0)
	for _, id := range ids {
		key := fmt.Sprintf(neo.REDIS_CHARACTER, id)
		result, err := s.redis.WithContext(ctx).Get(key).Bytes()
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

	dbTypes, err := s.Characters(ctx, neo.InUint64{Column: "id", Value: missing})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query db for missing type ids")
	}

	for _, character := range dbTypes {
		key := fmt.Sprintf(neo.REDIS_CHARACTER, character.ID)

		byteSlice, err := json.Marshal(character)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal character to slice of bytes")
		}

		_, err = s.redis.WithContext(ctx).Set(key, byteSlice, time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "unable to cache character in redis")
		}

		characters = append(characters, character)
	}

	return characters, nil

}

func (s *service) CreateCharacter(ctx context.Context, character *neo.Character) error {
	return s.CharacterRespository.CreateCharacter(ctx, character)
}

func (s *service) UpdateExpired(ctx context.Context) {

	for {
		txn := s.newrelic.StartTransaction("updateExpiredCharacters")
		ctx = newrelic.NewContext(ctx, txn)

		expired, err := s.Expired(ctx)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			s.logger.WithError(err).Error("Failed to fetch expired characters")
			return
		}

		if len(expired) == 0 {
			s.logger.Info("no expired characters found")
			time.Sleep(time.Minute * 5)
			continue
		}

		for _, character := range expired {
			entry := s.logger.WithContext(ctx).WithField("character_id", character.ID)
			seg := txn.StartSegment("handleCharacter")
			seg.AddAttribute("id", character.ID)
			s.tracker.GateKeeper(ctx)
			newCharacter, m := s.esi.GetCharactersCharacterID(ctx, character.ID, character.Etag)
			if m.IsErr() {
				txn.NoticeError(m.Msg)
				entry.WithError(err).Error("failed to fetch character from esi")
				continue
			}

			switch m.Code {
			case http.StatusNotModified:

				character.NotModifiedCount++

				if character.NotModifiedCount >= 2 && character.UpdatePriority <= 3 {
					character.NotModifiedCount = 0
					character.UpdatePriority++
				}

				character.CachedUntil = time.Unix(newCharacter.CachedUntil, 0).AddDate(0, 0, int(character.UpdatePriority)).Unix()
				character.Etag = newCharacter.Etag

				err = s.UpdateCharacter(ctx, character.ID, character)
			case http.StatusOK:
				err = s.UpdateCharacter(ctx, character.ID, newCharacter)
			default:
				entry.WithField("status_code", m.Code).Error("unaccounted for status code received from esi service")
			}
			if err != nil {
				txn.NoticeError(err)
				entry.WithError(err).Error("failed to update character")
				continue
			}

			entry.WithField("status_code", m.Code).Debug("character successfully updated")
			seg.End()
			time.Sleep(time.Millisecond * 50)
		}
		s.logger.WithContext(ctx).WithField("count", len(expired)).Info("characters successfully updated")
		txn.End()
		time.Sleep(time.Second)

	}

}
