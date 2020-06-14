package killmail

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/korovkin/limiter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *service) Importer(gLimit, gSleep int64) error {

	limit := limiter.NewConcurrencyLimiter(int(gLimit))

	for {
		count, err := s.redis.ZCount(neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
		if err != nil {
			s.logger.WithError(err).Error("unable to determine count of message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count == 0 {
			s.logger.Info("message queue is empty")
			time.Sleep(time.Second * 2)
			continue
		}

		results, err := s.redis.ZPopMin(neo.QUEUES_KILLMAIL_PROCESSING, gLimit).Result()
		if err != nil {
			s.logger.WithError(err).Fatal("unable to retrieve hashes from queue")
		}

		for _, result := range results {
			s.tracker.GateKeeper()
			message := result.Member.(string)
			limit.ExecuteWithTicket(func(workerID int) {
				s.processMessage([]byte(message), workerID, gSleep)
			})
		}
	}
}

func (s *service) processMessage(message []byte, workerID int, sleep int64) {

	var ctx = context.Background()

	var payload Message
	err := json.Unmarshal(message, &payload)
	if err != nil {
		s.logger.WithField("message", string(message)).Fatal("failed to unmarhal message into message struct")
	}

	killmailLoggerFields := logrus.Fields{
		"id":     payload.ID,
		"hash":   payload.Hash,
		"worker": workerID,
	}

	killmailID, err := strconv.ParseUint(payload.ID, 10, 64)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).Error("unable to parse killmail id to uint")
		return
	}

	exists, err := s.killmails.Exists(ctx, killmailID, payload.Hash)
	if err != nil {
		s.logger.WithError(err).
			WithFields(killmailLoggerFields).Error("error encountered checking if killmail exists")
	}

	if exists {
		s.logger.WithFields(killmailLoggerFields).Info("skipping existing killmail")
		return
	}

	killmail, m := s.esi.GetKillmailsKillmailIDKillmailHash(payload.ID, payload.Hash)
	if m.IsError() {
		s.logger.WithError(m.Msg).WithFields(killmailLoggerFields).WithFields(logrus.Fields{
			"code":  m.Code,
			"path":  m.Path,
			"query": m.Query,
		}).Error("failed to fetch killmail from esi")
		s.redis.ZAdd(neo.QUEUES_KILLMAIL_PROCESSING, &redis.Z{Score: 0, Member: message})
		return
	}

	if m.Code != 200 {
		s.logger.WithFields(killmailLoggerFields).WithFields(logrus.Fields{
			"code":  m.Code,
			"path":  m.Path,
			"query": m.Query,
		}).WithError(err).Error("unexpected response code from esi")
		s.redis.ZAdd(neo.QUEUES_KILLMAIL_PROCESSING, &redis.Z{Score: 0, Member: message})
		return
	}

	killmailLoggerFields["killTime"] = killmail.KillmailTime.Format("2006-01-02 15:04:05")

	killmail.Hash = payload.Hash

	s.primeKillmailNodes(ctx, killmail, killmailLoggerFields)

	txn, err := s.txn.Begin()
	if err != nil {
		s.logger.WithError(err).Error("failed to start transaction")
		return
	}

	_, err = s.killmails.CreateWithTxn(ctx, txn, killmail)
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

	date := killmail.KillmailTime

	var totalValue = make([]float64, 0)
	shipValue := s.market.FetchTypePrice(killmail.Victim.ShipTypeID, date)
	killmail.Victim.ShipValue = shipValue

	_, err = s.victim.CreateWithTxn(ctx, txn, killmail.Victim)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered inserting killmail victim into db")
	}

	for _, attacker := range killmail.Attackers {
		attacker.KillmailID = killmailID
	}

	_, err = s.attackers.CreateBulkWithTxn(ctx, txn, killmail.Attackers)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered inserting killmail attackers into db")
	}

	for _, item := range killmail.Victim.Items {
		item.KillmailID = killmail.ID
		itemValue := float64(0)
		if item.Singleton != 2 {
			itemValue = s.market.FetchTypePrice(item.ItemTypeID, date)
		} else {
			itemValue = 0.01
		}

		quantity := item.QuantityDestroyed.Uint64 + item.QuantityDropped.Uint64
		totalValue = append(totalValue, itemValue*float64(quantity))

		item.ItemValue = itemValue
	}

	_, err = s.items.CreateBulkWithTxn(ctx, txn, killmail.Victim.Items)
	if err != nil {
		s.logger.WithError(err).Error("failed to insert items into db")
	}

	destroyedValue := float64(0)
	droppedValue := float64(0)

	for _, item := range killmail.Victim.Items {
		if len(item.Items) > 0 {
			for _, subItem := range item.Items {
				subItem.KillmailID = killmailID
				subItem.ParentID.SetValid(item.ID)
				subItemValue := float64(0)
				if subItem.Singleton != 2 {
					subItemValue = s.market.FetchTypePrice(subItem.ItemTypeID, date)
				} else {
					subItemValue = 0.01
				}
				quantity := subItem.QuantityDestroyed.Uint64 + subItem.QuantityDropped.Uint64
				itemTotal := subItemValue * float64(quantity)
				totalValue = append(totalValue, itemTotal)
				subItem.ItemValue = subItemValue

				if subItem.QuantityDestroyed.Uint64 > 0 {
					destroyedValue += itemTotal
				}
				if subItem.QuantityDropped.Uint64 > 0 {
					droppedValue += itemTotal
				}
			}

			_, err = s.items.CreateBulkWithTxn(ctx, txn, item.Items)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered insert sub items")
			}
		}
		if item.QuantityDestroyed.Uint64 > 0 {
			destroyedValue += item.ItemValue * float64(item.QuantityDestroyed.Uint64)
		}
		if item.QuantityDropped.Uint64 > 0 {
			droppedValue += item.ItemValue * float64(item.QuantityDropped.Uint64)
		}
	}

	fittedValue := s.calculatedFittedValue(killmail.Victim.Items)
	fittedValue += shipValue
	totalValue = append(totalValue, shipValue)

	sum := float64(0)
	for _, v := range totalValue {
		sum += v
	}

	err = txn.Commit()
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to commit transaction")
		return
	}

	killmail.IsAwox = s.calcIsAwox(ctx, killmail)
	killmail.IsNPC = s.calcIsNPC(ctx, killmail)
	killmail.IsSolo = s.calcIsSolo(ctx, killmail)
	killmail.DestroyedValue = destroyedValue
	killmail.DroppedValue = droppedValue
	killmail.FittedValue = fittedValue
	killmail.TotalValue = sum

	err = s.killmails.Update(ctx, killmail.ID, killmail.Hash, killmail)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("error encountered inserting killmail victim into db")
	}

	s.logger.WithFields(killmailLoggerFields).Info("killmail successfully imported")

	if s.config.SlackNotifierEnabled {
		threshold := s.config.SlackNotifierValueThreshold * 1000000
		if killmail.TotalValue >= float64(threshold) {
			bytes, _ := json.Marshal(struct {
				ID   uint64 `json:"id"`
				Hash string `json:"hash"`
			}{
				ID:   killmail.ID,
				Hash: killmail.Hash,
			})

			_, err = s.redis.Publish(neo.REDIS_NOTIFICATION_PUBSUB, bytes).Result()
			if err != nil {
				s.logger.WithError(err).Error("failed to publish message")
			}
		}
	}

	time.Sleep(time.Millisecond * time.Duration(sleep))
}

