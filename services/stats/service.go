package stats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eveisesi/neo/services/killmail"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run() error
	// Recalculate(id int64, entityType neo.StatEntity, date time.Time) error
	neo.StatsRepository
}

type service struct {
	redis    *redis.Client
	logger   *logrus.Logger
	newrelic *newrelic.Application

	killmail killmail.Service

	neo.StatsRepository
}

func NewService(redis *redis.Client, logger *logrus.Logger, newrelic *newrelic.Application, killmail killmail.Service, stats neo.StatsRepository) Service {
	return &service{
		redis,
		logger,
		newrelic,
		killmail,
		stats,
	}
}

func (s *service) Run() error {

	for {
		entry := s.logger
		count, err := s.redis.ZCount(neo.QUEUES_KILLMAIL_STATS, "-inf", "+inf").Result()
		if err != nil {
			entry.WithError(err).Error("unable to determine count of message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count == 0 {
			entry.Info("stats queue is empty")
			time.Sleep(time.Second)
			continue
		}

		results, err := s.redis.ZPopMax(neo.QUEUES_KILLMAIL_STATS, 5).Result()
		if err != nil {
			entry.WithError(err).Fatal("unable to retrieve hashes from queue")
		}

		for _, result := range results {
			var message neo.Message
			err := json.Unmarshal([]byte(result.Member.(string)), &message)
			if err != nil {
				s.logger.WithError(err).WithField("membver", result.Member).Error("failed to unmarshal queue payload")
				continue
			}

			s.processMessage(message)
		}
	}

}

// func (s *service) Recalculate(id int64, entityType neo.StatEntity, date time.Time) error {

// 	txn := s.newrelic.StartTransaction("recalculate stats")
// 	txn.AddAttribute("id", id)
// 	txn.AddAttribute("type", entityType.String())
// 	txn.AddAttribute("date", date.Format("YYYYMMDD"))
// 	ctx := newrelic.NewContext(context.Background(), txn)

// 	err := s.DeleteStats(ctx)

// 	// err := s.DeleteRecordsByTypeAfterDate(ctx, id, entityType, date)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete existing stats: %w", err)
// 	}
// 	i := 1
// 	for {
// 		killmails, err := s.killmail.KillmailsByShipID(ctx, uint64(id), i)
// 		if err != nil && !errors.Is(err, sql.ErrNoRows) {
// 			return fmt.Errorf("failed to fetch killmails to recalculate stats")
// 		}
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil
// 		}

// 		spew.Dump(len(killmails))
// 		i++
// 		time.Sleep(time.Second)

// 	}

// 	return nil

// }

func (s *service) processMessage(msg neo.Message) {

	txn := s.newrelic.StartTransaction("process stats message")
	defer txn.End()

	ctx := newrelic.NewContext(context.Background(), txn)

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"id":   msg.ID,
		"hash": msg.Hash,
	})

	killmail, err := s.killmail.FullKillmail(ctx, msg.ID, false)
	if err != nil {
		entry.WithError(err).Error("failed to fetch full killmail for stats")
		return
	}
	if killmail.IsNPC {
		return
	}

	stats := make([]*neo.Stat, 0)
	stats = append(stats, s.location(killmail)...)
	stats = append(stats, s.victim(killmail)...)
	stats = append(stats, s.attackers(killmail)...)

	chunks := chunkSliceStats(stats, 100)
	for _, _ = range chunks {
		err = nil
		if err != nil {
			entry.WithError(err).Error("encountered error calculating stats")
			return
		}
	}

	entry.Info("stats calculated successfully")

}

func chunkSliceStats(slice []*neo.Stat, size int) [][]*neo.Stat {

	var chunk = make([][]*neo.Stat, 0)
	if len(slice) <= size {
		chunk = append(chunk, slice)
		slice = nil
		return chunk
	}

	for x := 0; x < len(slice); x += size {
		end := x + size

		if end > len(slice) {
			end = len(slice)
		}

		chunk = append(chunk, slice[x:end])
	}

	slice = nil

	return chunk

}

