package killmail

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/korovkin/limiter"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *service) Importer(gLimit, gSleep int64) error {

	limit := limiter.NewConcurrencyLimiter(int(gLimit))
	for {

		txn := s.newrelic.StartTransaction("killmail queue check")
		ctx := newrelic.NewContext(context.Background(), txn)

		stop, err := s.redis.WithContext(ctx).Get(neo.QUEUE_STOP).Int64()
		if err != nil {
			s.logger.WithError(err).Error("stop flag is missing. attempting to create with default value of 0")
			_, err := s.redis.WithContext(ctx).Set(neo.QUEUE_STOP, 0, 0).Result()
			if err != nil {
				txn.NoticeError(err)
				s.logger.WithError(err).Fatal("error encountered attempting to create stop flag with default value")
			}
			continue
		}

		if stop == 1 {
			s.logger.Info("stop signal set")
			if limit.GetNumInProgress() > 0 {
				s.logger.Info("calling limit.Wait")

				limit.Wait()
			}

			s.logger.Info("sleeping for 5 seconds")
			time.Sleep(time.Second * 5)
			continue
		}

		count, err := s.redis.WithContext(ctx).ZCount(neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
		if err != nil {
			txn.NoticeError(err)
			s.logger.WithError(err).Error("unable to determine count of message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count == 0 {
			txn.Ignore()
			s.logger.Info("message queue is empty")
			time.Sleep(time.Second * 2)
			continue
		}

		results, err := s.redis.WithContext(ctx).ZPopMax(neo.QUEUES_KILLMAIL_PROCESSING, gLimit).Result()
		if err != nil {
			txn.NoticeError(err)
			s.logger.WithError(err).Fatal("unable to retrieve hashes from queue")
		}

		for i, result := range results {
		LoopStart:
			proceed := s.tracker.Watchman(ctx)
			if !proceed {
				time.Sleep(time.Second)
				goto LoopStart
			}
			message := result.Member.(string)
			limit.ExecuteWithTicket(func(workerID int) {
				s.handleMessage(ctx, []byte(message), i, gSleep)
			})
		}

		txn.End()
	}
}

func (s *service) handleMessage(ctx context.Context, message []byte, workerID int, sleep int64) {

	txn := s.newrelic.StartTransaction("handleMessage")
	defer txn.End()
	ctx = newrelic.NewContext(ctx, txn.NewGoroutine())

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"worker": workerID,
	})

	killmail, err := s.ProcessMessage(ctx, entry, message)
	if err != nil {
		txn.NoticeError(err)
		s.logger.WithContext(ctx).WithError(err).Error("failed to handle message")
		return
	}

	// Killmails we already know about an have processed come back as nil from the processor
	if killmail == nil {
		txn.Ignore()
		return
	}

	msg, err := json.Marshal(neo.Message{
		ID:   killmail.ID,
		Hash: killmail.Hash,
	})
	if err != nil {
		entry.WithError(err).Error("failed to marshal message for stats queue")
		return
	}
	member := &redis.Z{Score: float64(killmail.ID), Member: msg}
	if s.config.SlackNotifierEnabled {
		threshold := s.config.SlackNotifierValueThreshold * 1000000
		if killmail.TotalValue >= float64(threshold) {
			_, err = s.redis.WithContext(ctx).ZAdd(neo.QUEUES_KILLMAIL_NOTIFICATION, member).Result()
			if err != nil {
				entry.WithError(err).Error("failed to publish message to notifications queue")
			}
		}
	}

	time.Sleep(time.Millisecond * time.Duration(sleep))

}

