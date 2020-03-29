package ingress

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

func (i *Ingresser) GetCorporationByID(id uint64) (*killboard.Corporation, error) {

	var corporation = new(killboard.Corporation)

	key := fmt.Sprintf("corporation:%d", id)

	result, err := i.Redis.Get(key).Result()
	if err != nil && err.Error() != RedisNilErr {
		return nil, err
	}

	if result != "" {
		bStr := []byte(result)
		err = json.Unmarshal(bStr, corporation)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal result onto struct")
		}

		return corporation, nil
	}

	err = boiler.Corporations(
		boiler.CorporationWhere.ID.EQ(uint64(id)),
	).Bind(context.Background(), i.DB, corporation)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "unable to query corporation record from the database")
	}

	if err == nil {
		byteCorporation, err := json.Marshal(corporation)
		if err != nil {
			i.Logger.WithField("id", id).WithError(err).Error("failed to marshal corporation")
		}

		_, err = i.Redis.Set(key, string(byteCorporation), time.Minute*60).Result()
		if err != nil {
			i.Logger.WithField("id", id).WithError(err).Error("failed to cache corporation")
		}

		return corporation, nil
	}

	response, err := i.ESI.GetCorporationsCorporationID(id, "")
	if err != nil {
		i.Logger.WithError(err).Error("unable to retrieve corporation for provided id")
		return nil, errors.Wrap(err, "unable to retrieve corporation for provided id")
	}

	corporation = response.Data.(*killboard.Corporation)

	bCorporation := boiler.Corporation{}
	err = copier.Copy(&bCorporation, corporation)
	if err != nil {
		i.Logger.WithError(err).Error("unable to copy corporation to data struct")
		return nil, errors.Wrap(err, "unable to copy corporation to data struct")
	}

	err = bCorporation.Insert(context.Background(), i.DB, boil.Infer())
	if err != nil {
		i.Logger.WithError(err).Error("unable to insert corporation into database")
		return nil, errors.Wrap(err, "unable to insert corporation into database")
	}

	byteCorporation, err := json.Marshal(corporation)
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to marshal corporation")
	}

	_, err = i.Redis.Set(key, string(byteCorporation), time.Minute*60).Result()
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to cache corporation")
	}

	return corporation, nil

}