func (s *service) date(t time.Time) *neo.Date {
	return &neo.Date{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

func (s *service) location(killmail *neo.Killmail) []*neo.Stat {

	date := s.date(killmail.KillmailTime)
	stats := make([]*neo.Stat, 0)
	stats = append(stats, &neo.Stat{
		EntityID:   uint64(killmail.SolarSystemID),
		EntityType: neo.StatEntitySystem,
		Category:   neo.StatCategoryShipsKilled,
		Frequency:  neo.StatFrequencyDaily,
		Date:       date,
		Value:      1,
	})
	stats = append(stats, &neo.Stat{
		EntityID:   uint64(killmail.SolarSystemID),
		EntityType: neo.StatEntitySystem,
		Category:   neo.StatCategoryISKKilled,
		Frequency:  neo.StatFrequencyDaily,
		Date:       date,
		Value:      killmail.TotalValue,
	})
	if killmail.System != nil {
		stats = append(stats, &neo.Stat{
			EntityID:   uint64(killmail.System.ConstellationID),
			EntityType: neo.StatEntityConstellation,
			Category:   neo.StatCategoryShipsKilled,
			Frequency:  neo.StatFrequencyDaily,
			Date:       date,
			Value:      1,
		})
		stats = append(stats, &neo.Stat{
			EntityID:   uint64(killmail.System.ConstellationID),
			EntityType: neo.StatEntityConstellation,
			Category:   neo.StatCategoryISKKilled,
			Frequency:  neo.StatFrequencyDaily,
			Date:       date,
			Value:      killmail.TotalValue,
		})
		if killmail.System.Constellation != nil {
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(killmail.System.Constellation.RegionID),
				EntityType: neo.StatEntityRegion,
				Category:   neo.StatCategoryShipsKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      1,
			})
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(killmail.System.Constellation.RegionID),
				EntityType: neo.StatEntityRegion,
				Category:   neo.StatCategoryISKKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      killmail.TotalValue,
			})
		}
	}
	return stats

}

func (s *service) victim(killmail *neo.Killmail) []*neo.Stat {
	date := s.date(killmail.KillmailTime)
	stats := make([]*neo.Stat, 0)
	victim := killmail.Victim

	if victim.CharacterID.Valid {
		stats = append(stats, &neo.Stat{
			EntityID:   victim.CharacterID.Uint64,
			EntityType: neo.StatEntityCharacter,
			Category:   neo.StatCategoryShipsLost,
			Frequency:  neo.StatFrequencyDaily,
			Date:       date,
			Value:      1,
		})
		stats = append(stats, &neo.Stat{
			EntityID:   victim.CharacterID.Uint64,
			EntityType: neo.StatEntityCharacter,
			Category:   neo.StatCategoryISKLost,
			Frequency:  neo.StatFrequencyDaily,
			Date:       date,
			Value:      killmail.TotalValue,
		})
	}
	if victim.CorporationID.Valid {
		stats = append(stats, &neo.Stat{
			EntityID:   uint64(victim.CorporationID.Uint),
			EntityType: neo.StatEntityCorporation,
			Category:   neo.StatCategoryShipsLost,
			Frequency:  neo.StatFrequencyDaily,
			Date:       date,
			Value:      1,
		})
		stats = append(stats, &neo.Stat{
			EntityID:   uint64(victim.CorporationID.Uint),
			EntityType: neo.StatEntityCorporation,
			Category:   neo.StatCategoryISKLost,
			Frequency:  neo.StatFrequencyDaily,
			Date:       date,
			Value:      killmail.TotalValue,
		})
	}
	if victim.AllianceID.Valid {
		stats = append(stats, &neo.Stat{
			EntityID:   uint64(victim.AllianceID.Uint),
			EntityType: neo.StatEntityAlliance,
			Category:   neo.StatCategoryShipsLost,
			Frequency:  neo.StatFrequencyDaily,
			Date:       date,
			Value:      1,
		})
		stats = append(stats, &neo.Stat{
			EntityID:   uint64(victim.AllianceID.Uint),
			EntityType: neo.StatEntityAlliance,
			Category:   neo.StatCategoryISKLost,
			Frequency:  neo.StatFrequencyDaily,
			Date:       date,
			Value:      killmail.TotalValue,
		})
	}

	stats = append(stats, &neo.Stat{
		EntityID:   uint64(victim.ShipTypeID),
		EntityType: neo.StatEntityShip,
		Category:   neo.StatCategoryShipsLost,
		Frequency:  neo.StatFrequencyDaily,
		Date:       date,
		Value:      1,
	})
	stats = append(stats, &neo.Stat{
		EntityID:   uint64(victim.ShipTypeID),
		EntityType: neo.StatEntityShip,
		Category:   neo.StatCategoryISKLost,
		Frequency:  neo.StatFrequencyDaily,
		Date:       date,
		Value:      killmail.TotalValue,
	})

	stats = append(stats, &neo.Stat{
		EntityID:   uint64(victim.ShipGroupID),
		EntityType: neo.StatEntityShipGroup,
		Category:   neo.StatCategoryShipsLost,
		Frequency:  neo.StatFrequencyDaily,
		Date:       date,
		Value:      1,
	})
	stats = append(stats, &neo.Stat{
		EntityID:   uint64(victim.ShipGroupID),
		EntityType: neo.StatEntityShipGroup,
		Category:   neo.StatCategoryISKLost,
		Frequency:  neo.StatFrequencyDaily,
		Date:       date,
		Value:      killmail.TotalValue,
	})

	return stats

}

