package mysql

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/eveisesi/neo/mysql/boiler"

	"github.com/eveisesi/neo"
	"github.com/jmoiron/sqlx"
)

type killmailRepository struct {
	db *sqlx.DB
}

func NewKillmailRepository(db *sqlx.DB) neo.KillmailRepository {
	return &killmailRepository{
		db,
	}
}

func (r *killmailRepository) Killmail(ctx context.Context, id uint64, hash string) (*neo.Killmail, error) {

	var killmail = neo.Killmail{}
	err := boiler.Killmails(
		boiler.KillmailWhere.ID.EQ(id),
		boiler.KillmailWhere.Hash.EQ(hash),
	).Bind(ctx, r.db, &killmail)

	return &killmail, err

}

func (r *killmailRepository) ByIDs(ctx context.Context, ids []uint64) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	err := boiler.Killmails(
		qm.WhereIn(boiler.KillmailColumns.ID+" IN ?", convertSliceUint64ToSliceInterface(ids)...),
		qm.OrderBy(boiler.KillmailColumns.ID+" DESC"),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}

func (r *killmailRepository) Create(ctx context.Context, killmail *neo.Killmail) (*neo.Killmail, error) {

	var bKillmail = new(boiler.Killmail)
	err := copier.Copy(bKillmail, killmail)
	if err != nil {
		return nil, errors.Wrap(err, "unable to copy killmail to orm")
	}

	err = bKillmail.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "unable to insert killmail into db")
	}

	err = copier.Copy(killmail, bKillmail)

	return killmail, errors.Wrap(err, "unable to copy orm to killmail")

}

func (r *killmailRepository) CreateWithTxn(ctx context.Context, txn neo.Transactioner, killmail *neo.Killmail) (*neo.Killmail, error) {
	var t = txn.(*transaction)
	var bKillmail = new(boiler.Killmail)
	err := copier.Copy(bKillmail, killmail)
	if err != nil {
		return nil, errors.Wrap(err, "unable to copy killmail to orm")
	}

	err = bKillmail.Insert(ctx, t, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "unable to insert killmail into db")
	}

	err = copier.Copy(killmail, bKillmail)

	return killmail, errors.Wrap(err, "unable to copy orm to killmail")

}

func (r *killmailRepository) Update(ctx context.Context, id uint64, hash string, killmail *neo.Killmail) error {

	var bKillmail = new(boiler.Killmail)
	err := copier.Copy(bKillmail, killmail)
	if err != nil {
		return errors.Wrap(err, "unable to copy killmail to orm")
	}

	bKillmail.ID = id
	bKillmail.Hash = hash

	_, err = bKillmail.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "failed to update killmail in db")
	}

	err = copier.Copy(killmail, bKillmail)

	return errors.Wrap(err, "unable to copy orm to killmail")

}

func (r *killmailRepository) Exists(ctx context.Context, id uint64, hash string) (bool, error) {
	return boiler.KillmailExists(ctx, r.db, id, hash)
}

func (r *killmailRepository) Recent(ctx context.Context, limit, offset int) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	err := boiler.Killmails(
		qm.Limit(limit),
		qm.Offset(offset),
		qm.OrderBy(
			fmt.Sprintf(
				"%s DESC",
				boiler.KillmailColumns.ID,
			),
		),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}

func (r *killmailRepository) ByCharacterID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {

	query := `
		SELECT
			killmails.id,
			killmails.hash,
			killmails.moon_id,
			killmails.solar_system_id,
			killmails.war_id,
			killmails.is_npc,
			killmails.is_awox,
			killmails.is_solo,
			killmails.dropped_value,
			killmails.destroyed_value,
			killmails.fitted_value,
			killmails.total_value,
			killmails.killmail_time
		FROM (
			(
				SELECT 
					DISTINCT(killmail_victim.killmail_id) 
				FROM killmail_victim 
				WHERE 
					killmail_victim.character_id = ?
				ORDER BY killmail_victim.killmail_id DESC 
				LIMIT 1000
			)
			UNION ALL
			(
				SELECT 
					DISTINCT(killmail_attackers.killmail_id) 
				FROM killmail_attackers
				WHERE 
					killmail_attackers.character_id = ? 
				ORDER BY killmail_attackers.killmail_id DESC 
				LIMIT 1000
			)
		) SELECTED_KMS
		LEFT JOIN killmails ON killmails.id = SELECTED_KMS.killmail_id
		ORDER BY killmails.killmail_time DESC
		LIMIT 1000
	`

	killmails := make([]*neo.Killmail, 0)
	err := r.db.SelectContext(ctx, &killmails, query, id, id)

	return killmails, err

}

