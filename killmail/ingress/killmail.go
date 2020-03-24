package ingress

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/korovkin/limiter"
	"github.com/volatiletech/sqlboiler/boil"

	"github.com/ddouglas/killboard"
	core "github.com/ddouglas/killboard/app"
	"github.com/ddouglas/killboard/esi"
	"github.com/ddouglas/killboard/mysql/boiler"
	"github.com/jinzhu/copier"
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

func Action(c *cli.Context) {
	channel := c.String("channel")

	i := &Ingresser{
		core.New(),
	}

	// Limiter is a wrapper around our job thats track GoRoutines and makes sure we don't have tomany in flight at a time.
	limit := limiter.NewConcurrencyLimiter(50)

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
			i.SleepDuringDowntime(time.Now())
			message := result.Member.(string)
			limit.ExecuteWithTicket(func(workerID int) {
				i.HandleMessage(message, workerID)
			})
			time.Sleep(time.Millisecond * 20)
		}

	}

}

// SleepDuringDowntime triggers a time.Sleep if the current server time falls within
// a Predetermined time range of Eve Onlines Downtime during which the API Server are
// unreachable
func (i *Ingresser) SleepDuringDowntime(t time.Time) {

	const lower = 39300 // 10:55 UTC 10h * 3600s + 55m * 60s
	const upper = 42900 // 11:25 UTC 11h * 3600s + 25m * 60s

	hm := (t.Hour() * 3600) + (t.Minute() * 60)

	if hm >= lower && hm <= upper {
		i.Logger.Info("Entering Sleep Phase for Downtime")
		time.Sleep(time.Second * time.Duration(upper-hm))
	}
}

