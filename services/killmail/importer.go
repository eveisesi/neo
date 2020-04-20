package killmail

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/neo"
	"github.com/sirupsen/logrus"
)

// now := time.Now()
// location, _ := time.LoadLocation("UTC")

// // then := time.Until
// return nil

func (s *service) sleepDuringDowntime(now time.Time) {

	location, _ := time.LoadLocation("UTC")
	start := time.Date(now.Year(), now.Month(), now.Day(), 21, 26, 0, 0, location)
	end := time.Date(now.Year(), now.Month(), now.Day(), 21, 28, 0, 0, location)
	for {

		now := time.Now()
		if now.Unix() < start.Unix() || now.Unix() > end.Unix() {
			break
		}

		duration := (end.Unix() - now.Unix()) + 1
		s.logger.WithFields(logrus.Fields{
			"duration": duration,
			"end":      end.Unix(),
		}).Info("downtime period detected, sleeping")

		time.Sleep(time.Second * time.Duration(duration))
	}

	return
}

func (s *service) Importer(channel string, gLimit, gSleep int64) error {

	// limit := limiter.NewConcurrencyLimiter(int(gLimit))

	for {
		count, err := s.redis.ZCount(channel, "-inf", "+inf").Result()
		if err != nil {
			s.logger.WithError(err).Fatal("unable to determine count of message queue")
		}

		if count == 0 {
			s.logger.Info("message queue is empty")
			time.Sleep(time.Second * 2)
			continue
		}

		results, err := s.redis.ZPopMin(channel, gLimit).Result()
		if err != nil {
			s.logger.WithError(err).Fatal("unable to retrieve hashes from queue")
		}

		for _, result := range results {
			s.sleepDuringDowntime(time.Now())
			message := result.Member.(string)
			// limit.ExecuteWithTicket(func(workerID int) {
			s.processMessage([]byte(message), 1)
			// })
			// time.Sleep(time.Millisecond * time.Duration(gSleep))
		}
	}
}

