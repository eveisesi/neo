package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type KillmailRepository interface {
	Killmail(ctx context.Context, id uint64, hash string) (*Killmail, error)
	Create(ctx context.Context, killmail *Killmail) (*Killmail, error)
	CreateWithTxn(ctx context.Context, txn Transactioner, killmail *Killmail) (*Killmail, error)
	Update(ctx context.Context, id uint64, hash string, killmail *Killmail) error
	UpdateWithTxn(ctx context.Context, txn Transactioner, killmail *Killmail) error

	Exists(ctx context.Context, id uint64, hash string) (bool, error)
	Recent(ctx context.Context, limit, offset int) ([]*Killmail, error)
	Recalculable(ctx context.Context, limit int, after uint64) ([]*Killmail, error)

	ByIDs(ctx context.Context, ids []uint64) ([]*Killmail, error)
	ByCharacterID(ctx context.Context, id uint64) ([]*Killmail, error)
	ByCorporationID(ctx context.Context, id uint64) ([]*Killmail, error)
	ByAllianceID(ctx context.Context, id uint64) ([]*Killmail, error)
	ByShipID(ctx context.Context, id uint64) ([]*Killmail, error)
	ByShipGroupID(ctx context.Context, id uint64) ([]*Killmail, error)
	BySystemID(ctx context.Context, id uint64) ([]*Killmail, error)
	ByConstellationID(ctx context.Context, id uint64) ([]*Killmail, error)
	ByRegionID(ctx context.Context, id uint64) ([]*Killmail, error)
}

type MVRepository interface {
	All(ctx context.Context, limit, age int) ([]*Killmail, error)
	KillsByCharacterID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	LossesByCharacterID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	KillsByCorporationID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	LossesByCorporationID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	KillsByAllianceID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	LossesByAllianceID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	KillsByShipID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	LossesByShipID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	KillsByShipGroupID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	LossesByShipGroupID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	KillsBySystemID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	KillsByConstellationID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
	KillsByRegionID(ctx context.Context, id uint64, limit, age int) ([]*Killmail, error)
}

type KillmailAttackerRepository interface {
	ByKillmailID(ctx context.Context, id uint64) ([]*KillmailAttacker, error)
	ByKillmailIDs(ctx context.Context, ids []uint64) ([]*KillmailAttacker, error)
	Create(ctx context.Context, attacker *KillmailAttacker) (*KillmailAttacker, error)
	CreateWithTxn(ctx context.Context, txn Transactioner, attacker *KillmailAttacker) (*KillmailAttacker, error)
	CreateBulk(ctx context.Context, attackers []*KillmailAttacker) ([]*KillmailAttacker, error)
	CreateBulkWithTxn(ctx context.Context, txn Transactioner, attackers []*KillmailAttacker) ([]*KillmailAttacker, error)
}

type KillmailItemRepository interface {
	ByKillmailID(ctx context.Context, id uint64) ([]*KillmailItem, error)
	ByKillmailIDs(ctx context.Context, ids []uint64) ([]*KillmailItem, error)
	Create(ctx context.Context, item *KillmailItem) (*KillmailItem, error)
	CreateWithTxn(ctx context.Context, txn Transactioner, item *KillmailItem) (*KillmailItem, error)
	CreateBulk(ctx context.Context, items []*KillmailItem) ([]*KillmailItem, error)
	CreateBulkWithTxn(ctx context.Context, txn Transactioner, items []*KillmailItem) ([]*KillmailItem, error)
	UpdateBulk(ctx context.Context, items []*KillmailItem) error
	UpdateBulkWithTxn(ctx context.Context, txn Transactioner, items []*KillmailItem) error
}

type KillmailVictimRepository interface {
	ByKillmailID(ctx context.Context, id uint64) (*KillmailVictim, error)
	ByKillmailIDs(ctx context.Context, ids []uint64) ([]*KillmailVictim, error)
	Create(ctx context.Context, victim *KillmailVictim) (*KillmailVictim, error)
	CreateWithTxn(ctx context.Context, txn Transactioner, victim *KillmailVictim) (*KillmailVictim, error)
	Update(ctx context.Context, victim *KillmailVictim) error
	UpdateWithTxn(ctx context.Context, txn Transactioner, victim *KillmailVictim) error
}

