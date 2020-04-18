package killmail

import (
	"context"
	"encoding/json"
	"runtime"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/korovkin/limiter"
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

	limit := limiter.NewConcurrencyLimiter(int(gLimit))

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

		results, err := s.redis.ZPopMax(channel, gLimit).Result()
		if err != nil {
			s.logger.WithError(err).Fatal("unable to retrieve hashes from queue")
		}

		for _, result := range results {
			s.sleepDuringDowntime(time.Now())
			message := result.Member.(string)
			limit.ExecuteWithTicket(func(workerID int) {
				s.processMessage([]byte(message), workerID)
			})
			time.Sleep(time.Millisecond * time.Duration(gSleep))
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

	s.logger.WithFields(killmailLoggerFields).Info("received message")

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

	killmail, err := s.Killmail(ctx, payload.ID, payload.Hash, true, true)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to fetch killmail")
	}

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

	for _, item := range killmail.Victim.Items {
		item.KillmailID = killmail.ID
	}

	_, err = s.KillmailRespository.CreateKillmailItemsTxn(ctx, txn, killmail.Victim.Items)

	for _, item := range killmail.Victim.Items {
		if len(item.Items) > 0 {
			for _, subItem := range item.Items {
				subItem.KillmailID = killmailID
				subItem.ParentID.SetValid(item.ID)
			}
			_, err = s.KillmailRespository.CreateKillmailItemsTxn(ctx, txn, item.Items)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered insert sub items")
			}
		}
	}

	err = txn.Commit()
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to commit transaction")
		return
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
