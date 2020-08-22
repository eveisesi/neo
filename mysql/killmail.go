package mysql

import (
	"context"
	"fmt"
	"strings"

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

func (r *killmailRepository) AllKillmails(ctx context.Context, coreMods []neo.Modifier, vicMods []neo.Modifier, attMods []neo.Modifier) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	var qmMods = make([]qm.QueryMod, 0)

	qmMods = append(qmMods, qm.Select("DISTINCT killmails.*"))
	qmMods = append(qmMods, qm.From("killmails"))

	if len(coreMods) > 0 {
		qmMods = append(qmMods, BuildQueryModifiers(boiler.TableNames.Killmails, coreMods...)...)
	}

	if len(vicMods) > 0 {
		vicModifiers, vicArgs := BuildJoinCondition(boiler.TableNames.KillmailVictim, vicMods...)
		vicModifiers = append([]string{
			fmt.Sprintf(
				"%s.%s = %s.%s",
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.KillmailID,
			),
		}, vicModifiers...)
		vicModStr := strings.Join(vicModifiers, " AND ")
		qmMods = append(qmMods, qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s",
				boiler.TableNames.KillmailVictim,
				vicModStr,
			), vicArgs...,
		))
	}

	if len(attMods) > 0 {
		attModifiers, attArgs := BuildJoinCondition(boiler.TableNames.KillmailAttackers, attMods...)
		attModifiers = append([]string{
			fmt.Sprintf(
				"%s.%s = %s.%s",
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.KillmailID,
			),
		}, attModifiers...)
		attModStr := strings.Join(attModifiers, " AND ")
		qmMods = append(qmMods, qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s",
				boiler.TableNames.KillmailAttackers,
				attModStr,
			), attArgs...,
		))
	}

	if len(qmMods) == 0 {
		return nil, errors.New("Atleast one modifier must be specified")
	}

	err := boiler.NewQuery(qmMods...).Bind(ctx, r.db, &killmails)

	return killmails, err
}

func (r *killmailRepository) Killmail(ctx context.Context, id uint) (*neo.Killmail, error) {

	var killmail = neo.Killmail{}
	err := boiler.Killmails(
		boiler.KillmailWhere.ID.EQ(id),
	).Bind(ctx, r.db, &killmail)

	return &killmail, err

}

func (r *killmailRepository) ByIDs(ctx context.Context, ids []uint) ([]*neo.Killmail, error) {

	var killmails = make([]*neo.Killmail, 0)
	err := boiler.Killmails(
		qm.WhereIn(boiler.KillmailColumns.ID+" IN ?", convertSliceUintToSliceInterface(ids)...),
		qm.OrderBy(boiler.KillmailColumns.ID+" DESC"),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}

func (r *killmailRepository) Create(ctx context.Context, killmail *neo.Killmail) (*neo.Killmail, error) {

	query := `
		INSERT INTO killmails
			(id, hash, moon_id, solar_system_id, constellation_id, region_id, war_id, is_npc, is_awox, is_solo, dropped_value, destroyed_value, fitted_value, total_value, killmail_time, created_at, updated_at)
		VALUES (
			?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW()
		)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		killmail.ID, killmail.Hash, killmail.MoonID.Uint,
		killmail.SolarSystemID, killmail.ConstellationID, killmail.RegionID,
		killmail.WarID.Uint, killmail.IsNPC, killmail.IsAwox,
		killmail.IsSolo, killmail.DroppedValue, killmail.DestroyedValue,
		killmail.FittedValue, killmail.TotalValue, killmail.KillmailTime,
	)

	return r.Killmail(ctx, killmail.ID)

}

func (r *killmailRepository) CreateWithTxn(ctx context.Context, txn neo.Transactioner, killmail *neo.Killmail) (*neo.Killmail, error) {
	var t = txn.(*transaction)

	query := `
		INSERT INTO killmails
			(id, hash, moon_id, solar_system_id, constellation_id, region_id, war_id, is_npc, is_awox, is_solo, dropped_value, destroyed_value, fitted_value, total_value, killmail_time, created_at, updated_at)
		VALUES (
			?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW()
		)
	`

	_, err := t.ExecContext(
		ctx,
		query,
		killmail.ID, killmail.Hash, killmail.MoonID.Uint,
		killmail.SolarSystemID, killmail.ConstellationID, killmail.RegionID,
		killmail.WarID.Uint, killmail.IsNPC, killmail.IsAwox,
		killmail.IsSolo, killmail.DroppedValue, killmail.DestroyedValue,
		killmail.FittedValue, killmail.TotalValue, killmail.KillmailTime,
	)

	return killmail, err

}

func (r *killmailRepository) Update(ctx context.Context, id uint, killmail *neo.Killmail) error {

	var bKillmail = new(boiler.Killmail)
	err := copier.Copy(bKillmail, killmail)
	if err != nil {
		return errors.Wrap(err, "unable to copy killmail to orm")
	}

	bKillmail.ID = id

	_, err = bKillmail.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "failed to update killmail in db")
	}

	err = copier.Copy(killmail, bKillmail)

	return errors.Wrap(err, "unable to copy orm to killmail")

}

func (r *killmailRepository) UpdateWithTxn(ctx context.Context, txn neo.Transactioner, killmail *neo.Killmail) error {
	var t = txn.(*transaction)
	var bKillmail = new(boiler.Killmail)
	err := copier.Copy(bKillmail, killmail)
	if err != nil {
		return errors.Wrap(err, "unable to copy killmail to orm")
	}

	_, err = bKillmail.Update(ctx, t, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "unable to update killmail in db")
	}

	return errors.Wrap(err, "unable to copy orm to killmail")
}

func (r *killmailRepository) Exists(ctx context.Context, id uint) (bool, error) {
	return boiler.KillmailExists(ctx, r.db, id)
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

func (r *killmailRepository) Recalculable(ctx context.Context, limit int, after uint) ([]*neo.Killmail, error) {

	mods := []qm.QueryMod{}
	mods = append(mods,
		boiler.KillmailWhere.ID.GT(after),
		qm.OrderBy(fmt.Sprintf("%s ASC", boiler.KillmailColumns.ID)),
	)
	mods = append(mods, qm.Limit(limit))

	var killmails = make([]*neo.Killmail, 0)
	err := boiler.Killmails(mods...).Bind(ctx, r.db, &killmails)

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

func (r *killmailRepository) ByCorporationID(ctx context.Context, id uint) ([]*neo.Killmail, error) {

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

func (r *killmailRepository) ByAllianceID(ctx context.Context, id uint) ([]*neo.Killmail, error) {

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

func (r *killmailRepository) ByShipID(ctx context.Context, id uint) ([]*neo.Killmail, error) {

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

func (r *killmailRepository) ByShipGroupID(ctx context.Context, id uint) ([]*neo.Killmail, error) {

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

func (r *killmailRepository) BySystemID(ctx context.Context, id uint) ([]*neo.Killmail, error) {

	killmails := make([]*neo.Killmail, 0)
	err := boiler.Killmails(
		boiler.KillmailWhere.SolarSystemID.EQ(id),
		qm.OrderBy(boiler.KillmailColumns.KillmailTime+" DESC"),
		qm.Limit(1000),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}

func (r *killmailRepository) ByConstellationID(ctx context.Context, id uint) ([]*neo.Killmail, error) {

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

func (r *killmailRepository) ByRegionID(ctx context.Context, id uint) ([]*neo.Killmail, error) {

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