func (s *service) RecalculatorDispatcher(limit int64, otherlimit int64) {

	for {

		count, err := s.redis.ZCount(neo.QUEUE_KILLMAIL_RECALCULATE, "-inf", "+inf").Result()
		if err != nil {
			s.logger.WithError(err).Error("unable to determine count of recalculation message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count > otherlimit {
			time.Sleep(time.Second * 10)
			continue
		}

		s.logger.Info("fetching killmail to recalculate")

		killmails, err := s.killmails.Recalculable(context.Background(), int(limit))
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("failed to fetch killmails")
			return
		}

		if len(killmails) == 0 {
			s.logger.Info("no killmails returned")
			return
		}

		s.logger.WithField("killmails", len(killmails)).Info("killmails retrieved successfully")

		for _, killmail := range killmails {

			msg := Message{
				ID:   strconv.FormatUint(killmail.ID, 10),
				Hash: killmail.Hash,
			}

			payload, err := json.Marshal(msg)
			if err != nil {
				s.logger.WithError(err).Error("failed to marshal message for recalculator queue")
				continue
			}

			_, err = s.redis.ZAdd(neo.QUEUE_KILLMAIL_RECALCULATE, &redis.Z{Score: 1, Member: string(payload)}).Result()
			if err != nil {
				s.logger.WithError(err).WithField("payload", string(payload)).Error("unable to push killmail to recalculating queue")
				return
			}
		}

		s.logger.Info("killmails dispatched successfully")
	}

}

