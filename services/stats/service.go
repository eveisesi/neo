package stats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eveisesi/neo/services/killmail"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run() error
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
			// entry.Info("stats queue is empty")
			time.Sleep(time.Second * 2)
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
			}

			s.processMessage(message)
		}
	}

}

func (s *service) processMessage(msg neo.Message) {

	txn := s.newrelic.StartTransaction("process stats message")
	defer txn.End()

	ctx := newrelic.NewContext(context.Background(), txn)

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"id":   msg.ID,
		"hash": msg.Hash,
	})

	killmail, err := s.killmail.FullKillmail(ctx, msg.ID, msg.Hash)
	if err != nil {
		entry.WithError(err).Error("failed to fetch full killmail for stats")
	}
	if killmail.IsNPC {
		return
	}

	stats := make([]*neo.Stat, 0)
	stats = append(stats, s.location(killmail)...)
	stats = append(stats, s.victim(killmail)...)
	stats = append(stats, s.attackers(killmail)...)

	chunks := chunkSliceStats(stats, 100)
	for _, chunk := range chunks {
		err := s.Save(ctx, chunk)
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
		ID:        killmail.SolarSystemID,
		Entity:    neo.StatEntitySystem,
		Category:  neo.StatCategoryShipsKilled,
		Frequency: neo.StatFrequencyDaily,
		Date:      date,
		Value:     1,
	})
	stats = append(stats, &neo.Stat{
		ID:        killmail.SolarSystemID,
		Entity:    neo.StatEntitySystem,
		Category:  neo.StatCategoryISKKilled,
		Frequency: neo.StatFrequencyDaily,
		Date:      date,
		Value:     killmail.TotalValue,
	})
	if killmail.System != nil {
		stats = append(stats, &neo.Stat{
			ID:        killmail.System.ConstellationID,
			Entity:    neo.StatEntityConstellation,
			Category:  neo.StatCategoryShipsKilled,
			Frequency: neo.StatFrequencyDaily,
			Date:      date,
			Value:     1,
		})
		stats = append(stats, &neo.Stat{
			ID:        killmail.System.ConstellationID,
			Entity:    neo.StatEntityConstellation,
			Category:  neo.StatCategoryISKKilled,
			Frequency: neo.StatFrequencyDaily,
			Date:      date,
			Value:     killmail.TotalValue,
		})
		if killmail.System.Constellation != nil {
			stats = append(stats, &neo.Stat{
				ID:        killmail.System.Constellation.RegionID,
				Entity:    neo.StatEntityRegion,
				Category:  neo.StatCategoryShipsKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     1,
			})
			stats = append(stats, &neo.Stat{
				ID:        killmail.System.Constellation.RegionID,
				Entity:    neo.StatEntityRegion,
				Category:  neo.StatCategoryISKKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     killmail.TotalValue,
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
			ID:        victim.CharacterID.Uint64,
			Entity:    neo.StatEntityCharacter,
			Category:  neo.StatCategoryShipsLost,
			Frequency: neo.StatFrequencyDaily,
			Date:      date,
			Value:     1,
		})
		stats = append(stats, &neo.Stat{
			ID:        victim.CharacterID.Uint64,
			Entity:    neo.StatEntityCharacter,
			Category:  neo.StatCategoryISKLost,
			Frequency: neo.StatFrequencyDaily,
			Date:      date,
			Value:     killmail.TotalValue,
		})
	}
	if victim.CorporationID.Valid {
		stats = append(stats, &neo.Stat{
			ID:        victim.CorporationID.Uint64,
			Entity:    neo.StatEntityCorporation,
			Category:  neo.StatCategoryShipsLost,
			Frequency: neo.StatFrequencyDaily,
			Date:      date,
			Value:     1,
		})
		stats = append(stats, &neo.Stat{
			ID:        victim.CorporationID.Uint64,
			Entity:    neo.StatEntityCorporation,
			Category:  neo.StatCategoryISKLost,
			Frequency: neo.StatFrequencyDaily,
			Date:      date,
			Value:     killmail.TotalValue,
		})
	}
	if victim.AllianceID.Valid {
		stats = append(stats, &neo.Stat{
			ID:        victim.AllianceID.Uint64,
			Entity:    neo.StatEntityAlliance,
			Category:  neo.StatCategoryShipsLost,
			Frequency: neo.StatFrequencyDaily,
			Date:      date,
			Value:     1,
		})
		stats = append(stats, &neo.Stat{
			ID:        victim.AllianceID.Uint64,
			Entity:    neo.StatEntityAlliance,
			Category:  neo.StatCategoryISKLost,
			Frequency: neo.StatFrequencyDaily,
			Date:      date,
			Value:     killmail.TotalValue,
		})
	}

	stats = append(stats, &neo.Stat{
		ID:        victim.ShipTypeID,
		Entity:    neo.StatEntityShip,
		Category:  neo.StatCategoryShipsLost,
		Frequency: neo.StatFrequencyDaily,
		Date:      date,
		Value:     1,
	})
	stats = append(stats, &neo.Stat{
		ID:        victim.ShipTypeID,
		Entity:    neo.StatEntityShip,
		Category:  neo.StatCategoryISKLost,
		Frequency: neo.StatFrequencyDaily,
		Date:      date,
		Value:     killmail.TotalValue,
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
				ID:        attacker.CharacterID.Uint64,
				Entity:    neo.StatEntityCharacter,
				Category:  neo.StatCategoryShipsKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     1,
			})
			stats = append(stats, &neo.Stat{
				ID:        attacker.CharacterID.Uint64,
				Entity:    neo.StatEntityCharacter,
				Category:  neo.StatCategoryISKKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     killmail.TotalValue,
			})
		}
		if attacker.CorporationID.Valid {
			stats = append(stats, &neo.Stat{
				ID:        attacker.CorporationID.Uint64,
				Entity:    neo.StatEntityCorporation,
				Category:  neo.StatCategoryShipsKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     1,
			})
			stats = append(stats, &neo.Stat{
				ID:        attacker.CorporationID.Uint64,
				Entity:    neo.StatEntityCorporation,
				Category:  neo.StatCategoryISKKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     killmail.TotalValue,
			})
		}
		if attacker.AllianceID.Valid {
			stats = append(stats, &neo.Stat{
				ID:        attacker.AllianceID.Uint64,
				Entity:    neo.StatEntityAlliance,
				Category:  neo.StatCategoryShipsKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     1,
			})
			stats = append(stats, &neo.Stat{
				ID:        attacker.AllianceID.Uint64,
				Entity:    neo.StatEntityAlliance,
				Category:  neo.StatCategoryISKKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     killmail.TotalValue,
			})
		}

		if attacker.ShipTypeID.Valid {
			stats = append(stats, &neo.Stat{
				ID:        attacker.ShipTypeID.Uint64,
				Entity:    neo.StatEntityShip,
				Category:  neo.StatCategoryShipsKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     1,
			})
			stats = append(stats, &neo.Stat{
				ID:        attacker.ShipTypeID.Uint64,
				Entity:    neo.StatEntityShip,
				Category:  neo.StatCategoryISKKilled,
				Frequency: neo.StatFrequencyDaily,
				Date:      date,
				Value:     killmail.TotalValue,
			})
		}
	}

	return stats
}
