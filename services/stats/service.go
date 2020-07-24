package stats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eveisesi/neo/services/killmail"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Calculate() error
}

type service struct {
	redis  *redis.Client
	logger *logrus.Logger

	killmail killmail.Service

	neo.StatsRepository
}

func NewService(redis *redis.Client, logger *logrus.Logger, killmail killmail.Service, stats neo.StatsRepository) Service {
	return &service{
		redis,
		logger,
		killmail,
		stats,
	}
}

func (s *service) Calculate() error {

	subscription := s.redis.Subscribe(neo.QUEUES_KILLMAIL_STATS)
	channel := subscription.Channel()
	for msg := range channel {
		b := msg.Payload
		var message = neo.Message{}
		err := json.Unmarshal([]byte(b), &message)
		if err != nil {
			s.logger.WithError(err).WithField("message", string(b)).Error("unable to decode payload")
			time.Sleep(time.Millisecond * 100)
			continue
		}
		s.logger.WithFields(logrus.Fields{
			"id": message.ID, "hash": message.Hash,
		}).Infoln()
		err = s.calculate(message.ID, message.Hash)
		if err != nil {
			s.logger.WithError(err).WithField("message", string(b)).Error("failed to calculate stats")
			return err
		}
	}

	return nil
}

func (s *service) calculate(id uint64, hash string) error {

	var ctx = context.Background()

	killmail, err := s.killmail.FullKillmail(ctx, id, hash)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch full killmail for stats")
	}
	if killmail.IsNPC {
		return nil
	}

	stats := make([]*neo.Stat, 0)
	stats = append(stats, s.location(killmail)...)
	stats = append(stats, s.victim(killmail)...)
	stats = append(stats, s.attackers(killmail)...)

	chunks := chunkSliceStats(stats, 100)
	for _, chunk := range chunks {
		err := s.Save(ctx, chunk)
		if err != nil {
			return err
		}
	}

	return nil

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