func (r *killmailRepository) ByCorporationID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {

	query := `
		SELECT
			killmails.id,
			killmails.hash,
			killmails.moon_id,
			killmails.solar_system_id,
			killmails.war_id,
			killmails.is_npc,
			killmails.is_awox,
			killmails.is_solo,
			killmails.dropped_value,
			killmails.destroyed_value,
			killmails.fitted_value,
			killmails.total_value,
			killmails.killmail_time
		FROM (
			(
				SELECT 
					DISTINCT(killmail_victim.killmail_id) 
				FROM killmail_victim 
				WHERE 
					killmail_victim.corporation_id = ?
				ORDER BY killmail_victim.killmail_id DESC 
				LIMIT 1000
			)
			UNION ALL
			(
				SELECT 
					DISTINCT(killmail_attackers.killmail_id) 
				FROM killmail_attackers
				WHERE 
					killmail_attackers.corporation_id = ? 
				ORDER BY killmail_attackers.killmail_id DESC 
				LIMIT 1000
			)
		) SELECTED_KMS
		LEFT JOIN killmails ON killmails.id = SELECTED_KMS.killmail_id
		ORDER BY killmails.killmail_time DESC
		LIMIT 1000
	`

	killmails := make([]*neo.Killmail, 0)
	err := r.db.SelectContext(ctx, &killmails, query, id, id)

	return killmails, err

}

func (r *killmailRepository) ByAllianceID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {

	query := `
		SELECT
			killmails.id,
			killmails.hash,
			killmails.moon_id,
			killmails.solar_system_id,
			killmails.war_id,
			killmails.is_npc,
			killmails.is_awox,
			killmails.is_solo,
			killmails.dropped_value,
			killmails.destroyed_value,
			killmails.fitted_value,
			killmails.total_value,
			killmails.killmail_time
		FROM (
			(
				SELECT 
					DISTINCT(killmail_victim.killmail_id) 
				FROM killmail_victim 
				WHERE 
					killmail_victim.alliance_id = ?
				ORDER BY killmail_victim.killmail_id DESC 
				LIMIT 1000
			)
			UNION ALL
			(
				SELECT 
					DISTINCT(killmail_attackers.killmail_id) 
				FROM killmail_attackers
				WHERE 
					killmail_attackers.alliance_id = ? 
				ORDER BY killmail_attackers.killmail_id DESC 
				LIMIT 1000
			)
		) SELECTED_KMS
		LEFT JOIN killmails ON killmails.id = SELECTED_KMS.killmail_id
		ORDER BY killmails.killmail_time DESC
		LIMIT 1000
	`

	killmails := make([]*neo.Killmail, 0)
	err := r.db.SelectContext(ctx, &killmails, query, id, id)

	return killmails, err

}

func (r *killmailRepository) ByShipID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {

	query := `
		SELECT
			killmails.id,
			killmails.hash,
			killmails.moon_id,
			killmails.solar_system_id,
			killmails.war_id,
			killmails.is_npc,
			killmails.is_awox,
			killmails.is_solo,
			killmails.dropped_value,
			killmails.destroyed_value,
			killmails.fitted_value,
			killmails.total_value,
			killmails.killmail_time
		FROM (
			(
				SELECT 
					DISTINCT(killmail_victim.killmail_id) 
				FROM killmail_victim 
				WHERE 
					killmail_victim.ship_type_id = ? 
				ORDER BY killmail_victim.killmail_id DESC 
				LIMIT 1000
			)
			UNION ALL
			(
				SELECT 
					DISTINCT(killmail_attackers.killmail_id) 
				FROM killmail_attackers
				WHERE 
					killmail_attackers.ship_type_id = ? 
				ORDER BY killmail_attackers.killmail_id DESC 
				LIMIT 1000
			)
		) SELECTED_KMS
		LEFT JOIN killmails ON killmails.id = SELECTED_KMS.killmail_id
		ORDER BY killmails.killmail_time DESC
		LIMIT 1000
	`

	killmails := make([]*neo.Killmail, 0)
	err := r.db.SelectContext(ctx, &killmails, query, id, id)

	return killmails, err

}

