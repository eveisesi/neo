package killboard

import "time"

type Killmail struct {
	Attackers     []*KillmailAttacker `json:"attackers"`
	KillmailID    int64               `db:"killmail_id" json:"killmail_id"`
	KillmailTime  time.Time           `db:"killmail_time" json:"killmail_time"`
	MoonID        int64               `db:"moon_id" json:"moon_id"`
	SolarSystemID int64               `db:"solar_system_id" json:"solar_system_id"`
	Victim        *KillmailVictim     `json:"victim"`
	WarID         int64               `json:"war_id"`
}

type KillmailVictim struct {
	AllianceID    int64                 `db:"alliance_id" json:"allianceID"`
	CharacterID   int64                 `db:"character_id" json:"character_id"`
	CorporationID int64                 `db:"corporation_id" json:"corporation_id"`
	FactionID     int64                 `db:"faction_id" json:"faction_id"`
	DamageTaken   int64                 `db:"damage_taken" json:"damage_taken"`
	ShipTypeID    int64                 `db:"ship_type_id" json:"ship_type_id"`
	PosX          float64               `db:"pos_x" json:"pos_x"`
	PosY          float64               `db:"pos_y" json:"pos_y"`
	PosZ          float64               `db:"pos_z" json:"pos_z"`
	Items         []*KillmailVictimItem `json:"items"`
}

type KillmailVictimItem struct {
	Flag              int64                 `db:"flag_id" json:"flag_id"`
	ItemTypeID        int64                 `db:"item_type_id" json:"item_type_id"`
	QuantityDestroyed int64                 `db:"quantity_destroyed" json:"quantity_destroyed"`
	QuantityDropped   int64                 `db:"quantity_dropped" json:"quantity_dropped"`
	Singleton         bool                  `db:"singleton" json:"singleton"`
	HasItems          bool                  `db:"has_items" json:"has_items"`
	Items             []*KillmailVictimItem `json:"items"`
}

type KillmailAttacker struct {
	AllianceID     int64   `db:"alliance_id" json:"allianceID"`
	CharacterID    int64   `db:"character_id" json:"character_id"`
	CorporationID  int64   `db:"corporation_id" json:"corporation_id"`
	DamageDone     int64   `db:"damage_done" json:"damage_done"`
	FactionID      int64   `db:"faction_id" json:"faction_id"`
	FinalBlow      bool    `db:"final_blow" json:"final_blow"`
	SecurityStatus float64 `db:"security_status" json:"security_status"`
	ShipTypeID     int64   `db:"ship_type_id" json:"ship_type_id"`
	WeaponTypeID   int64   `db:"weapon_type_id" json:"weapon_type_id"`
}
