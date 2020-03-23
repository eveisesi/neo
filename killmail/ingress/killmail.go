package ingress

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ddouglas/killboard"
	core "github.com/ddouglas/killboard/app"
	"github.com/ddouglas/killboard/esi"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Message struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
}

type Ingresser struct {
	*core.App
}

// var (
// 	mux sync.Mutex
// )

func Action(c *cli.Context) {
	channel := c.String("channel")

	i := &Ingresser{
		core.New(),
	}

	// Limiter is a wrapper around our job thats track GoRoutines and makes sure we don't have tomany in flight at a time.
	// limit := limiter.NewConcurrencyLimiter(1)

	for {
		count, err := i.Redis.ZCount(channel, "-inf", "+inf").Result()
		if err != nil {
			i.Logger.WithError(err).Fatal("unable to determine count of channel")
		}

		if count == 0 {
			i.Logger.Info("message set is empty")
			time.Sleep(time.Second * 2)
			continue
		}

		// Hashes are on the set. Fetch them
		results, err := i.Redis.ZPopMax(channel, 10).Result()
		if err != nil {
			i.Logger.WithError(err).Fatal("unable to retrieve records from Redis")
		}

		for _, result := range results {
			message := result.Member.(string)

			// limit.Execute(func() {
			i.HandleMessage(message)
			// })
			// time.Sleep(100 * time.Second)
		}

	}

}

func (i *Ingresser) HandleMessage(msg string) {
	var response esi.Response
	payload := Message{}
	err := json.Unmarshal([]byte(msg), &payload)
	if err != nil {
		log.Fatal(err)
		return
	}

	i.Logger.Infof("Working Hash %s:%s", payload.ID, payload.Hash)
	attempts := 1
	for {
		if attempts > 3 {
			i.Logger.WithFields(logrus.Fields{
				"id":   payload.ID,
				"hash": payload.Hash,
			}).Error("all attempts exhausted. failed to fetch killmail")
		}
		// attempt to get killmail
		response, err = i.ESI.GetKillmailsKillmailIDKillmailHash(payload.ID, payload.Hash)
		if err != nil {
			i.Logger.WithFields(logrus.Fields{
				"id":   payload.ID,
				"hash": payload.Hash,
			}).WithError(err).Error("esi request failed")
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		i.Logger.WithFields(logrus.Fields{
			"code":     response.Code,
			"path":     response.Path,
			"sleep":    5,
			"attempts": attempts,
		}).Warn("bad response code from esi received")
		time.Sleep(5 * time.Second)
	}

	killmail := response.Data.(killboard.Killmail)

	_, err = i.GetSolarSystemByID(killmail.SolarSystemID)
	if err != nil {
		i.Logger.WithError(err).Error("failed to fetch solar system")
	}

	victim := killmail.Victim

	if victim.AllianceID.Valid {
		i.GetAllianceByID(victim.AllianceID.Uint64)
	}
	if victim.CharacterID.Valid {
		i.GetCharacterByID(victim.CharacterID.Uint64)
	}
	i.GetCorporationByID(victim.CorporationID)
	i.GetTypeByID(victim.ShipTypeID)

	if len(victim.Items) > 0 {
		i.HandleVictimItems(victim.Items)
	}
	for _, attacker := range killmail.Attackers {
		if attacker.AllianceID.Valid {
			i.GetAllianceByID(attacker.AllianceID.Uint64)
		}
		if attacker.CharacterID.Valid {
			i.GetCharacterByID(attacker.CharacterID.Uint64)
		}
		i.GetCorporationByID(attacker.CorporationID)
		if attacker.ShipTypeID.Valid {
			i.GetTypeByID(attacker.ShipTypeID.Uint64)
		}
		if attacker.WeaponTypeID.Valid {
			i.GetTypeByID(attacker.WeaponTypeID.Uint64)
		}
	}

}

func (i *Ingresser) HandleVictimItems(items []*killboard.KillmailItem) {
	for _, item := range items {
		i.GetTypeByID(item.ItemTypeID)
		if len(item.Items) > 0 {
			i.HandleVictimItems(item.Items)
		}
	}
}