type Killmail struct {
	ID             uint64     `db:"id" json:"id"`
	Hash           string     `db:"hash" json:"hash"`
	MoonID         null.Int64 `db:"moon_id" json:"moon_id,omitempty"`
	SolarSystemID  uint64     `db:"solar_system_id" json:"solar_system_id"`
	WarID          null.Int64 `db:"war_id" json:"war_id,omitempty"`
	IsNPC          bool       `db:"is_npc" json:"isNPC"`
	IsAwox         bool       `db:"is_awox" json:"isAwox"`
	IsSolo         bool       `db:"is_solo" json:"isSolo"`
	DroppedValue   float64    `db:"dropped_value" json:"droppedValue"`
	DestroyedValue float64    `db:"destroyed_value" json:"destroyedValue"`
	FittedValue    float64    `db:"fitted_value" json:"fittedValue"`
	TotalValue     float64    `db:"total_value" json:"totalValue"`
	KillmailTime   time.Time  `db:"killmail_time" json:"killmail_time"`

	Attackers []*KillmailAttacker `json:"attackers"`
	Victim    *KillmailVictim     `json:"victim"`

	System *SolarSystem `json:"-"`
}

type KillmailAttacker struct {
	ID             uint64      `json:"id"`
	KillmailID     uint64      `json:"killmail_id"`
	AllianceID     null.Uint64 `json:"alliance_id"`
	CharacterID    null.Uint64 `json:"character_id"`
	CorporationID  null.Uint64 `json:"corporation_id"`
	FactionID      null.Uint64 `json:"faction_id"`
	DamageDone     uint64      `json:"damage_done"`
	FinalBlow      bool        `json:"final_blow"`
	SecurityStatus float64     `json:"security_status"`
	ShipTypeID     null.Uint64 `json:"ship_type_id"`
	WeaponTypeID   null.Uint64 `json:"weapon_type_id"`

	Alliance    *Alliance    `json:"-"`
	Character   *Character   `json:"-"`
	Corporation *Corporation `json:"-"`
	Ship        *Type        `json:"-"`
	Weapon      *Type        `json:"-"`
}

type KillmailItem struct {
	ID                uint64      `json:"id"`
	ParentID          null.Uint64 `json:"parent_id"`
	KillmailID        uint64      `json:"killmail_id"`
	Flag              uint64      `json:"flag"`
	ItemTypeID        uint64      `json:"item_type_id"`
	QuantityDropped   null.Uint64 `json:"quantity_dropped"`
	QuantityDestroyed null.Uint64 `json:"quantity_destroyed"`
	ItemValue         float64     `json:"itemValue"`
	TotalValue        float64     `json:"totalValue"`
	Singleton         uint64      `json:"singleton"`
	IsParent          bool        `json:"is_parent"`

	Item  *Type           `json:"-"`
	Items []*KillmailItem `json:"items"`
}

type KillmailVictim struct {
	ID            uint64            `json:"id"`
	KillmailID    uint64            `json:"killmail_id"`
	AllianceID    null.Uint64       `json:"alliance_id"`
	CharacterID   null.Uint64       `json:"character_id"`
	CorporationID null.Uint64       `json:"corporation_id"`
	FactionID     null.Uint64       `json:"faction_id"`
	DamageTaken   uint64            `json:"damage_taken"`
	ShipTypeID    uint64            `json:"ship_type_id"`
	ShipValue     float64           `json:"shipValue" `
	Position      *KillmailPosition `json:"position"`
	PosX          null.Float64      `json:"pos_x"`
	PosY          null.Float64      `json:"pos_y"`
	PosZ          null.Float64      `json:"pos_z"`

	Alliance    *Alliance    `json:"-"`
	Character   *Character   `json:"-"`
	Corporation *Corporation `json:"-"`
	Ship        *Type        `json:"-"`

	Items []*KillmailItem `json:"items"`
}

type KillmailPosition struct {
	X null.Float64 `json:"x"`
	Y null.Float64 `json:"y"`
	Z null.Float64 `json:"z"`
}