func (s *service) attackers(killmail *neo.Killmail) []*neo.Stat {
	date := s.date(killmail.KillmailTime)
	stats := make([]*neo.Stat, 0)
	attackers := killmail.Attackers

	for _, attacker := range attackers {
		if attacker.CharacterID.Valid {
			stats = append(stats, &neo.Stat{
				EntityID:   attacker.CharacterID.Uint64,
				EntityType: neo.StatEntityCharacter,
				Category:   neo.StatCategoryShipsKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      1,
			})
			stats = append(stats, &neo.Stat{
				EntityID:   attacker.CharacterID.Uint64,
				EntityType: neo.StatEntityCharacter,
				Category:   neo.StatCategoryISKKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      killmail.TotalValue,
			})
		}
		if attacker.CorporationID.Valid {
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(attacker.CorporationID.Uint),
				EntityType: neo.StatEntityCorporation,
				Category:   neo.StatCategoryShipsKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      1,
			})
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(attacker.CorporationID.Uint),
				EntityType: neo.StatEntityCorporation,
				Category:   neo.StatCategoryISKKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      killmail.TotalValue,
			})
		}
		if attacker.AllianceID.Valid {
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(attacker.AllianceID.Uint),
				EntityType: neo.StatEntityAlliance,
				Category:   neo.StatCategoryShipsKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      1,
			})
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(attacker.AllianceID.Uint),
				EntityType: neo.StatEntityAlliance,
				Category:   neo.StatCategoryISKKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      killmail.TotalValue,
			})
		}

		if attacker.ShipTypeID.Valid {
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(attacker.ShipTypeID.Uint),
				EntityType: neo.StatEntityShip,
				Category:   neo.StatCategoryShipsKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      1,
			})
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(attacker.ShipTypeID.Uint),
				EntityType: neo.StatEntityShip,
				Category:   neo.StatCategoryISKKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      killmail.TotalValue,
			})

		}

		if attacker.ShipGroupID.Valid {
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(attacker.ShipGroupID.Uint),
				EntityType: neo.StatEntityShipGroup,
				Category:   neo.StatCategoryShipsKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      1,
			})
			stats = append(stats, &neo.Stat{
				EntityID:   uint64(attacker.ShipGroupID.Uint),
				EntityType: neo.StatEntityShipGroup,
				Category:   neo.StatCategoryISKKilled,
				Frequency:  neo.StatFrequencyDaily,
				Date:       date,
				Value:      killmail.TotalValue,
			})
		}
	}

	return stats
}
