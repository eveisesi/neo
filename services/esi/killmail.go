package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eveisesi/neo"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type Killmail struct {
	ID            uint      `json:"killmail_id"`
	Hash          string    `json:"hash"`
	MoonID        *uint     `json:"moon_id"`
	SolarSystemID uint      `json:"solar_system_id"`
	WarID         *uint     `json:"war_id"`
	KillmailTime  time.Time `json:"killmail_time"`

	Attackers []*KillmailAttacker `json:"attackers"`
	Victim    *KillmailVictim     `json:"victim"`
}

type KillmailAttacker struct {
	AllianceID     *uint   `json:"alliance_id"`
	CharacterID    *uint   `json:"character_id"`
	CorporationID  *uint   `json:"corporation_id"`
	DamageDone     uint    `json:"damage_done"`
	FactionID      *uint   `json:"faction_id"`
	FinalBlow      bool    `json:"final_blow"`
	SecurityStatus float64 `json:"security_status"`
	ShipTypeID     *uint   `json:"ship_type_id"`
	WeaponTypeID   *uint   `json:"weapon_type_id"`
}

type KillmailVictim struct {
	AllianceID    *uint           `json:"alliance_id"`
	CharacterID   *uint           `json:"character_id"`
	CorporationID *uint           `json:"corporation_id"`
	DamageTaken   uint            `json:"damage_taken"`
	FactionID     *uint           `json:"faction_id"`
	ShipTypeID    uint            `json:"ship_type_id"`
	Position      *Position       `json:"position"`
	Items         []*KillmailItem `json:"items"`
}

type KillmailItem struct {
	Flag              uint            `json:"flag"`
	ItemTypeID        uint            `json:"item_type_id"`
	QuantityDropped   *uint           `json:"quantity_dropped"`
	QuantityDestroyed *uint           `json:"quantity_destroyed"`
	Singleton         uint8           `json:"singleton"`
	Items             []*KillmailItem `json:"items"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (s *service) GetKillmailsKillmailIDKillmailHash(ctx context.Context, id uint, hash string) (*neo.Killmail, Meta) {

	path := fmt.Sprintf("/v1/killmails/%d/%s/", id, hash)

	request := request{
		method: http.MethodGet,
		path:   path,
	}

	response, m := s.request(ctx, request)
	if m.IsErr() {
		return nil, m
	}

	esiKillmail := new(Killmail)

	err = json.Unmarshal(response, esiKillmail)
	if err != nil {
		m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
		return nil, m
	}

	esiKillmail.Hash = hash

	var killmail = new(neo.Killmail)
	err = copier.Copy(killmail, esiKillmail)
	if err != nil {
		m.Msg = err
		return nil, m
	}

	return killmail, m
}
