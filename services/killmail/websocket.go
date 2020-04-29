package killmail

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *service) Websocket(inputchan string) error {

	channel = inputchan

	for {
		for {
			// Attempt to connect to Websocket
			conn, err = s.connect()
			if err != nil {
				s.logger.WithError(err).Error("failed to establish connection to websocket")
				time.Sleep(time.Second * 2)
				continue
			}

			break
		}

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if err, ok := err.(*websocket.CloseError); ok {
					if err.Code == 1000 {
						s.logger.Info("gracefully closing connection with websocket")
						return nil
					}

					s.logger.WithError(err).Error("error enconnected. Attempting to disconnect and reconnect")

					break
				}
				eerr := conn.Close()
				if eerr != nil {
					s.logger.WithError(eerr).Error("unable to close connection after error")
				}
				break
			}
			go s.handleWSSPayload(message)
		}

		s.logger.Info("bottom of parent loop. Sleep and attemp to reconnect")
		time.Sleep(time.Second * 2)
	}

}

func (s *service) handleWSSPayload(msg []byte) {

	var message WSPayload
	err := json.Unmarshal(msg, &message)
	if err != nil {
		s.logger.WithError(err).Fatal("failed to unmarhal message into message struct")
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
	score += 1

	payload, _ := json.Marshal(struct {
		ID   string `json:"id"`
		Hash string `json:"hash"`
	}{
		ID:   strconv.FormatUint(uint64(message.KillID), 10),
		Hash: message.Hash,
	})

	_, err = s.redis.ZAdd(channel, redis.Z{Score: score, Member: payload}).Result()
	if err != nil {
		s.logger.WithError(err).Fatal("something is wrong")
	}

	s.logger.WithFields(logrus.Fields{
		"id":   message.KillID,
		"hash": message.Hash,
	}).Info("message received and queued successfully")
}

func (s *service) connect() (*websocket.Conn, error) {
	address := url.URL{
		Scheme: "wss",
		Host:   "zkillboard.com:2096",
	}

	body := struct {
		Action  string `json:"action"`
		Channel string `json:"channel"`
	}{
		Action:  "sub",
		Channel: "all:*",
	}

	msg, err := json.Marshal(body)
	if err != nil {
		s.logger.WithError(err).WithField("request", body).Error("Encoutered Error Attempting marshal sub message")
		return nil, err
	}

	s.logger.WithField("address", address.String()).Info("attempting to connect to websocket")

	c, _, err := websocket.DefaultDialer.Dial(address.String(), nil)
	if err != nil {
		return nil, err
	}

	s.logger.Info("successfully connected to websocket. Sending Initial Msg")

	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send initial message")
	}

	return c, err
}