func (r *killmailRepository) ByShipGroupID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {

	query := `
		SELECT
			killmails.id,
			killmails.hash,
			killmails.moon_id,
			killmails.solar_system_id,
			killmails.war_id,
			killmails.is_npc,
			killmails.is_awox,
			killmails.is_solo,
			killmails.dropped_value,
			killmails.destroyed_value,
			killmails.fitted_value,
			killmails.total_value,
			killmails.killmail_time
		FROM ( 
			(
				SELECT 
						DISTINCT(killmail_attackers.killmail_id)
				FROM killmail_attackers
				WHERE 
					ship_type_id IN (
						SELECT id from types where group_id = ? AND published = 1
					)
				ORDER BY killmail_attackers.killmail_id DESC 
				LIMIT 1000
			) UNION ALL (
				SELECT 
					DISTINCT(killmail_victim.killmail_id) 
				FROM killmail_victim 
				WHERE 
					ship_type_id IN (
						SELECT id from types where group_id = ? AND published = 1
					)
				ORDER BY killmail_victim.killmail_id DESC 
				LIMIT 1000
			) 
		) SELECTED_KMS
		LEFT JOIN killmails ON killmails.id = SELECTED_KMS.killmail_id
		ORDER BY killmails.killmail_time DESC
		LIMIT 1000
	`

	killmails := make([]*neo.Killmail, 0)
	err = r.db.SelectContext(ctx, &killmails, query, id, id)

	return killmails, err

}

func (r *killmailRepository) BySystemID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {

	killmails := make([]*neo.Killmail, 0)
	err := boiler.Killmails(
		boiler.KillmailWhere.SolarSystemID.EQ(id),
		qm.OrderBy(boiler.KillmailColumns.KillmailTime+" DESC"),
		qm.Limit(1000),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}

func (r *killmailRepository) ByConstellationID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {

	query := `
		SELECT 
			killmails.id,
			killmails.hash,
			killmails.moon_id,
			killmails.solar_system_id,
			killmails.war_id,
			killmails.is_npc,
			killmails.is_awox,
			killmails.is_solo,
			killmails.dropped_value,
			killmails.destroyed_value,
			killmails.fitted_value,
			killmails.total_value,
			killmails.killmail_time
		FROM killmails 
		WHERE 
			killmails.solar_system_id IN (
				SELECT 
					id 
				FROM solar_systems 
				WHERE constellation_id = ?
			) 
		ORDER BY killmails.killmail_time DESC 
		LIMIT 5000
	`

	killmails := make([]*neo.Killmail, 0)
	err = r.db.SelectContext(ctx, &killmails, query, id)

	return killmails, err

}

func (r *killmailRepository) ByRegionID(ctx context.Context, id uint64) ([]*neo.Killmail, error) {

	query := `
		SELECT 
			killmails.id,
			killmails.hash,
			killmails.moon_id,
			killmails.solar_system_id,
			killmails.war_id,
			killmails.is_npc,
			killmails.is_awox,
			killmails.is_solo,
			killmails.dropped_value,
			killmails.destroyed_value,
			killmails.fitted_value,
			killmails.total_value,
			killmails.killmail_time
		FROM killmails 
		WHERE 
			killmails.solar_system_id IN (
				SELECT 
					id 
				FROM solar_systems 
				WHERE constellation_id IN (
					SELECT 
						id 
					FROM constellations 
					WHERE region_id = ?
				)
			) 
		ORDER BY killmails.killmail_time DESC 
		LIMIT 5000
	`

	killmails := make([]*neo.Killmail, 0)
	err = r.db.SelectContext(ctx, &killmails, query, id)

	return killmails, err

}
