package ingress

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/neo/esi"

	"github.com/ddouglas/neo"
	"github.com/ddouglas/neo/mysql/boiler"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (i *Ingresser) GetTypeByID(id uint64) (*killboard.Type, error) {

	var invType = new(killboard.Type)

	key := fmt.Sprintf("type:%d", id)
	ikey := fmt.Sprintf("itype:%d", id)

	result, err := i.Redis.Get(ikey).Result()
	if err != nil && err.Error() != RedisNilErr {
		return nil, err
	}

	if result != "" {
		return nil, errors.New("type invalid, returning early")
	}

	result, err = i.Redis.Get(key).Result()
	if err != nil && err.Error() != RedisNilErr {
		return nil, err
	}

	if result != "" {
		bStr := []byte(result)
		err = json.Unmarshal(bStr, invType)
		if err != nil {
			fmt.Printf("\n\n")
			fmt.Printf(`%s`, string(bStr))
			fmt.Printf("\n\n")
			i.Logger.WithError(err).Error("unable to unmarshal result onto struct")
			return nil, errors.Wrap(err, "unable to unmarshal result onto struct")
		}

		return invType, nil
	}

	err = boiler.Types(
		qm.Where(boiler.TypeColumns.ID+"=?", id),
	).Bind(context.Background(), i.DB, invType)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "unable to query type record from the database")
	}

	if err == nil {
		bType, err := json.Marshal(invType)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal type")
		}

		_, err = i.Redis.Set(key, string(bType), time.Minute*60).Result()
		if err != nil {
			return nil, errors.Wrap(err, "failed to cache type")
		}

		return invType, nil
	}

	response, err := i.ESI.GetUniverseTypesTypeID(id)
	if err != nil {
		if err == esi.TypeNotFound {
			_, err := i.Redis.Set(ikey, "invalid", 0).Result()
			if err != nil {
				return nil, errors.Wrap(err, "failed to cache type")
			}
		}
		return nil, errors.Wrap(err, "unable to retrieve type for provided id")
	}

	invType = response.Data.(*killboard.Type)

	bType := boiler.Type{}
	err = copier.Copy(&bType, invType)

	if err != nil {
		return nil, errors.Wrap(err, "unable to copy type to data struct")
	}

	err = bType.Insert(context.Background(), i.DB, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "unable to insert type into database")
	}

	byteInvType, err := json.Marshal(invType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal type")
	}

	_, err = i.Redis.Set(key, string(byteInvType), time.Minute*60).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to cache type")
	}

	return invType, nil

}
