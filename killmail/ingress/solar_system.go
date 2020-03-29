package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/neo"
	"github.com/ddouglas/neo/mysql/boiler"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (i *Ingresser) GetSolarSystemByID(id uint64) (*killboard.SolarSystem, error) {

	var solarSystem = new(killboard.SolarSystem)

	key := fmt.Sprintf("system:%d", id)

	result, err := i.Redis.Get(key).Result()
	if err != nil && err.Error() != RedisNilErr {
		return nil, err
	}

	if result != "" {
		bStr := []byte(result)
		err = json.Unmarshal(bStr, solarSystem)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal result onto struct")
		}

		return solarSystem, nil
	}

	err = boiler.SolarSystems(
		qm.Where(boiler.SolarSystemColumns.ID+"=?", id),
	).Bind(context.Background(), i.DB, solarSystem)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query solar system record from the database")
	}

	bSolarSystem, err := json.Marshal(solarSystem)
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to marshal solar system")
	}

	_, err = i.Redis.Set(key, string(bSolarSystem), time.Minute*60).Result()
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to cache solar system")
	}

	return solarSystem, nil

}
