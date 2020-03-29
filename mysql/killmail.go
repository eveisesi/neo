package mysql

import (
	"context"
	"fmt"

	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/ddouglas/killboard/mysql/boiler"

	"github.com/ddouglas/killboard"
	"github.com/jmoiron/sqlx"
)

type killmailRepository struct {
	db *sqlx.DB
}

func NewKillmailRepository(db *sqlx.DB) killboard.KillmailRespository {
	return &killmailRepository{
		db,
	}
}

func (r *killmailRepository) Killmail(ctx context.Context, id uint64) (*killboard.Killmail, error) {

	var killmail = killboard.Killmail{}
	err := boiler.Killmails(
		boiler.KillmailWhere.ID.EQ(id),
	).Bind(ctx, r.db, &killmail)

	return &killmail, err

}

func (r *killmailRepository) KillmailAttackersByKillmailIDs(ctx context.Context, ids []uint64) ([]*killboard.KillmailAttacker, error) {

	var attackers = make([]*killboard.KillmailAttacker, 0)
	err := boiler.KillmailAttackers(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.KillmailAttackerColumns.KillmailID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &attackers)

	return attackers, err

}

func (r *killmailRepository) KillmailItemsByKillmailIDs(ctx context.Context, ids []uint64) ([]*killboard.KillmailItem, error) {

	var items = make([]*killboard.KillmailItem, 0)
	err := boiler.KillmailItems(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.KillmailItemColumns.KillmailID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &items)

	return items, err

}

func (r *killmailRepository) KillmailItemsByParentIDs(ctx context.Context, ids []uint64) ([]*killboard.KillmailItem, error) {

	var items = make([]*killboard.KillmailItem, 0)
	err := boiler.KillmailItems(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.KillmailItemColumns.ParentID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &items)

	return items, err

}

func (r *killmailRepository) KillmailVictimsByKillmailIDs(ctx context.Context, ids []uint64) ([]*killboard.KillmailVictim, error) {

	var victims = make([]*killboard.KillmailVictim, 0)
	err := boiler.KillmailVictims(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.KillmailVictimColumns.KillmailID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &victims)

	return victims, err

}

func (r *killmailRepository) KillmailsByCharacterID(ctx context.Context, id uint64) ([]*killboard.Killmail, error) {

	var killmails = make([]*killboard.Killmail, 0)

	err := boiler.Killmails(
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				boiler.TableNames.KillmailVictim,
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.KillmailID,
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
			),
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				boiler.TableNames.KillmailAttackers,
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.KillmailID,
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
			),
		),
		qm.Where(
			fmt.Sprintf(
				"%s.%s = ?",
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.CharacterID,
			),
			id,
		),
		qm.Where(
			fmt.Sprintf(
				"%s.%s = ?",
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.CharacterID,
			),
			id,
		),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}

func (r *killmailRepository) KillmailsByCorporationID(ctx context.Context, id uint64) ([]*killboard.Killmail, error) {

	var killmails = make([]*killboard.Killmail, 0)

	err := boiler.Killmails(
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				boiler.TableNames.KillmailVictim,
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.KillmailID,
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
			),
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				boiler.TableNames.KillmailAttackers,
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.KillmailID,
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
			),
		),
		qm.Where(
			fmt.Sprintf(
				"%s.%s = ?",
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.CharacterID,
			),
			id,
		),
		qm.Where(
			fmt.Sprintf(
				"%s.%s = ?",
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.CharacterID,
			),
			id,
		),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}

func (r *killmailRepository) KillmailsByAllianceID(ctx context.Context, id uint64) ([]*killboard.Killmail, error) {

	var killmails = make([]*killboard.Killmail, 0)

	err := boiler.Killmails(
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				boiler.TableNames.KillmailVictim,
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.KillmailID,
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
			),
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				boiler.TableNames.KillmailAttackers,
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.KillmailID,
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
			),
		),
		qm.Where(
			fmt.Sprintf(
				"%s.%s = ?",
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.CharacterID,
			),
			id,
		),
		qm.Where(
			fmt.Sprintf(
				"%s.%s = ?",
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.CharacterID,
			),
			id,
		),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}
func (r *killmailRepository) KillmailsByFactionID(ctx context.Context, id uint64) ([]*killboard.Killmail, error) {

	var killmails = make([]*killboard.Killmail, 0)

	err := boiler.Killmails(
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				boiler.TableNames.KillmailVictim,
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.KillmailID,
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
			),
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s.%s = %s.%s",
				boiler.TableNames.KillmailAttackers,
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.KillmailID,
				boiler.TableNames.Killmails,
				boiler.KillmailColumns.ID,
			),
		),
		qm.Where(
			fmt.Sprintf(
				"%s.%s = ?",
				boiler.TableNames.KillmailVictim,
				boiler.KillmailVictimColumns.CharacterID,
			),
			id,
		),
		qm.Where(
			fmt.Sprintf(
				"%s.%s = ?",
				boiler.TableNames.KillmailAttackers,
				boiler.KillmailAttackerColumns.CharacterID,
			),
			id,
		),
	).Bind(ctx, r.db, &killmails)

	return killmails, err

}
