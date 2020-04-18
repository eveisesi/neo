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

func (s *service) HistoryExporter(channel, date string) error {

	redisKey := "neo:egress:date"

	result, err := s.redis.Get(redisKey).Result()
	if err != nil && err.Error() != "redis: nil" {
		s.logger.WithError(err).Fatal("redis returned invalid response to query for egress date")
	}

	if result == "" {
		if date == "" {
			return nil
		}
		result = date
	}

	parsed, err := time.Parse("20060102", result)
	if err != nil {
		s.logger.WithError(err).Fatal("unable to parse provided date")
	}

	_, err = s.redis.Set(redisKey, parsed.Format("20060102"), -1).Result()
	if err != nil {
		s.logger.WithError(err).Error("redis returned invalid response while setting egress date")
	}

	attempts := 1
	for {

		if attempts > 3 {
			return cli.NewExitError("maximum allowed attempts reeached", 1)
		}

		entry := s.logger.WithField("date", parsed.Format("20060102"))

		uri := fmt.Sprintf(neo.ZKILLBOARD_HISTORY_API, parsed.Format("20060102"))

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
			parsed = parsed.AddDate(0, 0, -1)
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

		parsed = parsed.AddDate(0, 0, -1)
		_, err = s.redis.Set(redisKey, parsed.Format("20060102"), -1).Result()
		if err != nil {
			s.logger.WithError(err).Error("redis returned invalid response while setting egress date")
		}

		entry.Info("handling hashes")
		s.handleHashes(channel, hashes)
		entry.Info("finished with hashes")

		attempts = 1
	}

}

func (s *service) handleHashes(channel string, hashes map[string]string) {

	// Make Sure the Redis Server is still alive and nothing has happened to it
	pong, err := s.redis.Ping().Result()
	if err != nil {
		s.logger.WithError(err).Fatal("unable to ping redis server")
	}
	// Make sure that redis returned pong to ping
	if pong != "PONG" {
		s.logger.WithField("pong", pong).Fatal("unexpected response to redis server ping received")
	}

	// Lets get the most recent record from the end of the set to determine the score to use
	results, err := s.redis.ZRevRangeByScoreWithScores(channel, redis.ZRangeBy{Min: "-inf", Max: "+inf", Count: 1}).Result()
	if err != nil {
		s.logger.WithError(err).Fatal("unable to get max score of redis z range")
	}

	// If we received more than one result, something is wrong and we need to bail
	if len(results) > 1 {
		s.logger.WithError(err).Fatal("unable to determine score")
	}
	// Default the score to 0 incase the set is empty
	score := float64(0)
	if len(results) == 1 {
		// Get the score
		score = results[0].Score
	}

	// Start the dispatch iterator
	dispatched := 0

	// Start a loop over the hashes that we got from ZKill
	for id, hash := range hashes {

		score++
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

		s.redis.ZAdd(channel, redis.Z{Score: score, Member: msg})
		dispatched++

	}

	for {

		count, err := s.redis.ZCount(channel, "-inf", "+inf").Result()
		if err != nil {
			s.logger.WithError(err).Fatal("unable to get count of redis zset")
		}

		s.logger.WithFields(logrus.Fields{
			"total":         len(hashes),
			"dispatched":    dispatched,
			"remaining":     len(hashes) - dispatched,
			"current_queue": count,
		}).Info("queue status")
		if count < 200 {
			return
		}

		time.Sleep(time.Second * 15)

	}

}