func (i *Ingresser) HandleMessage(msg string, workerID int) {
	var response esi.Response
	payload := Message{}
	err := json.Unmarshal([]byte(msg), &payload)
	if err != nil {
		log.Fatal(err)
		return
	}

	killmailID, err := strconv.ParseUint(payload.ID, 10, 64)
	if err != nil {
		i.Logger.WithError(err).WithFields(logrus.Fields{
			"id":   payload.ID,
			"hash": payload.Hash,
		}).Error("unable to parse killmail id to uint")
		return
	}

	exists, err := boiler.KillmailExists(context.Background(), i.DB, killmailID, payload.Hash)
	if err != nil {
		i.Logger.WithError(err).WithFields(logrus.Fields{
			"id":   payload.ID,
			"hash": payload.Hash,
		}).Error("encountered checking if killmail exists")
		return
	}

	if exists {
		i.Logger.WithFields(logrus.Fields{
			"id":     payload.ID,
			"hash":   payload.Hash,
			"worker": workerID,
		}).Info("killmail successfully ingested")
		return
	}

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
			"id":       payload.ID,
			"hash":     payload.Hash,
		}).Error("bad response code from esi received")
		time.Sleep(5 * time.Second)
	}

	killmail := response.Data.(killboard.Killmail)

	_, err = i.GetSolarSystemByID(killmail.SolarSystemID)
	if err != nil {
		i.Logger.WithError(err).Error("failed to fetch solar system")
		return
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
		if attacker.CorporationID.Valid {
			i.GetCorporationByID(attacker.CorporationID.Uint64)
		}
		if attacker.ShipTypeID.Valid {
			i.GetTypeByID(attacker.ShipTypeID.Uint64)
		}
		if attacker.WeaponTypeID.Valid {
			i.GetTypeByID(attacker.WeaponTypeID.Uint64)
		}
	}

	// START A FUCKING TRANSACTION
	dbTx, err := i.DB.BeginTxx(context.Background(), nil)
	if err != nil {
		i.Logger.WithError(err).Fatal("unable to start db transaction")
		return
	}

	var boilKillmail = new(boiler.Killmail)
	err = copier.Copy(boilKillmail, &killmail)
	if err != nil {
		i.Logger.WithError(err).Error("unable to copy killmail to boiler.Killmail")
		return
	}
	boilKillmail.Hash = payload.Hash

	err = boilKillmail.Insert(context.Background(), dbTx, boil.Infer())
	if err != nil {
		i.Logger.WithError(err).Error("failed to insert killmail into database")
		return
	}

	if victim.Position != nil {
		position := victim.Position
		victim.PosX.SetValid(position.X.Float64)
		victim.PosY.SetValid(position.Y.Float64)
		victim.PosZ.SetValid(position.Z.Float64)
	}

	var boilVictim = new(boiler.KillmailVictim)
	err = copier.Copy(boilVictim, victim)
	if err != nil {
		i.Logger.WithError(err).Error("unable to copy killmail to boiler.Killmail")
		return
	}

	boilVictim.KillmailID = boilKillmail.ID

	err = boilVictim.Insert(context.Background(), dbTx, boil.Infer())
	if err != nil {
		i.Logger.WithError(err).Error("failed to insert killmail victim into database")
		return
	}

	for _, item := range victim.Items {
		var boilItem = new(boiler.KillmailItem)
		err = copier.Copy(boilItem, item)
		if err != nil {
			i.Logger.WithError(err).Error("failed to copy killmail item into boil killmailItem")
			continue
		}

		boilItem.KillmailID = boilKillmail.ID

		if len(item.Items) > 0 {
			boilItem.IsParent = true
		}

		err = boilItem.Insert(context.Background(), dbTx, boil.Infer())
		if err != nil {
			i.Logger.WithField("id", payload.ID).WithError(err).Error("failed to insert killmail item into database")
			break
		}

		if len(item.Items) > 0 {
			for _, subItem := range item.Items {

				var boilSubItem = new(boiler.KillmailItem)
				err = copier.Copy(boilSubItem, subItem)
				if err != nil {
					i.Logger.WithError(err).Error("failed to copy killmail item into boil killmailItem")
					continue
				}
				boilSubItem.KillmailID = boilKillmail.ID
				boilSubItem.ParentID.SetValid(boilItem.ID)
				err = boilSubItem.Insert(context.Background(), dbTx, boil.Infer())
				if err != nil {
					i.Logger.WithField("id", payload.ID).WithError(err).Error("failed to insert killmail item into database")
					break
				}

			}
		}
	}

	for _, attacker := range killmail.Attackers {
		var boilAttacker = new(boiler.KillmailAttacker)
		err = copier.Copy(boilAttacker, attacker)
		if err != nil {
			i.Logger.WithError(err).Error("failed to copy killmail attacker into boil killmailAttacker")
			continue
		}

		boilAttacker.KillmailID = boilKillmail.ID
		err = boilAttacker.Insert(context.Background(), dbTx, boil.Infer())
		if err != nil {
			i.Logger.WithError(err).Error("failed to insert killmail attacker into database")
			continue
		}
	}

	err = dbTx.Commit()
	if err != nil {
		err = dbTx.Rollback()
		if err != nil {
			i.Logger.WithError(err).Fatal("unable to rollback failed commit")
			return
		}
		i.Logger.WithError(err).Error("commit failed, successfully rollbacked")
		return
	}

	i.Logger.WithFields(logrus.Fields{
		"id":        boilKillmail.ID,
		"hash":      boilKillmail.Hash,
		"worker":    workerID,
		"numGoRout": runtime.NumGoroutine(),
	}).Info("killmail successfully ingested")

}

func (i *Ingresser) HandleVictimItems(items []*killboard.KillmailItem) {
	for _, item := range items {
		_, err := i.GetTypeByID(item.ItemTypeID)
		if err != nil {
			i.Logger.WithError(err).WithFields(logrus.Fields{
				"item_type_id": item.ItemTypeID,
				"func":         "HandleVictimItems",
			}).Error("encountered error")
		}
		if len(item.Items) > 0 {
			i.HandleVictimItems(item.Items)
		}
	}
}
