package neo

import (
	"context"
	"time"
)

type KillmailRepository interface {
	Killmail(ctx context.Context, id uint) (*Killmail, error)
	Killmails(ctx context.Context, mods ...Modifier) ([]*Killmail, error)
	CreateKillmail(ctx context.Context, killmail *Killmail) error
	// Update(ctx context.Context, id uint, killmail *Killmail) error

	Exists(ctx context.Context, id uint) (bool, error)
	// Recalculable(ctx context.Context, limit int, after uint) ([]*Killmail, error)
}

type Killmail struct {
	ID              uint      `bson:"id" json:"id"`
	Hash            string    `bson:"hash" json:"hash"`
	MoonID          *uint     `bson:"moonID,omitempty" json:"moonID,omitempty"`
	SolarSystemID   uint      `bson:"solarSystemID" json:"solarSystemID"`
	ConstellationID uint      `bson:"constellationID"`
	RegionID        uint      `bson:"regionID"`
	WarID           *uint     `bson:"warID,omitempty" json:"warID,omitempty"`
	IsNPC           bool      `bson:"isNPC" json:"isNPC"`
	IsAwox          bool      `bson:"isAwox" json:"isAwox"`
	IsSolo          bool      `bson:"isSolo" json:"isSolo"`
	DroppedValue    float64   `bson:"droppedValue" json:"droppedValue"`
	DestroyedValue  float64   `bson:"destroyedValue" json:"destroyedValue"`
	FittedValue     float64   `bson:"fittedValue" json:"fittedValue"`
	TotalValue      float64   `bson:"totalValue" json:"totalValue"`
	KillmailTime    time.Time `bson:"killmailTime" json:"killmailTime"`

	System    *SolarSystem        `bson:"-" json:"-"`
	Attackers []*KillmailAttacker `bson:"attackers" json:"attackers"`
	Victim    *KillmailVictim     `bson:"victim" json:"victim"`
}

type KillmailAttacker struct {
	KillmailID     uint    `bson:"killmailID" json:"killmailID"`
	AllianceID     *uint   `bson:"allianceID" json:"allianceID"`
	CharacterID    *uint64 `bson:"characterID" json:"characterID"`
	CorporationID  *uint   `bson:"corporationID" json:"corporationID"`
	FactionID      *uint   `bson:"factionID" json:"factionID"`
	DamageDone     uint    `bson:"damageDone" json:"damageDone"`
	FinalBlow      bool    `bson:"finalBlow" json:"finalBlow"`
	SecurityStatus float64 `bson:"securityStatus" json:"securityStatus"`
	ShipTypeID     *uint   `bson:"shipTypeID" json:"shipTypeID"`
	ShipGroupID    *uint   `bson:"shipGroupID" json:"shipGroupID"`
	WeaponTypeID   *uint   `bson:"weaponTypeID" json:"weaponTypeID"`
	WeaponGroupID  *uint   `bson:"weaponGroupID" json:"weaponGroupID"`

	Alliance    *Alliance    `bson:"-" json:"-"`
	Character   *Character   `bson:"-" json:"-"`
	Corporation *Corporation `bson:"-" json:"-"`
	Ship        *Type        `bson:"-" json:"-"`
	Weapon      *Type        `bson:"-" json:"-"`
}

type KillmailItem struct {
	KillmailID        uint    `bson:"killmailID" json:"killmailID"`
	Flag              uint    `bson:"flag" json:"flag"`
	ItemTypeID        uint    `bson:"itemTypeID" json:"itemTypeID"`
	ItemGroupID       uint    `bson:"itemGroupID" json:"itemGroupID"`
	QuantityDropped   *uint   `bson:"quantityDropped" json:"quantityDropped"`
	QuantityDestroyed *uint   `bson:"quantityDestroyed" json:"quantityDestroyed"`
	ItemValue         float64 `bson:"itemValue" json:"itemValue"`
	TotalValue        float64 `bson:"totalValue" json:"totalValue"`
	Singleton         uint8   `bson:"singleton" json:"singleton"`
	IsParent          bool    `bson:"isparent" json:"isparent"`

	Type  *Type           `bson:"-" json:"-"`
	Items []*KillmailItem `bson:"items" json:"items"`
}

type KillmailVictim struct {
	KillmailID    uint    `bson:"killmailID" json:"killmailID"`
	AllianceID    *uint   `bson:"allianceID" json:"allianceID"`
	CharacterID   *uint64 `bson:"characterID" json:"characterID"`
	CorporationID *uint   `bson:"corporationID" json:"corporationID"`
	FactionID     *uint   `bson:"factionID" json:"factionID"`
	DamageTaken   uint    `bson:"damageTaken" json:"damageTaken"`
	ShipTypeID    uint    `bson:"shipTypeID" json:"shipTypeID"`
	ShipGroupID   uint    `bson:"shipGroupID" json:"shipGroupID"`
	ShipValue     float64 `bson:"shipValue" json:"shipValue"`

	Alliance    *Alliance    `bson:"-" json:"-"`
	Character   *Character   `bson:"-" json:"-"`
	Corporation *Corporation `bson:"-" json:"-"`
	Ship        *Type        `bson:"-" json:"-"`

	Position *Position       `bson:"position" json:"position"`
	Items    []*KillmailItem `bson:"items" json:"items"`
}

type Position struct {
	X float64 `bson:"x" json:"x"`
	Y float64 `bson:"y" json:"y"`
	Z float64 `bson:"z" json:"z"`
}
