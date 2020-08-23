package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type KillmailRepository interface {
	Killmail(ctx context.Context, id uint) (*Killmail, error)
	Killmails(ctx context.Context, coreMods []Modifier, vicMods []Modifier, attMods []Modifier) ([]*Killmail, error)
	Create(ctx context.Context, killmail *Killmail) (*Killmail, error)
	CreateWithTxn(ctx context.Context, txn Transactioner, killmail *Killmail) (*Killmail, error)
	Update(ctx context.Context, id uint, killmail *Killmail) error
	UpdateWithTxn(ctx context.Context, txn Transactioner, killmail *Killmail) error

	Exists(ctx context.Context, id uint) (bool, error)
	Recent(ctx context.Context, limit, offset int) ([]*Killmail, error)
	Recalculable(ctx context.Context, limit int, after uint) ([]*Killmail, error)
}

type KillmailAttackerRepository interface {
	ByKillmailID(ctx context.Context, id uint) ([]*KillmailAttacker, error)
	ByKillmailIDs(ctx context.Context, ids []uint) ([]*KillmailAttacker, error)
	Create(ctx context.Context, attacker *KillmailAttacker) (*KillmailAttacker, error)
	CreateWithTxn(ctx context.Context, txn Transactioner, attacker *KillmailAttacker) (*KillmailAttacker, error)
	CreateBulk(ctx context.Context, attackers []*KillmailAttacker) ([]*KillmailAttacker, error)
	CreateBulkWithTxn(ctx context.Context, txn Transactioner, attackers []*KillmailAttacker) ([]*KillmailAttacker, error)
}

type KillmailItemRepository interface {
	ByKillmailID(ctx context.Context, id uint) ([]*KillmailItem, error)
	ByKillmailIDs(ctx context.Context, ids []uint) ([]*KillmailItem, error)
	Create(ctx context.Context, item *KillmailItem) (*KillmailItem, error)
	CreateWithTxn(ctx context.Context, txn Transactioner, item *KillmailItem) (*KillmailItem, error)
	CreateBulk(ctx context.Context, items []*KillmailItem) ([]*KillmailItem, error)
	CreateBulkWithTxn(ctx context.Context, txn Transactioner, items []*KillmailItem) ([]*KillmailItem, error)
	UpdateBulk(ctx context.Context, items []*KillmailItem) error
	UpdateBulkWithTxn(ctx context.Context, txn Transactioner, items []*KillmailItem) error
}

type KillmailVictimRepository interface {
	ByKillmailID(ctx context.Context, id uint) (*KillmailVictim, error)
	ByKillmailIDs(ctx context.Context, ids []uint) ([]*KillmailVictim, error)
	Create(ctx context.Context, victim *KillmailVictim) (*KillmailVictim, error)
	CreateWithTxn(ctx context.Context, txn Transactioner, victim *KillmailVictim) (*KillmailVictim, error)
	Update(ctx context.Context, victim *KillmailVictim) error
	UpdateWithTxn(ctx context.Context, txn Transactioner, victim *KillmailVictim) error
}

type Killmail struct {
	ID              uint      `db:"id" json:"id"`
	Hash            string    `db:"hash" json:"hash"`
	MoonID          null.Uint `db:"moon_id" json:"moon_id,omitempty"`
	SolarSystemID   uint      `db:"solar_system_id" json:"solar_system_id"`
	ConstellationID uint      `db:"constellation_id"`
	RegionID        uint      `db:"region_id"`
	WarID           null.Uint `db:"war_id" json:"war_id,omitempty"`
	IsNPC           bool      `db:"is_npc" json:"isNPC"`
	IsAwox          bool      `db:"is_awox" json:"isAwox"`
	IsSolo          bool      `db:"is_solo" json:"isSolo"`
	DroppedValue    float64   `db:"dropped_value" json:"droppedValue"`
	DestroyedValue  float64   `db:"destroyed_value" json:"destroyedValue"`
	FittedValue     float64   `db:"fitted_value" json:"fittedValue"`
	TotalValue      float64   `db:"total_value" json:"totalValue"`
	KillmailTime    time.Time `db:"killmail_time" json:"killmail_time"`

	Attackers []*KillmailAttacker `json:"attackers"`
	Victim    *KillmailVictim     `json:"victim"`

	System *SolarSystem `json:"-"`
}

type KillmailAttacker struct {
	ID             uint        `json:"id"`
	KillmailID     uint        `json:"killmail_id"`
	AllianceID     null.Uint   `json:"alliance_id"`
	CharacterID    null.Uint64 `json:"character_id"`
	CorporationID  null.Uint   `json:"corporation_id"`
	FactionID      null.Uint   `json:"faction_id"`
	DamageDone     uint        `json:"damage_done"`
	FinalBlow      bool        `json:"final_blow"`
	SecurityStatus float64     `json:"security_status"`
	ShipTypeID     null.Uint   `json:"ship_type_id"`
	ShipGroupID    null.Uint
	WeaponTypeID   null.Uint `json:"weapon_type_id"`
	WeaponGroupID  null.Uint

	Alliance    *Alliance    `json:"-"`
	Character   *Character   `json:"-"`
	Corporation *Corporation `json:"-"`
	Ship        *Type        `json:"-"`
	Weapon      *Type        `json:"-"`
}

type KillmailItem struct {
	ID                uint      `json:"id"`
	ParentID          null.Uint `json:"parent_id"`
	KillmailID        uint      `json:"killmail_id"`
	Flag              uint      `json:"flag"`
	ItemTypeID        uint      `json:"item_type_id"`
	ItemGroupID       uint
	QuantityDropped   null.Uint `json:"quantity_dropped"`
	QuantityDestroyed null.Uint `json:"quantity_destroyed"`
	ItemValue         float64   `json:"itemValue"`
	TotalValue        float64   `json:"totalValue"`
	Singleton         uint8     `json:"singleton"`
	IsParent          bool      `json:"is_parent"`

	Item  *Type           `json:"-"`
	Items []*KillmailItem `json:"items"`
}

type KillmailVictim struct {
	ID            uint        `json:"id"`
	KillmailID    uint        `json:"killmail_id"`
	AllianceID    null.Uint   `json:"alliance_id"`
	CharacterID   null.Uint64 `json:"character_id"`
	CorporationID null.Uint   `json:"corporation_id"`
	FactionID     null.Uint   `json:"faction_id"`
	DamageTaken   uint        `json:"damage_taken"`
	ShipTypeID    uint        `json:"ship_type_id"`
	ShipGroupID   uint
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