func (s *service) Recalculator(gLimit int64) {

	limit := limiter.NewConcurrencyLimiter(int(gLimit))

	attempts := 0
	for {
		count, err := s.redis.ZCount(neo.QUEUE_KILLMAIL_RECALCULATE, "-inf", "+inf").Result()
		if err != nil {
			s.logger.WithError(err).Error("unable to determine count of recalculation message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count == 0 {
			attempts++
			if attempts >= 10 {
				s.logger.Info("done with recalculation")
				break
			}
			s.logger.WithField("count", count).Info("no messages on queue")
			time.Sleep(time.Second * 10)
			continue
		}

		s.logger.WithField("messages", count).Info("processing recalculable messages")

		results, err := s.redis.ZPopMin(neo.QUEUE_KILLMAIL_RECALCULATE, 1000).Result()
		if err != nil {
			s.logger.WithError(err).Fatal("unable to retrieve hashes from queue")
		}

		for _, result := range results {
			message := result.Member.(string)
			limit.ExecuteWithTicket(func(workerID int) {
				s.recalculateKillmail([]byte(message), workerID)
			})
		}

	}

}

func (s *service) recalculateKillmail(message []byte, workerID int) {

	var ctx = context.Background()

	var payload Message

	err := json.Unmarshal(message, &payload)
	if err != nil {
		s.logger.WithField("message", string(message)).Error("failed to unmarshal message into message struct")
		return
	}

	entry := s.logger.WithFields(logrus.Fields{
		"id":   payload.ID,
		"hash": payload.Hash,
	})

	killmailID, err := strconv.ParseUint(payload.ID, 10, 64)
	if err != nil {
		entry.WithError(err).Error("unable to parse killmail id to uint")
		return
	}

	killmail, err := s.killmails.Killmail(ctx, killmailID, payload.Hash)
	if err != nil {
		entry.WithError(err).Error("unable to retreive killmail from db")
		return
	}

	killmail.Victim, err = s.victim.ByKillmailID(ctx, killmail.ID)
	if err != nil {
		entry.WithError(err).Error("encountered error fetching victim")
		return
	}

	killmail.Victim.Items, err = s.items.ByKillmailID(ctx, killmail.ID)
	if err != nil {
		entry.WithError(err).Error("encountered error fetching victim items")
		return
	}

	s.primeKillmailNodes(ctx, killmail, logrus.Fields{})

	txn, err := s.txn.Begin()
	if err != nil {
		entry.WithError(err).Error("failed to start transaction")
		return
	}

	var totalValue = make([]float64, 0)
	shipValue := s.market.FetchTypePrice(killmail.Victim.ShipTypeID, killmail.KillmailTime)
	killmail.Victim.ShipValue = shipValue

	err = s.victim.UpdateWithTxn(ctx, txn, killmail.Victim)
	if err != nil {
		rollErr := txn.Rollback()
		if err != nil {
			err = errors.Wrap(err, errors.Wrap(rollErr, "failed to rollback txn").Error())
		}
		entry.WithError(err).Error("failed to update killmail victim")
		return
	}
	totalValue = append(totalValue, shipValue)

	destroyedValue := float64(0)
	droppedValue := float64(0)

	for _, item := range killmail.Victim.Items {
		item.KillmailID = killmail.ID
		itemValue := float64(0)
		if item.Singleton != 2 {
			itemValue = s.market.FetchTypePrice(item.ItemTypeID, killmail.KillmailTime)
		} else {
			itemValue = 0.01
		}

		quantity := item.QuantityDestroyed.Uint64 + item.QuantityDropped.Uint64
		totalValue = append(totalValue, itemValue*float64(quantity))
		if item.QuantityDestroyed.Uint64 > 0 {
			destroyedValue += itemValue * float64(item.QuantityDestroyed.Uint64)
		}
		if item.QuantityDropped.Uint64 > 0 {
			droppedValue += itemValue * float64(item.QuantityDropped.Uint64)
		}
		item.ItemValue = itemValue

	}

	err = s.items.UpdateBulkWithTxn(ctx, txn, killmail.Victim.Items)
	if err != nil {
		rollErr := txn.Rollback()
		if err != nil {
			err = errors.Wrap(err, errors.Wrap(rollErr, "failed to rollback txn").Error())
		}
		entry.WithError(err).Error("failed to update killmail victim items first level")
		return
	}

	for _, item := range killmail.Victim.Items {
		if len(item.Items) > 0 {
			for _, subItem := range item.Items {
				subItemValue := float64(0)
				if item.Singleton != 2 {
					subItemValue = s.market.FetchTypePrice(subItem.ItemTypeID, killmail.KillmailTime)
				} else {
					subItemValue = 0.01
				}

				quantity := subItem.QuantityDestroyed.Uint64 + subItem.QuantityDropped.Uint64
				itemTotal := subItemValue * float64(quantity)
				totalValue = append(totalValue, itemTotal)
				subItem.ItemValue = subItemValue

				if subItem.QuantityDestroyed.Uint64 > 0 {
					destroyedValue += itemTotal
				}
				if subItem.QuantityDropped.Uint64 > 0 {
					droppedValue += itemTotal
				}
			}

			err = s.items.UpdateBulkWithTxn(ctx, txn, item.Items)
			if err != nil {
				rollErr := txn.Rollback()
				if err != nil {
					err = errors.Wrap(err, errors.Wrap(rollErr, "failed to rollback txn").Error())
				}
				entry.WithError(err).WithField("parent_id", item.ID).Error("failed to update killmail victim items nested level")
				return
			}
		}
	}

	fittedValue := s.calculatedFittedValue(killmail.Victim.Items)
	fittedValue += shipValue
	destroyedValue += shipValue
	sum := float64(0)
	for _, v := range totalValue {
		sum += v
	}

	killmail.DestroyedValue = destroyedValue
	killmail.DroppedValue = droppedValue
	killmail.FittedValue = fittedValue
	killmail.TotalValue = sum

	err = s.killmails.UpdateWithTxn(ctx, txn, killmail)
	if err != nil {
		rollErr := txn.Rollback()
		if err != nil {
			err = errors.Wrap(err, errors.Wrap(rollErr, "failed to rollback txn").Error())
		}
		entry.WithError(err).Error("failed to update killmail")

		return
	}

	err = txn.Commit()
	if err != nil {
		entry.WithError(err).Error("failed to commit txn")
		return
	}

}

func (s *service) calcIsAwox(ctx context.Context, killmail *neo.Killmail) bool {

	if killmail.Victim == nil {
		return false
	}
	if killmail.Attackers == nil {
		return false
	}

	if !killmail.Victim.CorporationID.Valid {
		return false
	}

	//
	victimCorporationID := killmail.Victim.CorporationID.Uint64
	// Victim is in an NPC Corp. This is not an AWOX since characters
	// cannot choose which NPC Corp they are in
	if victimCorporationID < 98000000 {
		return false
	}

	shipType, err := s.universe.Type(ctx, killmail.Victim.ShipTypeID)
	if err != nil {
		return false
	}

	switch shipType.GroupID {
	// Capsule, Shuttle, Corvette, Citizen Ships
	case 29, 31, 237, 361, 2001:
		return false
	}

	for _, attacker := range killmail.Attackers {
		if !attacker.CorporationID.Valid {
			continue
		}

		if !attacker.ShipTypeID.Valid {
			continue
		}

		attackerShip, err := s.universe.Type(context.Background(), attacker.ShipTypeID.Uint64)
		if err != nil {
			continue
		}

		switch attackerShip.GroupID {
		// Capsule, Shuttle, Corvette, Citizen Ships
		case 29, 31, 237, 361, 2001:
			goto LoopEnd
		}

		if attacker.CorporationID.Uint64 == victimCorporationID {
			return true
		}
	LoopEnd:
	}

	return false

}

func (s *service) calcIsNPC(ctx context.Context, killmail *neo.Killmail) bool {

	if killmail.Victim == nil {
		return false
	}

	if killmail.Attackers == nil {
		return false
	}

	for _, attacker := range killmail.Attackers {
		if !attacker.CorporationID.Valid {
			continue
		}

		if attacker.CorporationID.Uint64 >= 98000000 {
			return false
		}
	}

	return true

}

func (s *service) calcIsSolo(ctx context.Context, killmail *neo.Killmail) bool {

	if killmail.Victim == nil {
		return false
	}

	if killmail.Attackers == nil {
		return false
	}

	if len(killmail.Attackers) > 1 {
		return false
	}

	attacker := killmail.Attackers[0]

	return attacker.CorporationID.Uint64 >= 98000000

}

func (s *service) primeKillmailNodes(ctx context.Context, killmail *neo.Killmail, killmailLoggerFields logrus.Fields) {
	_, err = s.universe.SolarSystem(ctx, killmail.SolarSystemID)
	if err != nil {
		s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime solar system")
	}

	victim := killmail.Victim

	if victim != nil {
		if victim.AllianceID.Valid {
			_, err := s.alliance.Alliance(ctx, victim.AllianceID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime victim alliance")
			}
		}

		if victim.CorporationID.Valid {
			_, err := s.corporation.Corporation(ctx, victim.CorporationID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime victim character")
			}
		}

		if victim.CharacterID.Valid {
			_, err := s.character.Character(ctx, victim.CharacterID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime victim character")
			}
		}

		_, err = s.universe.Type(ctx, victim.ShipTypeID)
		if err != nil {
			s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime victim ship type")
		}
	}

	for _, attacker := range killmail.Attackers {
		if attacker.AllianceID.Valid {
			_, err := s.alliance.Alliance(ctx, attacker.AllianceID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime attacker alliance")
			}
		}

		if attacker.CorporationID.Valid {
			_, err := s.corporation.Corporation(ctx, attacker.CorporationID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime attacker character")
			}
		}

		if attacker.CharacterID.Valid {
			_, err := s.character.Character(ctx, attacker.CharacterID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime attacker character")
			}
		}

		if attacker.ShipTypeID.Valid {
			_, err = s.universe.Type(ctx, attacker.ShipTypeID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime attacker ship type")
			}
		}

		if attacker.WeaponTypeID.Valid {
			_, err = s.universe.Type(ctx, attacker.WeaponTypeID.Uint64)
			if err != nil {
				s.logger.WithFields(killmailLoggerFields).WithError(err).Error("failed to prime attacker ship type")
			}
		}
	}

	s.handleVictimItems(killmail.Victim.Items)
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

var fittedFlags = map[uint64]bool{
	// LoSlots
	11: true, 12: true, 13: true, 14: true, 15: true, 16: true, 17: true, 18: true,
	// MedSlot
	19: true, 20: true, 21: true, 22: true, 23: true, 24: true, 25: true, 26: true,
	// HiSlot
	27: true, 28: true, 29: true, 30: true, 31: true, 32: true, 33: true, 34: true,
	// Drone Bay
	87: true,
	// Implants
	89: true,
	// Rig Slots
	92: true, 93: true, 94: true, 95: true, 96: true, 97: true, 98: true, 99: true,
	//  Subsystem Slots
	125: true, 126: true, 127: true, 128: true, 129: true, 130: true, 131: true, 132: true,
	// Fighter Tubes
	159: true, 160: true, 161: true, 162: true, 163: true,
	// Structure Service Slots
	164: true, 165: true, 166: true, 167: true, 168: true, 169: true, 170: true, 171: true,
}

func (s *service) calculatedFittedValue(items []*neo.KillmailItem) float64 {

	total := float64(0)
	for _, item := range items {
		if _, ok := fittedFlags[uint64(item.Flag)]; !ok {
			continue
		}
		total += item.ItemValue * float64(item.QuantityDestroyed.Uint64+item.QuantityDropped.Uint64)
	}

	return total
}

// func (s *service) Something() {

// 	query := "
// 	SELECT
// 		k.id,
// 		k.hash,
// 		k.destroyed_value,
// 		k.dropped_value,
// 		kv.ship_value,
// 		(k.destroyed_value + k.dropped_value + kv.ship_value) as calculated_value,
// 		k.total_value
// 	FROM killmails k
// 	LEFT JOIN killmail_victim kv ON (k.id = kv.killmail_id)
// 	WHERE total_value != 0 LIMIT 1000
// 	"

// }
