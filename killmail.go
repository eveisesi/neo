package killboard

import (
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/types"
)

type Killmail struct {
	ID            uint64     `json:"id"`
	Hash          string     `json:"hash"`
	MoonID        null.Int64 `json:"moon_id,omitempty"`
	SolarSystemID uint64     `json:"solar_system_id"`
	WarID         null.Int64 `json:"war_id,omitempty"`
	KillmailTime  time.Time  `json:"killmail_time"`

	Attackers []*KillmailAttacker `json:"attackers"`
	Victim    *KillmailVictim     `json:"victim"`
}

type KillmailAttacker struct {
	ID             uint64        `json:"id"`
	KillmailID     uint64        `json:"killmail_id"`
	AllianceID     null.Uint64   `json:"alliance_id"`
	CharacterID    null.Uint64   `json:"character_id"`
	CorporationID  uint64        `json:"corporation_id"`
	FactionID      null.Uint64   `json:"faction_id"`
	DamageDone     uint64        `json:"damage_done"`
	FinalBlow      bool          `json:"final_blow"`
	SecurityStatus types.Decimal `json:"security_status"`
	ShipTypeID     null.Uint64   `json:"ship_type_id"`
	WeaponTypeID   null.Uint64   `json:"weapon_type_id"`
}

type KillmailItem struct {
	ID                uint64      `json:"id"`
	ParentID          null.Uint64 `json:"parent_id"`
	KillmailID        uint64      `json:"killmail_id"`
	FlagID            uint64      `json:"flag_id"`
	ItemTypeID        uint64      `json:"item_type_id"`
	QuantityDropped   null.Uint64 `json:"quantity_dropped"`
	QuantityDestroyed null.Uint64 `json:"quantity_destroyed"`
	Singleton         uint64      `json:"singleton"`
	IsParent          int8        `json:"is_parent"`

	Items []*KillmailItem `json:"items"`
}

type KillmailVictim struct {
	ID            uint          `json:"id"`
	KillmailID    uint64        `json:"killmail_id"`
	AllianceID    null.Uint64   `json:"alliance_id"`
	CharacterID   null.Uint64   `json:"character_id"`
	CorporationID uint64        `json:"corporation_id"`
	FactionID     null.Uint64   `json:"faction_id"`
	DamageTaken   uint64        `json:"damage_taken"`
	ShipTypeID    uint64        `json:"ship_type_id"`
	PosX          types.Decimal `json:"pos_x"`
	PosY          types.Decimal `json:"pos_y"`
	PosZ          types.Decimal `json:"pos_z"`

	Items []*KillmailItem `json:"items"`
}
