package killmail

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func (s *service) HistoryExporter(min, max string, datehold bool, threshold int64) error {

	txn := s.newrelic.StartTransaction("import history")
	ctx := newrelic.NewContext(context.Background(), txn)

	// Attempt to fetch Current Date from Redis
	current, err := s.redis.WithContext(ctx).Get(neo.ZKB_HISTORY_DATE).Result()
	if err != nil && err.Error() != "redis: nil" {
		s.logger.WithError(err).Fatal("redis returned invalid response to query for egress date")
	}
	// If Current is an empty string, set it to the passed in max string
	if current == "" {
		current = min
	}

	// Date Parsing to get time.Time's to deal with
	mindate, err := time.Parse("20060102", min)
	if err != nil {
		s.logger.WithError(err).Fatal("unable to parse provided date")
	}

	var maxdate time.Time
	if max == "" {
		now := time.Now()
		maxdate = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
	} else {
		maxdate, err = time.Parse("20060102", max)
		if err != nil {
			s.logger.WithError(err).Fatal("unable to parse provided date")
		}
	}

	currentdate, err := time.Parse("20060102", current)
	if err != nil {
		s.logger.WithError(err).Fatal("unable to parse provided date")
	}

	if currentdate.Unix() > maxdate.Unix() || currentdate.Unix() < mindate.Unix() {
		s.redis.WithContext(ctx).Del(neo.ZKB_HISTORY_DATE)
		return nil
	}

	// Store the new currentdate in Redis in case we panic
	_, err = s.redis.WithContext(ctx).Set(neo.ZKB_HISTORY_DATE, currentdate.Format("20060102"), -1).Result()
	if err != nil {
		s.logger.WithError(err).Error("redis returned invalid response while setting egress date")
	}

	attempts := 1
	for {

		if attempts > 3 {
			return cli.NewExitError("maximum allowed attempts reeached", 1)
		}

		entry := s.logger.WithContext(ctx).WithField("date", currentdate.Format("20060102"))
		entry.Info("pulling killmail history for date")

		uri := fmt.Sprintf(neo.ZKILLBOARD_HISTORY_API, currentdate.Format("20060102"))

		request, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			s.logger.WithError(err).Fatal("unable to generate request for zkillboard history api")
		}

		request.Header.Set("User-Agent", s.config.ZUAgent)

		response, err := s.client.Do(request)
		if err != nil {
			s.logger.WithError(err).Warn("unable to execute request to zkillboard history api")
		}
		entry = entry.WithField("code", response.StatusCode)

		if response.StatusCode != 200 {
			entry.Warn("unexpected status code recieved")
			time.Sleep(time.Second * 10)
			attempts++
			continue
		}

		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			entry.WithError(err).Warn("unable to read response body")
			time.Sleep(time.Second * 10)
			attempts++
			continue
		}
		response.Body.Close()

		if len(data) == 0 {
			entry.WithField("uri", uri).Warn("no data received from zkillboard api")
			// This maybe a bad date. Let decrement the date and try again. If attempts reaches 3, then this process will terminate
			currentdate = currentdate.AddDate(0, 0, -1)
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

		currentdate = currentdate.AddDate(0, 0, 1)
		_, err = s.redis.WithContext(ctx).Set(neo.ZKB_HISTORY_DATE, currentdate.Format("20060102"), -1).Result()
		if err != nil {
			s.logger.WithError(err).Error("redis returned invalid response while setting egress date")
		}

		entry.Info("handling hashes")
		s.handleHashes(ctx, hashes)
		entry.Info("finished with hashes && done pulling killmail history for date")

		if currentdate.Unix() > maxdate.Unix() || currentdate.Unix() < mindate.Unix() {
			s.redis.WithContext(ctx).Del(neo.ZKB_HISTORY_DATE)
			return nil
		}

		if datehold && threshold > 0 {
			i := 0
			for {
				count, err := s.redis.WithContext(ctx).ZCount(neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
				if err != nil {
					s.logger.WithError(err).Fatal("unable to get count of redis zset")
				}

				if count < threshold {
					break
				}
				if i%10 == 0 {
					entry.WithField("count", count).Infoln()
				}
				time.Sleep(time.Second)
				i++
			}
		}

		time.Sleep(time.Millisecond * 500)

		attempts = 1
	}

}

func (s *service) handleHashes(ctx context.Context, hashes map[string]string) {

	// Make Sure the Redis Server is still alive and nothing has happened to it
	pong, err := s.redis.WithContext(ctx).Ping().Result()
	if err != nil {
		s.logger.WithError(err).Fatal("unable to ping redis server")
	}
	// Make sure that redis returned pong to ping
	if pong != "PONG" {
		s.logger.WithField("pong", pong).Fatal("unexpected response to redis server ping received")
	}

	// Start the dispatch iterator
	dispatched := 0

	members := make([]*redis.Z, 0)

	// Start a loop over the hashes that we got from ZKill
	for id, hash := range hashes {

		killmailID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			s.logger.WithFields(logrus.Fields{"id": id, "hash": hash}).Error("unable to parse killmail id to uint")
			return
		}

		msg, err := json.Marshal(neo.Message{
			ID:   killmailID,
			Hash: hash,
		})
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"id":   id,
				"hash": hash,
			}).Error("unable to marshal id and hash for pubsub")
			continue
		}

		dispatched++

		members = append(members, &redis.Z{Score: float64(killmailID), Member: msg})
		if len(members) >= 250 {
			_, err := s.redis.WithContext(ctx).ZAdd(neo.QUEUES_KILLMAIL_PROCESSING, members...).Result()
			if err != nil {
				// Log error message
				s.logger.Error("failed to add historical hashes to redis queue")
				// Sleep for a second to see if this helps
				time.Sleep(time.Second)
			}
			members = make([]*redis.Z, 0)
		}

	}

	_, err = s.redis.WithContext(ctx).ZAdd(neo.QUEUES_KILLMAIL_PROCESSING, members...).Result()
	if err != nil {
		// Log error message
		s.logger.Error("failed to add historical hashes to redis queue")
		// Sleep for a second to see if this helps
		time.Sleep(time.Second)
	}

	count, err := s.redis.WithContext(ctx).ZCount(neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
	if err != nil {
		s.logger.WithError(err).Fatal("unable to get count of redis zset")
	}

	s.logger.WithFields(logrus.Fields{
		"dispatched":    dispatched,
		"current_queue": count,
	}).Info("queue status")

}
