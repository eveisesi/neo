package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type KillmailRespository interface {
	Killmail(ctx context.Context, id uint64, hash string) (*Killmail, error)
	CreateKillmail(ctx context.Context, killmail *Killmail) (*Killmail, error)
	CreateKillmailTxn(ctx context.Context, txn Transactioner, killmail *Killmail) (*Killmail, error)
	UpdateKillmail(ctx context.Context, id uint64, hash string, killmail *Killmail) error

	KillmailGTID(ctx context.Context, id null.Uint64) ([]*Killmail, error)
	KillmailExists(ctx context.Context, id uint64, hash string) (bool, error)
	KillmailRecent(ctx context.Context, page null.Int) ([]*Killmail, error)
	KillmailTop(ctx context.Context, age uint64, limit uint64) ([]*Killmail, error)
	KillmailsByCharacterID(ctx context.Context, id uint64) ([]*Killmail, error)
	KillmailsByCorporationID(ctx context.Context, id uint64) ([]*Killmail, error)
	KillmailsByAllianceID(ctx context.Context, id uint64) ([]*Killmail, error)

	KillmailAttackersByKillmailID(ctx context.Context, id uint64) ([]*KillmailAttacker, error)
	KillmailAttackersByKillmailIDs(ctx context.Context, ids []uint64) ([]*KillmailAttacker, error)
	CreateKillmailAttacker(ctx context.Context, attacker *KillmailAttacker) (*KillmailAttacker, error)
	CreateKillmailAttackerTxn(ctx context.Context, txn Transactioner, attacker *KillmailAttacker) (*KillmailAttacker, error)
	CreateKillmailAttackers(ctx context.Context, attackers []*KillmailAttacker) ([]*KillmailAttacker, error)
	CreateKillmailAttackersTxn(ctx context.Context, txn Transactioner, attackers []*KillmailAttacker) ([]*KillmailAttacker, error)

	KillmailItemsByKillmailID(ctx context.Context, id uint64) ([]*KillmailItem, error)
	KillmailItemsByKillmailIDs(ctx context.Context, ids []uint64) ([]*KillmailItem, error)
	CreateKillmailItem(ctx context.Context, item *KillmailItem) (*KillmailItem, error)
	CreateKillmailItemTxn(ctx context.Context, txn Transactioner, item *KillmailItem) (*KillmailItem, error)
	CreateKillmailItems(ctx context.Context, items []*KillmailItem) ([]*KillmailItem, error)
	CreateKillmailItemsTxn(ctx context.Context, txn Transactioner, items []*KillmailItem) ([]*KillmailItem, error)
	UpdateKillmailItems(ctx context.Context, items []*KillmailItem) error
	UpdateKillmailItemsTxn(ctx context.Context, txn Transactioner, items []*KillmailItem) error

	KillmailVictimByKillmailID(ctx context.Context, id uint64) (*KillmailVictim, error)
	KillmailVictimsByKillmailIDs(ctx context.Context, ids []uint64) ([]*KillmailVictim, error)
	CreateKillmailVictim(ctx context.Context, attacker *KillmailVictim) (*KillmailVictim, error)
	CreateKillmailVictimTxn(ctx context.Context, txn Transactioner, victim *KillmailVictim) (*KillmailVictim, error)
	UpdateKillmailVictim(ctx context.Context, victim *KillmailVictim) error
	UpdateKillmailVictimTxn(ctx context.Context, txn Transactioner, victim *KillmailVictim) error
}

type Killmail struct {
	ID             uint64     `json:"id"`
	Hash           string     `json:"hash"`
	MoonID         null.Int64 `json:"moon_id,omitempty"`
	SolarSystemID  uint64     `json:"solar_system_id"`
	WarID          null.Int64 `json:"war_id,omitempty"`
	IsNPC          bool       `json:"isNPC"`
	IsAwox         bool       `json:"isAwox"`
	IsSolo         bool       `json:"isSolo"`
	DroppedValue   float64    `json:"droppedValue"`
	DestroyedValue float64    `json:"destroyedValue"`
	FittedValue    float64    `json:"fittedValue"`
	TotalValue     float64    `json:"totalValue"`
	KillmailTime   time.Time  `json:"killmail_time"`

	Attackers []*KillmailAttacker `json:"attackers"`
	Victim    *KillmailVictim     `json:"victim"`
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

	Items []*KillmailItem `json:"items"`
}

type KillmailPosition struct {
	X null.Float64 `json:"x"`
	Y null.Float64 `json:"y"`
	Z null.Float64 `json:"z"`
}