func (s *service) ProcessMessage(ctx context.Context, entry *logrus.Entry, message []byte) (*neo.Killmail, error) {

	// var wg sync.WaitGroup

	txn := newrelic.FromContext(ctx)

	var payload = neo.Message{}
	err := json.Unmarshal(message, &payload)
	if err != nil {
		txn.NoticeError(err)
		s.logger.WithContext(ctx).WithField("message", string(message)).Error("failed to unmarhal message into message struct")
		return nil, err
	}

	entry = entry.WithFields(logrus.Fields{
		"id":   payload.ID,
		"hash": payload.Hash,
	})

	exists, err := s.killmails.Exists(ctx, payload.ID)
	if err != nil {
		txn.NoticeError(err)
		entry.WithError(err).
			Error("error encountered checking if killmail exists")
		return nil, err
	}
	if exists {
		entry.Info("skipping existing killmail")
		return nil, nil
	}

	killmail, m := s.esi.GetKillmailsKillmailIDKillmailHash(ctx, payload.ID, payload.Hash)
	if m.IsErr() {
		txn.NoticeError(m.Msg)
		entry.WithError(m.Msg).WithFields(logrus.Fields{
			"code":  m.Code,
			"path":  m.Path,
			"query": m.Query,
		}).Error("failed to fetch killmail from esi")
		if m.Code >= 500 {
			s.redis.ZAdd(neo.QUEUES_KILLMAIL_PROCESSING, &redis.Z{Score: float64(payload.ID), Member: message})
		}
		return nil, err
	}

	if m.Code != 200 {
		txn.NoticeError(errors.New("unexpected response code from esi"))
		entry.WithFields(logrus.Fields{
			"code":  m.Code,
			"path":  m.Path,
			"query": m.Query,
		}).Error("unexpected response code from esi")

		if m.Code == http.StatusUnprocessableEntity {
			s.redis.WithContext(ctx).ZAdd(neo.ZKB_INVALID_HASH, &redis.Z{Score: float64(payload.ID), Member: message})
			return nil, err
		}
		s.redis.WithContext(ctx).ZAdd(neo.QUEUES_KILLMAIL_PROCESSING, &redis.Z{Score: float64(payload.ID), Member: message})
		return nil, err
	}

	// wg.Add(1)
	// go s.backupKillmail(ctx, &wg, killmail.KillmailTime, payload, m.Data)

	entry = entry.WithField("killtime", killmail.KillmailTime.Format("2006-01-02 15:04:05"))
	s.primeKillmailNodes(ctx, killmail, entry)

	killmail.Victim.KillmailID = killmail.ID

	date := killmail.KillmailTime
	shipValue := s.market.FetchTypePrice(killmail.Victim.ShipTypeID, date)
	killmail.Victim.ShipValue = shipValue

	victimShipType, err := s.universe.Type(ctx, killmail.Victim.ShipTypeID)
	if err != nil {
		entry.WithError(err).Error("error encountered looking up type information for victim ship")
	}

	killmail.Victim.ShipGroupID = victimShipType.GroupID

	for _, attacker := range killmail.Attackers {
		attacker.KillmailID = killmail.ID

		if attacker.ShipTypeID != nil {
			attackerShipType, err := s.universe.Type(ctx, *attacker.ShipTypeID)
			if err != nil {
				entry.WithError(err).WithField("ship_type_id", attacker.ShipTypeID).Error("encountered looking up type information for attacker ship")
				continue
			}

			attacker.ShipGroupID = &attackerShipType.GroupID
		}

		if attacker.WeaponTypeID != nil {
			attackerWeaponType, err := s.universe.Type(ctx, *attacker.WeaponTypeID)
			if err != nil {
				entry.WithError(err).WithField("weapon_type_id", attacker.WeaponTypeID).Error("encountered looking up type information for attacker weapon")
				continue
			}

			attacker.WeaponGroupID = &attackerWeaponType.GroupID
		}
	}

	destroyedValue := float64(0)
	droppedValue := float64(0)

	for _, item := range killmail.Victim.Items {
		item.KillmailID = killmail.ID
		itemValue := float64(0)
		if item.Singleton != 2 {
			itemValue = s.market.FetchTypePrice(item.ItemTypeID, date)
		} else {
			itemValue = 0.01
		}

		item.ItemValue = itemValue
		if item.QuantityDestroyed != nil && *item.QuantityDestroyed > 0 {
			destroyedValue += item.ItemValue * float64(*item.QuantityDestroyed)
		}
		if item.QuantityDropped != nil && *item.QuantityDropped > 0 {
			droppedValue += item.ItemValue * float64(*item.QuantityDropped)
		}
		if len(item.Items) > 0 {
			item.IsParent = true
		}

		itemType, err := s.universe.Type(ctx, item.ItemTypeID)
		if err != nil {
			entry.WithField("item_id", item.ItemTypeID).WithError(err).Error("failed to fetch type infor for type")
			continue
		}

		item.ItemGroupID = itemType.GroupID
	}

	for _, item := range killmail.Victim.Items {
		if len(item.Items) > 0 {
			for _, subItem := range item.Items {
				subItem.KillmailID = killmail.ID
				subItemValue := float64(0)
				if subItem.Singleton != 2 {
					subItemValue = s.market.FetchTypePrice(subItem.ItemTypeID, date)
				} else {
					subItemValue = 0.01
				}
				subItem.ItemValue = subItemValue

				if subItem.QuantityDestroyed != nil && *subItem.QuantityDestroyed > 0 {
					destroyedValue += subItemValue * float64(*subItem.QuantityDestroyed)
				}
				if subItem.QuantityDropped != nil && *subItem.QuantityDropped > 0 {
					droppedValue += subItemValue * float64(*subItem.QuantityDropped)
				}

				subItemType, err := s.universe.Type(ctx, subItem.ItemTypeID)
				if err != nil {
					entry.WithField("item_id", subItem.ItemTypeID).WithError(err).Error("failed to fetch type infor for type")
					continue
				}

				subItem.ItemGroupID = subItemType.GroupID
			}
		}
	}

	fittedValue := s.calculatedFittedValue(killmail.Victim.Items)
	fittedValue += shipValue
	destroyedValue += shipValue
	sum := droppedValue + destroyedValue

	killmail.IsAwox = s.calcIsAwox(ctx, killmail)
	killmail.IsNPC = s.calcIsNPC(ctx, killmail)
	killmail.IsSolo = s.calcIsSolo(ctx, killmail)
	killmail.DestroyedValue = destroyedValue
	killmail.DroppedValue = droppedValue
	killmail.FittedValue = fittedValue
	killmail.TotalValue = sum

	err = s.killmails.CreateKillmail(ctx, killmail)
	if err != nil {
		entry.WithError(err).Error("error encountered inserting killmail victim into db")
		return nil, err
	}

	entry.Info("killmail successfully imported")

	// wg.Wait()

	return nil, nil
}

