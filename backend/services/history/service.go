package history

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v8"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run(startDateStr, endDateStr string, incrementer int64, stats bool)
}

type service struct {
	client   *http.Client
	redis    *redis.Client
	logger   *logrus.Logger
	config   *neo.Config
	newrelic *newrelic.Application

	neo.KillmailRepository
}

const (
	defaultTimeFormat string = "20060102"
)

func NewService(
	client *http.Client,
	redis *redis.Client,
	logger *logrus.Logger,
	nr *newrelic.Application,
	config *neo.Config,

	killmail neo.KillmailRepository,
) Service {
	return &service{
		client:   client,
		redis:    redis,
		logger:   logger,
		newrelic: nr,
		config:   config,

		KillmailRepository: killmail,
	}
}

func (s *service) Run(startDateStr, endDateStr string, incrementer int64, stats bool) {

	var now = time.Now()
	var startDate time.Time
	var endDate = new(time.Time)
	var err error

	ctx := context.Background()

	entry := s.logger.WithContext(ctx)
	// If not startDate was provided, default to yesterday
	if startDateStr == "" {
		startDateStr = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC).Format(defaultTimeFormat)
	}
	startDate, err = time.Parse(defaultTimeFormat, startDateStr)
	if err != nil {
		entry.WithError(err).Error("failed to parse startDate")
	}

	// If an endDate was provided, set it.
	if endDateStr != "" {
		x, err := time.Parse(defaultTimeFormat, endDateStr)
		if err != nil {
			entry.WithError(err).Error("failed to parse startDate")
		}

		*endDate = x
	}

	killsByDayMap := make(map[string]int64)
	totalsURL := fmt.Sprintf(neo.ZKILLBOARD_HISTORY_API, "totals")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, totalsURL, nil)
	if err != nil {
		entry.WithError(err).Fatal("unable to create request to totals API")
	}

	req.Header.Set("User-Agent", s.config.ZUAgent)

	res, err := s.client.Do(req)
	if err != nil {
		entry.WithError(err).Fatal("failed to make request to zkillboard for killmail totals")
	}

	entry = entry.WithField("status_code", res.StatusCode)

	if res.StatusCode != http.StatusOK {
		entry.Fatal("unexpected status code received from zkill totals api")
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		entry.WithError(err).Fatal("failed to read response body from zkillboard")
	}

	err = res.Body.Close()
	if err != nil {
		entry.WithError(err).Fatal("failed to close resp body for request to zkillboard")
	}

	err = json.Unmarshal(data, &killsByDayMap)
	if err != nil {
		entry.WithError(err).Fatal("unable to decode zkill totals response into map")
	}

	current := startDate
	missingDateCounter := 0
	totalAttempts := 1
	for {
		if missingDateCounter >= 3 {
			entry.Error("missing date counter violation, breaking loop")
			break
		}
		if endDate != nil && (incrementer == 1 && current.After(*endDate)) || (incrementer == -1 && current.Before(*endDate)) {
			entry.Info("endDate reached, breaking loop")
			break
		}

		currentStr := current.Format(defaultTimeFormat)

		entry = entry.WithField("date", currentStr)
		totalEntry, ok := killsByDayMap[currentStr]
		if !ok {
			entry.Error("totalMap does not contain entry for date")
			current = current.AddDate(0, 0, int(incrementer))
			missingDateCounter++
			continue
		}
		missingDateCounter = 0

		countKillmailsFilters := []*neo.Operator{
			neo.NewGreaterThanEqualToOperator("killmailTime", time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)),
			neo.NewLessThanEqualToOperator("killmailTime", time.Date(current.Year(), current.Month(), current.Day(), 23, 59, 59, 0, time.UTC)),
		}

		killmailCount, err := s.CountKillmails(ctx, countKillmailsFilters...)
		if err != nil {
			entry.WithError(err).Error("encountered error querying killmail count for date")
		}

		entry = entry.WithFields(logrus.Fields{
			"currentDate":   currentStr,
			"totalEntry":    totalEntry,
			"killmailCount": killmailCount,
			"pass":          totalEntry == killmailCount,
		})

		if stats {
			entry.Infoln()
			current = current.AddDate(0, 0, int(incrementer))
			continue
		}

		if killmailCount >= totalEntry {
			// TODO: Write this fact to the DB and then also check this table
			// before counting killmails and after calling ZKillboard
			entry.Info("killmail count equal history api, skipping date")
			time.Sleep(time.Millisecond * 500)
			current = current.AddDate(0, 0, int(incrementer))
			continue
		}

		time.Sleep(time.Millisecond * 500)

		attempts := 1
		for {

			if attempts > 3 {
				entry.Fatal("maximum allowed attempts reeached")
			}

			uri := fmt.Sprintf(neo.ZKILLBOARD_HISTORY_API, current.Format("20060102"))
			entry.WithField("uri", uri).Info("pulling killmail history for date")

			request, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
			if err != nil {
				s.logger.WithError(err).Fatal("unable to generate request for zkillboard history api")
			}

			request.Header.Set("User-Agent", s.config.ZUAgent)
			// extSeg := newrelic.StartExternalSegment(txn, request)
			response, err := s.client.Do(request)
			if err != nil {
				s.logger.WithError(err).Warn("unable to execute request to zkillboard history api")
			}
			// extSeg.Response = response
			// extSeg.End()

			entry = entry.WithField("code", response.StatusCode)
			if response.StatusCode != 200 {
				entry.Fatal("unexpected status code recieved")
			}

			data, err := ioutil.ReadAll(response.Body)
			if err != nil {
				entry.WithError(err).Warn("unable to read response body")
				response.Body.Close()
				time.Sleep(time.Second * 10)
				attempts++
				continue
			}

			if len(data) == 0 {
				entry.WithField("uri", uri).Warn("no data received from zkillboard api")
				response.Body.Close()
				time.Sleep(time.Second * 10)
				attempts++
				continue
			}

			err = response.Body.Close()
			if err != nil {
				entry.WithError(err).WithField("uri", uri).Fatal("unable to close response body stream")
			}

			var hashes = make(map[string]string)
			err = json.Unmarshal(data, &hashes)
			if err != nil {
				entry.WithError(err).Warn("unable to read response body")
				time.Sleep(time.Second * 10)
				attempts++
				continue
			}

			// We want to be able to perform bulk lookups to the DB, so we need to convert this map[string]string to a slice of ints
			ids := make([]neo.ModValue, len(hashes))
			i := 0
			for killID, hash := range hashes {
				id, err := strconv.Atoi(killID)
				if err != nil {
					entry.WithFields(logrus.Fields{
						"id":   killID,
						"hash": hash,
					}).Fatal("failed to parse str killID to int killID")
				}

				ids[i] = id
				i++
			}
			countKillmailsFilters := append(countKillmailsFilters, neo.NewInOperator("id", ids))
			killmails, err := s.Killmails(ctx, countKillmailsFilters...)
			if err != nil {
				entry.WithError(err).Error("failed to query killmails for array of killIDs")
				return
			}
			missing := make([]neo.Message, 0)
			for _, id := range ids {
				found := false
				for _, killmail := range killmails {
					if killmail.ID == uint(id.(int)) {
						found = true
						break
					}
				}
				if !found {
					missing = append(missing, neo.Message{
						ID:   uint(id.(int)),
						Hash: hashes[strconv.Itoa(id.(int))],
					})
				}
			}

			if len(missing) == 0 {
				entry.Info("no killmails missing, breaking hashes loop")
				break
			}

			entry.Info("handling hashes")
			s.handleHashes(ctx, missing)
			entry.Info("finished with hashes && done pulling killmail history for date")

			break
		}
		i := 0
		for {

			count, err := s.redis.ZCount(ctx, neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
			if err != nil {
				s.logger.WithError(err).Fatal("unable to get count of redis zset")
			}

			if count == 0 {
				entry.Info("queue is at zero, break zero check loop")
				break
			}
			if i%20 == 0 {
				entry.WithField("count", count).Infoln()
			}

			time.Sleep(time.Second)
			i++
		}
		entry.Info("confirm killmail count for date")
		killmailCount, err = s.CountKillmails(ctx, countKillmailsFilters...)
		if err != nil {
			entry.WithError(err).Error("encountered error querying killmail count for date")
		}

		if killmailCount < totalEntry && totalAttempts < 2 {
			totalAttempts++
			entry.WithFields(logrus.Fields{
				"killmailCount": killmailCount,
				"totalEntry":    totalEntry,
			}).Info("count do not equal, trying again.")
			time.Sleep(time.Second)
		} else {
			// TODO: Write this fact to the DB and then also check this table
			// before counting killmails and after calling ZKillboard
			entry.Info("killmail count equal history api, moving to next date")
			totalAttempts = 1
			time.Sleep(time.Millisecond * 500)
			current = current.AddDate(0, 0, int(incrementer))
		}

	}

}

