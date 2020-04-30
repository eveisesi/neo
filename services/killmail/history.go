package killmail

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func (s *service) HistoryExporter(min, max string) error {
	// Attempt to fetch Current Date from Redis
	current, err := s.redis.Get(neo.ZKB_HISTORY_DATE).Result()
	if err != nil && err.Error() != "redis: nil" {
		s.logger.WithError(err).Fatal("redis returned invalid response to query for egress date")
	}
	// If Current is an empty string, set it to the passed in max string
	if current == "" {
		current = max
	}

	// Date Parsing to get time.Time's to deal with
	mindate, err := time.Parse("20060102", min)
	if err != nil {
		s.logger.WithError(err).Fatal("unable to parse provided date")
	}

	maxdate, err := time.Parse("20060102", max)
	if err != nil {
		s.logger.WithError(err).Fatal("unable to parse provided date")
	}

	currentdate, err := time.Parse("20060102", current)
	if err != nil {
		s.logger.WithError(err).Fatal("unable to parse provided date")
	}

	if currentdate.Unix() > maxdate.Unix() || currentdate.Unix() < mindate.Unix() {
		s.redis.Del(neo.ZKB_HISTORY_DATE)
		return nil
	}

	// Store the new currentdate in Redis in case we panic
	_, err = s.redis.Set(neo.ZKB_HISTORY_DATE, currentdate.Format("20060102"), -1).Result()
	if err != nil {
		s.logger.WithError(err).Error("redis returned invalid response while setting egress date")
	}

	attempts := 1
	for {

		if attempts > 3 {
			return cli.NewExitError("maximum allowed attempts reeached", 1)
		}

		entry := s.logger.WithField("date", currentdate.Format("20060102"))
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

		currentdate = currentdate.AddDate(0, 0, -1)
		_, err = s.redis.Set(neo.ZKB_HISTORY_DATE, currentdate.Format("20060102"), -1).Result()
		if err != nil {
			s.logger.WithError(err).Error("redis returned invalid response while setting egress date")
		}

		entry.Info("handling hashes")
		s.handleHashes(hashes)
		entry.Info("finished with hashes && done pulling killmail history for date")

		if currentdate.Unix() > maxdate.Unix() || currentdate.Unix() < mindate.Unix() {
			s.redis.Del(neo.ZKB_HISTORY_DATE)
			return nil
		}

		time.Sleep(time.Millisecond * 500)

		attempts = 1
	}

}

func (s *service) handleHashes(hashes map[string]string) {

	// Make Sure the Redis Server is still alive and nothing has happened to it
	pong, err := s.redis.Ping().Result()
	if err != nil {
		s.logger.WithError(err).Fatal("unable to ping redis server")
	}
	// Make sure that redis returned pong to ping
	if pong != "PONG" {
		s.logger.WithField("pong", pong).Fatal("unexpected response to redis server ping received")
	}

	// Start the dispatch iterator
	dispatched := 0

	// Start a loop over the hashes that we got from ZKill
	for id, hash := range hashes {

		msg, err := json.Marshal(Message{
			ID:   id,
			Hash: hash,
		})
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"id":   id,
				"hash": hash,
			}).Error("unable to marshal id and hash for pubsub")
			continue
		}

		s.redis.ZAdd(neo.QUEUES_KILLMAIL_PROCESSING, redis.Z{Score: 2, Member: msg})
		dispatched++

	}

	count, err := s.redis.ZCount(neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
	if err != nil {
		s.logger.WithError(err).Fatal("unable to get count of redis zset")
	}

	s.logger.WithFields(logrus.Fields{
		"dispatched":    dispatched,
		"current_queue": count,
	}).Info("queue status")

}
