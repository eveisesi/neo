package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/killboard"
	"github.com/ddouglas/killboard/mysql/boiler"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (i *Ingresser) GetTypeByID(id uint64) (*killboard.Type, error) {

	var invType killboard.Type

	key := fmt.Sprintf("type:%d", id)

	result, err := i.Redis.Get(key).Result()
	if err != nil && err.Error() != RedisNilErr {
		return nil, err
	}

	if result != "" {
		bStr := []byte(result)
		err = json.Unmarshal(bStr, &invType)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal result onto struct")
		}

		return &invType, nil
	}

	err = boiler.Types(
		qm.Where(boiler.TypeColumns.ID+"=?", id),
	).Bind(context.Background(), i.DB, &invType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query solar system record from the database")
	}

	bType, err := json.Marshal(invType)
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to marshal solar system")
	}

	_, err = i.Redis.Set(key, string(bType), time.Minute*60).Result()
	if err != nil {
		i.Logger.WithField("id", id).WithError(err).Error("failed to cache solar system")
	}

	return &invType, nil

}
