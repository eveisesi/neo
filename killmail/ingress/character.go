package ingress

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/killboard"
	"github.com/ddouglas/killboard/mysql/boiler"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

func (i *Ingresser) GetCharacterByID(id uint64) (*killboard.Character, error) {

	var character *killboard.Character

	key := fmt.Sprintf("character:%d", id)

	result, err := i.Redis.Get(key).Result()
	if err != nil && err.Error() != RedisNilErr {
		return nil, err
	}

	if result != "" {
		bStr := []byte(result)
		err = json.Unmarshal(bStr, character)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal result onto struct")
		}

		return character, nil
	}

	err = boiler.Characters(
		boiler.CharacterWhere.ID.EQ(uint64(id)),
	).Bind(context.Background(), i.DB, character)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "unable to query character record from the database")
	}

	if err == nil {
		byteCharacter, err := json.Marshal(character)
		if err != nil {
			i.Logger.WithField("id", id).WithError(err).Error("failed to marshal character")
		}

		_, err = i.Redis.Set(key, string(byteCharacter), time.Minute*60).Result()
		if err != nil {
			i.Logger.WithField("id", id).WithError(err).Error("failed to cache character")
		}

		return character, nil
	}

	response, err := i.ESI.GetCharactersCharacterID(id, "")
	if err != nil {
		i.Logger.WithError(err).Error("unable to retrieve character for provided id")
		return nil, errors.Wrap(err, "unable to retrieve character for provided id")
	}

	character = response.Data.(*killboard.Character)

	bCharacter := boiler.Character{}
	err = copier.Copy(&bCharacter, character)
	if err != nil {
		i.Logger.WithError(err).Error("unable to copy character to data struct")
		return nil, errors.Wrap(err, "unable to copy character to data struct")
	}

	err = bCharacter.Insert(context.Background(), i.DB, boil.Infer())
	if err != nil {
		i.Logger.WithError(err).Error("unable to insert character into database")
		return nil, errors.Wrap(err, "unable to insert character into database")
	}

	byteCharacter, err := json.Marshal(character)
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to marshal character")
	}

	_, err = i.Redis.Set(key, string(byteCharacter), time.Minute*60).Result()
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to cache character")
	}

	return character, nil

}