func (s *service) processMessage(message []byte, workerID int) {

	var ctx = context.Background()

	var payload Message
	err := json.Unmarshal(message, &payload)
	if err != nil {
		s.logger.WithField("message", string(message)).Fatal("failed to unmarhal message into message struct")
	}

	killmailLoggerFields := logrus.Fields{
		"id":        payload.ID,
		"hash":      payload.Hash,
		"worker":    workerID,
		"numGoRout": runtime.NumGoroutine(),
	}

	killmailID, err := strconv.ParseUint(payload.ID, 10, 64)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).Error("unable to parse killmail id to uint")
		return
	}

	exists, err := s.KillmailExists(ctx, killmailID, payload.Hash)
	if err != nil {
		s.logger.WithError(err).
			WithFields(killmailLoggerFields).Error("error encountered checking if killmail exists")
	}

	if exists {
		s.logger.WithFields(killmailLoggerFields).Info("skipping existing killmail")
		return
	}

	res, err := s.esi.GetKillmailsKillmailIDKillmailHash(payload.ID, payload.Hash)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch killmail from esi")
		return
	}

	if res.Code != 200 {
		s.logger.WithFields(killmailLoggerFields).WithField("code", res.Code).WithError(err).Error("unexpected response code from esi")
		return
	}

	killmail := res.Data.(*neo.Killmail)
	killmail.Hash = payload.Hash

	_, err = s.universe.SolarSystem(ctx, killmail.SolarSystemID)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch solar system")
	}

	victim := killmail.Victim

	if victim.AllianceID.Valid {
		_, err := s.alliance.Alliance(ctx, victim.AllianceID.Uint64)
		if err != nil {
			s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch victim alliance")
		}
	}

	if victim.CorporationID.Valid {
		_, err := s.corporation.Corporation(ctx, victim.CorporationID.Uint64)
		if err != nil {
			s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch victim character")
		}
	}

	if victim.CharacterID.Valid {
		_, err := s.character.Character(ctx, victim.CharacterID.Uint64)
		if err != nil {
			s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch victim character")
		}
	}

	_, err = s.universe.Type(ctx, victim.ShipTypeID)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch victim ship type")
	}

	for _, attacker := range killmail.Attackers {
		if attacker.AllianceID.Valid {
			_, err := s.alliance.Alliance(ctx, attacker.AllianceID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch attacker alliance")
			}
		}

		if attacker.CorporationID.Valid {
			_, err := s.corporation.Corporation(ctx, attacker.CorporationID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch attacker character")
			}
		}

		if attacker.CharacterID.Valid {
			_, err := s.character.Character(ctx, attacker.CharacterID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch attacker character")
			}
		}

		if attacker.ShipTypeID.Valid {
			_, err = s.universe.Type(ctx, attacker.ShipTypeID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch attacker ship type")
			}
		}

		if attacker.WeaponTypeID.Valid {
			_, err = s.universe.Type(ctx, attacker.WeaponTypeID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch attacker ship type")
			}
		}
	}

	txn, err := s.txn.Begin()
	_, err = s.KillmailRespository.CreateKillmailTxn(ctx, txn, killmail)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered inserting killmail into db")
		return
	}

	killmail.Victim.KillmailID = killmail.ID

	if killmail.Victim.Position != nil {
		killmail.Victim.PosX.SetValid(killmail.Victim.Position.X.Float64)
		killmail.Victim.PosY.SetValid(killmail.Victim.Position.Y.Float64)
		killmail.Victim.PosZ.SetValid(killmail.Victim.Position.Z.Float64)
	}

	_, err = s.KillmailRespository.CreateKillmailVictimTxn(ctx, txn, killmail.Victim)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered inserting killmail victim into db")
	}

	for _, attacker := range killmail.Attackers {
		attacker.KillmailID = killmailID
	}

	_, err = s.KillmailRespository.CreateKillmailAttackersTxn(ctx, txn, killmail.Attackers)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered inserting killmail attackers into db")
	}

	minDate := time.Date(killmail.KillmailTime.Year(), killmail.KillmailTime.Month(), killmail.KillmailTime.Day(), 0, 0, 0, 0, time.UTC)
	maxDate := time.Date(killmail.KillmailTime.Year(), killmail.KillmailTime.Month(), killmail.KillmailTime.Day(), 23, 59, 59, 0, time.UTC)

	var totalValue = make([]float64, 0)
	fmt.Printf("Calculating Raw Material Cost for Ship %d\n", killmail.Victim.ShipTypeID)
	shipValue := append(totalValue, s.market.CalculateRawMaterialCost(killmail.Victim.ShipTypeID, minDate, maxDate))
	fmt.Printf("Done Calculating Ship. Item %d cost %.f to build\n", killmail.Victim.ShipTypeID, shipValue)
	totalValue = shipValue

	for _, item := range killmail.Victim.Items {
		item.KillmailID = killmail.ID
		fmt.Printf("Calculating Raw Material Cost for Item %d\n", item.ItemTypeID)
		itemValue := s.market.CalculateRawMaterialCost(item.ItemTypeID, minDate, maxDate)
		totalValue = append(totalValue, itemValue*float64(item.QuantityDestroyed.Uint64+item.QuantityDropped.Uint64))
		fmt.Printf("Done Calculating Item. Item %d cost %.f to build\n", item.ItemTypeID, itemValue)

		item.ItemValue = itemValue
	}

	_, err = s.KillmailRespository.CreateKillmailItemsTxn(ctx, txn, killmail.Victim.Items)

	for _, item := range killmail.Victim.Items {
		if len(item.Items) > 0 {
			for _, subItem := range item.Items {
				subItem.KillmailID = killmailID
				subItem.ParentID.SetValid(item.ID)

				fmt.Printf("Calculating Raw Material Cost for SubItem %d", subItem.ItemTypeID)
				subItemValue := s.market.CalculateRawMaterialCost(subItem.ItemTypeID, minDate, maxDate)
				totalValue = append(totalValue, subItemValue*float64(subItem.QuantityDestroyed.Uint64+subItem.QuantityDropped.Uint64))
				fmt.Printf("Done Calculating SubItem. Item %d cost %.f to build", subItem.ItemTypeID, subItemValue)

				subItem.ItemValue = subItemValue
			}
			_, err = s.KillmailRespository.CreateKillmailItemsTxn(ctx, txn, item.Items)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered insert sub items")
			}
		}
	}

	sum := float64(0)
	for _, v := range totalValue {
		sum += v
	}

	spew.Dump(sum)

	err = txn.Commit()
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to commit transaction")
		return
	}

	killmail.TotalValue = sum
	err = s.KillmailRespository.UpdateKillmail(ctx, killmail.ID, killmail.Hash, killmail)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered inserting killmail victim into db")
	}

	s.logger.WithFields(killmailLoggerFields).Info("killmail successfully imported")

}

func (s *service) handleVictimItems(items []*neo.KillmailItem) {
	for _, item := range items {
		_, err := s.universe.Type(context.Background(), item.ItemTypeID)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"item_type_id": item.ItemTypeID,
			}).Error("encountered error")
		}
		if len(item.Items) > 0 {
			s.handleVictimItems(item.Items)
		}
	}
}
