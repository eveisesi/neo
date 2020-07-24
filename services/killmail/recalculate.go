package killmail

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/korovkin/limiter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *service) RecalculatorDispatcher(limit, trigger int64, after uint64) {

	for {

		count, err := s.redis.ZCount(neo.QUEUES_KILLMAIL_RECALCULATE, "-inf", "+inf").Result()
		if err != nil {
			s.logger.WithError(err).Error("unable to determine count of recalculation message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count >= trigger {
			time.Sleep(time.Second * 10)
			continue
		}

		s.logger.Info("fetching killmail to recalculate")

		killmails, err := s.killmails.Recalculable(context.Background(), int(limit), after)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("failed to fetch killmails")
			return
		}

		if len(killmails) == 0 {
			s.logger.Info("no killmails returned")
			return
		}

		s.logger.WithField("killmails", len(killmails)).Info("killmails retrieved successfully")

		after = killmails[len(killmails)-1].ID

		for _, killmail := range killmails {

			msg := neo.Message{
				ID:   killmail.ID,
				Hash: killmail.Hash,
			}

			payload, err := json.Marshal(msg)
			if err != nil {
				s.logger.WithError(err).Error("failed to marshal message for recalculator queue")
				continue
			}

			_, err = s.redis.ZAdd(neo.QUEUES_KILLMAIL_RECALCULATE, &redis.Z{Score: float64(killmail.ID), Member: string(payload)}).Result()
			if err != nil {
				s.logger.WithError(err).WithField("payload", string(payload)).Error("unable to push killmail to recalculating queue")
				return
			}
		}

		s.logger.WithField("after", after).Info("killmails dispatched successfully")
	}

}

func (s *service) Recalculator(gLimit int64) {

	limit := limiter.NewConcurrencyLimiter(int(gLimit))

	attempts := 0
	for {
		count, err := s.redis.ZCount(neo.QUEUES_KILLMAIL_RECALCULATE, "-inf", "+inf").Result()
		if err != nil {
			s.logger.WithError(err).Error("unable to determine count of recalculation message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count == 0 {
			attempts++
			if attempts >= 100 {
				s.logger.Info("done with recalculation")
				break
			}
			s.logger.WithField("count", count).Info("no messages on queue")
			time.Sleep(time.Second * 10)
			continue
		}

		s.logger.WithField("messages", count).Info("processing recalculable messages")

		results, err := s.redis.ZPopMax(neo.QUEUES_KILLMAIL_RECALCULATE, 1000).Result()
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

	var payload = neo.Message{}

	err := json.Unmarshal(message, &payload)
	if err != nil {
		s.logger.WithField("message", string(message)).Error("failed to unmarshal message into message struct")
		return
	}

	entry := s.logger.WithFields(logrus.Fields{
		"id":   payload.ID,
		"hash": payload.Hash,
	})

	entry.Debugln()

	killmail, err := s.killmails.Killmail(ctx, payload.ID, payload.Hash)
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

	killmail.Attackers, err = s.attackers.ByKillmailID(ctx, killmail.ID)
	if err != nil {
		entry.WithError(err).Error("encountered error fetching attackers")
		return
	}

	s.primeKillmailNodes(ctx, killmail, entry)

}