func (s *service) calcIsAwox(ctx context.Context, killmail *neo.Killmail) bool {

	if killmail.Victim == nil {
		return false
	}

	if killmail.Attackers == nil {
		return false
	}

	if killmail.Victim.CorporationID == nil {
		return false
	}

	victimCorporationID := *killmail.Victim.CorporationID
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
		if attacker.CorporationID == nil {
			continue
		}

		if attacker.ShipTypeID == nil {
			continue
		}

		attackerShip, err := s.universe.Type(ctx, *attacker.ShipTypeID)
		if err != nil {
			continue
		}

		switch attackerShip.GroupID {
		// Capsule, Shuttle, Corvette, Citizen Ships
		case 29, 31, 237, 361, 2001:
			goto LoopEnd
		}

		if *attacker.CorporationID == victimCorporationID {
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
		if attacker.CorporationID == nil {
			continue
		}

		if *attacker.CorporationID >= 98000000 {
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

	if attacker.CorporationID == nil {
		return false
	}

	return *attacker.CorporationID >= 98000000

}

func (s *service) primeKillmailNodes(ctx context.Context, killmail *neo.Killmail, entry *logrus.Entry) {
	system, err := s.universe.SolarSystem(ctx, killmail.SolarSystemID)
	if err != nil {
		entry.WithError(err).Error("failed to prime solar system")
	}

	constellation, err := s.universe.Constellation(ctx, system.ConstellationID)
	if err != nil {
		entry.WithError(err).Error("failed to prime constellation")
	}

	region, err := s.universe.Region(ctx, constellation.RegionID)
	if err != nil {
		entry.WithError(err).Error("failed to prime region")
	}

	constellation.Region = region
	system.Constellation = constellation
	killmail.ConstellationID = system.ConstellationID
	killmail.RegionID = constellation.RegionID

	victim := killmail.Victim

	if victim != nil {
		if victim.AllianceID != nil {
			_, err := s.alliance.Alliance(ctx, *victim.AllianceID)
			if err != nil {
				entry.WithError(err).Error("failed to prime victim alliance")
			}
		}
		if victim.CorporationID != nil {
			_, err := s.corporation.Corporation(ctx, *victim.CorporationID)
			if err != nil {
				entry.WithError(err).Error("failed to prime victim character")
			}
		}
		if victim.CharacterID != nil {
			_, err := s.character.Character(ctx, *victim.CharacterID)
			if err != nil {
				entry.WithError(err).Error("failed to prime victim character")
			}

		}
		_, err = s.universe.Type(ctx, victim.ShipTypeID)
		if err != nil {
			entry.WithError(err).Error("failed to prime victim ship type")
		}
	}
	for _, attacker := range killmail.Attackers {
		if attacker.AllianceID != nil {
			_, err := s.alliance.Alliance(ctx, *attacker.AllianceID)
			if err != nil {
				entry.WithError(err).Error("failed to prime attacker alliance")
			}
		}
		if attacker.CorporationID != nil {
			_, err := s.corporation.Corporation(ctx, *attacker.CorporationID)
			if err != nil {
				entry.WithError(err).Error("failed to prime attacker character")
			}
		}
		if attacker.CharacterID != nil {
			_, err := s.character.Character(ctx, *attacker.CharacterID)
			if err != nil {
				entry.WithError(err).Error("failed to prime attacker character")
			}
		}
		if attacker.ShipTypeID != nil {
			_, err = s.universe.Type(ctx, *attacker.ShipTypeID)
			if err != nil {
				entry.WithError(err).Error("failed to prime attacker ship type")
			}
		}
		if attacker.WeaponTypeID != nil {
			_, err = s.universe.Type(ctx, *attacker.WeaponTypeID)
			if err != nil {
				entry.WithError(err).Error("failed to prime attacker ship type")
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

		var qDestroyed uint = 0
		if item.QuantityDestroyed != nil {
			qDestroyed = *item.QuantityDestroyed
		}
		var qDropped uint = 0
		if item.QuantityDropped != nil {
			qDropped = *item.QuantityDropped
		}

		total += item.ItemValue * float64(qDestroyed+qDropped)
	}

	return total
}
