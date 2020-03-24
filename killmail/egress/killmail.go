package egress

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"

	"github.com/ddouglas/killboard"

	"github.com/urfave/cli"

	core "github.com/ddouglas/killboard/app"
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

	date, err := time.Parse("20060102", c.String("date"))
	if err != nil {
		e.Logger.WithError(err).Fatal("unable to parse provided date")
	}

	attempts := 1
	for {

		if attempts > 3 {
			return cli.NewExitError("maximum allowed attempts reeached", 1)
		}

		entry := e.Logger.WithField("date", date.Format("20060102"))

		uri := fmt.Sprintf(killboard.ZKILLBOARD_HISTORY_API, date.Format("20060102"))

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

	}

}

func (e *Egressor) HandleHashes(c *cli.Context, hashes map[string]string) {

	// Make Sure the Redis Server is still alive and nothing has happened to it
	pong, err := e.Redis.Ping().Result()
	if err != nil {
		e.Logger.WithError(err).Fatal("unable to ping redis server")
	}

	if pong != "PONG" {
		e.Logger.WithField("pong", pong).Fatal("unexpected response to redis server ping received")
	}
	channel := c.String("channel")

	results, err := e.Redis.ZRevRangeByScoreWithScores(channel, redis.ZRangeBy{Min: "-inf", Max: "+inf", Count: 1}).Result()
	if err != nil {
		e.Logger.WithError(err).Fatal("unable to get max score of redis z range")
	}
	if len(results) > 1 {
		e.Logger.WithError(err).Fatal("unable to determine score")
	}
	score := float64(0)
	if len(results) == 1 {
		score = results[0].Score
	}

	dispatched := 0

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

		for {
			count, err := e.Redis.ZCount("killhashes", "-inf", "+inf").Result()
			if err != nil {
				e.Logger.WithError(err).Fatal("unable to get count of redis zset")
			}

			if math.Mod(float64(dispatched), 100) == float64(0) {
				e.Logger.WithFields(logrus.Fields{
					"total":         len(hashes),
					"dispatched":    dispatched,
					"remaining":     len(hashes) - dispatched,
					"current_queue": count,
				}).Info("dispatched hold")
			}

			if count < 1000 {
				break
			}

			time.Sleep(time.Second * 5)
		}

	}

}
