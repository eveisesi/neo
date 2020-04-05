package egress

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli"

	"github.com/eveisesi/neo"
	core "github.com/eveisesi/neo/app"
)

type Message struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
}

type Egressor struct {
	*core.App
}

func Action(c *cli.Context) error {

	e := &Egressor{
		core.New(),
	}

	redisKey := "neo:egress:date"
	result, err := e.Redis.Get(redisKey).Result()
	if err != nil && err.Error() != "redis: nil" {
		e.Logger.WithError(err).Fatal("redis returned invalid response to query for egress date")
	}

	if result == "" {
		result = c.String("date")
	}

	date, err := time.Parse("20060102", result)
	if err != nil {
		e.Logger.WithError(err).Fatal("unable to parse provided date")
	}

	_, err = e.Redis.Set(redisKey, date.Format("20060102"), -1).Result()
	if err != nil {
		e.Logger.WithError(err).Error("redis returned invalid response while setting egress date")
	}

	attempts := 1
	for {

		if attempts > 3 {
			return cli.NewExitError("maximum allowed attempts reeached", 1)
		}

		entry := e.Logger.WithField("date", date.Format("20060102"))

		uri := fmt.Sprintf(neo.ZKILLBOARD_HISTORY_API, date.Format("20060102"))

		request, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			e.Logger.WithError(err).Fatal("unable to generate request for zkillboard history api")
		}

		request.Header.Set("User-Agent", e.Config.ZUAgent)

		response, err := e.Client.Do(request)
		if err != nil {
			e.Logger.WithError(err).Warn("unable to execute request to zkillboard history api")
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

		if len(data) == 0 {
			entry.WithField("uri", uri).Warn("no data received from zkillboard api")
			// This maybe a bad date. Let increment the date and try again. If attempts reaches 3, then this process will terminate
			date = date.AddDate(0, 0, 1)
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

		entry.Info("handling hashes")
		e.HandleHashes(c, hashes)
		entry.Info("finished with hashes")

		attempts = 1
		date = date.AddDate(0, 0, 1)
		_, err = e.Redis.Set(redisKey, date.Format("20060102"), -1).Result()
		if err != nil {
			e.Logger.WithError(err).Error("redis returned invalid response while setting egress date")
		}
	}

}

func (e *Egressor) HandleHashes(c *cli.Context, hashes map[string]string) {

	// Make Sure the Redis Server is still alive and nothing has happened to it
	pong, err := e.Redis.Ping().Result()
	if err != nil {
		e.Logger.WithError(err).Fatal("unable to ping redis server")
	}
	// Make sure that redis returned pong to ping
	if pong != "PONG" {
		e.Logger.WithField("pong", pong).Fatal("unexpected response to redis server ping received")
	}
	// Lets retrieve the channel from the cli
	channel := c.String("channel")

	// Lets get the most recent record from the end of the set to determine the score to use
	results, err := e.Redis.ZRevRangeByScoreWithScores(channel, redis.ZRangeBy{Min: "-inf", Max: "+inf", Count: 1}).Result()
	if err != nil {
		e.Logger.WithError(err).Fatal("unable to get max score of redis z range")
	}

	// If we received more than one result, something is wrong and we need to bail
	if len(results) > 1 {
		e.Logger.WithError(err).Fatal("unable to determine score")
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
			e.Logger.WithFields(logrus.Fields{
				"id":   id,
				"hash": hash,
			}).Error("unable to marshal id and hash for pubsub")
			continue
		}

		e.Redis.ZAdd(channel, redis.Z{Score: score, Member: msg})
		dispatched++

	}

	for {

		count, err := e.Redis.ZCount(channel, "-inf", "+inf").Result()
		if err != nil {
			e.Logger.WithError(err).Fatal("unable to get count of redis zset")
		}

		e.Logger.WithFields(logrus.Fields{
			"total":         len(hashes),
			"dispatched":    dispatched,
			"remaining":     len(hashes) - dispatched,
			"current_queue": count,
		}).Info("queue status")
		if count < 200 {
			return
		}

		time.Sleep(time.Minute * 1)

	}

}
