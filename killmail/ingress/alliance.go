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

func (i *Ingresser) GetAllianceByID(id uint64) (*killboard.Alliance, error) {
	var alliance = new(killboard.Alliance)

	key := fmt.Sprintf("alliance:%d", id)

	result, err := i.Redis.Get(key).Result()
	if err != nil && err.Error() != RedisNilErr {
		return nil, err
	}

	if result != "" {
		bStr := []byte(result)
		err = json.Unmarshal(bStr, alliance)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal result onto struct")
		}

		return alliance, nil
	}

	err = boiler.Alliances(
		boiler.AllianceWhere.ID.EQ(uint64(id)),
	).Bind(context.Background(), i.DB, alliance)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "unable to query alliance record from the database")
	}

	if err == nil {
		byteAlliance, err := json.Marshal(alliance)
		if err != nil {
			i.Logger.WithField("id", id).WithError(err).Error("failed to marshal alliance")
		}

		_, err = i.Redis.Set(key, string(byteAlliance), time.Minute*60).Result()
		if err != nil {
			i.Logger.WithField("id", id).WithError(err).Error("failed to cache alliance")
		}

		return alliance, nil
	}

	response, err := i.ESI.GetAlliancesAllianceID(id, "")
	if err != nil {
		i.Logger.WithError(err).Error("unable to retrieve alliance for provided id")
		return nil, errors.Wrap(err, "unable to retrieve alliance for provided id")
	}

	alliance = response.Data.(*killboard.Alliance)

	bAlliance := boiler.Alliance{}
	err = copier.Copy(&bAlliance, alliance)
	if err != nil {
		i.Logger.WithError(err).Error("unable to copy alliance to data struct")
		return nil, errors.Wrap(err, "unable to copy alliance to data struct")
	}

	err = bAlliance.Insert(context.Background(), i.DB, boil.Infer())
	if err != nil {
		i.Logger.WithError(err).Error("unable to insert alliance into database")
		return nil, errors.Wrap(err, "unable to insert alliance into database")
	}

	byteAlliance, err := json.Marshal(alliance)
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to marshal alliance")
	}

	_, err = i.Redis.Set(key, string(byteAlliance), time.Minute*60).Result()
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to cache alliance")
	}

	return alliance, nil

}