func (s *service) handleHashes(ctx context.Context, missing []neo.Message) {

	// Start the dispatch iterator
	dispatched := 0

	members := make([]*redis.Z, 0)

	// Start a loop over the hashes that we got from ZKill
	for _, msg := range missing {

		data, err := json.Marshal(msg)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"id":   msg.ID,
				"hash": msg.Hash,
			}).Error("unable to marshal id and hash for pubsub")
			continue
		}

		dispatched++

		members = append(members, &redis.Z{Score: float64(msg.ID), Member: data})
		if len(members) >= 250 {
			_, err := s.redis.ZAdd(ctx, neo.QUEUES_KILLMAIL_PROCESSING, members...).Result()
			if err != nil {
				// Log error message
				s.logger.WithError(err).Error("failed to add historical hashes to redis queue")
				// Sleep for a second to see if this helps
				time.Sleep(time.Second)
			}
			members = make([]*redis.Z, 0)
		}

	}

	_, err := s.redis.ZAdd(ctx, neo.QUEUES_KILLMAIL_PROCESSING, members...).Result()
	if err != nil {
		// Log error message
		s.logger.Error("failed to add historical hashes to redis queue")
		// Sleep for a second to see if this helps
		time.Sleep(time.Second)
	}

	count, err := s.redis.ZCount(ctx, neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
	if err != nil {
		s.logger.WithError(err).Fatal("unable to get count of redis zset")
	}

	s.logger.WithFields(logrus.Fields{
		"dispatched":    dispatched,
		"current_queue": count,
	}).Info("queue status")

}
