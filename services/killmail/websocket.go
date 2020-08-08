package killmail

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/eveisesi/neo"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *service) Websocket() error {

	for {
		for {
			txn := s.newrelic.StartTransaction("connect to zkillboard")
			// Attempt to connect to Websocket

			ctx := newrelic.NewContext(context.Background(), txn)

			conn, err = s.connect(ctx)
			if err != nil {
				txn.NoticeError(err)
				s.logger.WithError(err).Error("failed to establish connection to websocket")
				time.Sleep(time.Second * 2)
				continue
			}
			txn.End()
			break
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if err, ok := err.(*websocket.CloseError); ok {
					if err.Code == 1000 {
						s.logger.Info("gracefully closing connection with websocket")
						break
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

			var message map[string]interface{}
			err = json.Unmarshal(msg, &message)
			if err != nil {
				s.logger.WithError(err).WithField("msg", string(msg)).Error("failed to unmarhal message into message struct")
			}

			if _, ok := message["killID"]; !ok {
				continue
			}
			if _, ok := message["hash"]; !ok {
				continue
			}

			neoMsg := &neo.Message{
				ID:   uint64(message["killID"].(float64)),
				Hash: message["hash"].(string),
			}

			go s.DispatchPayload(neoMsg)
		}

		s.logger.Info("bottom of parent loop. Sleep and attemp to reconnect")
		time.Sleep(time.Second * 2)
	}

}

func (s *service) DispatchPayload(msg *neo.Message) {
	txn := s.newrelic.StartTransaction("listen")
	defer txn.End()

	txn.AddAttribute("id", msg.ID)
	txn.AddAttribute("hash", msg.Hash)

	ctx := newrelic.NewContext(context.Background(), txn)

	payload, err := json.Marshal(msg)
	if err != nil {
		txn.NoticeError(err)
		s.logger.WithContext(ctx).WithError(err).Error("unable to marshal WSSPayload")
		return
	}
	_, err = s.redis.WithContext(ctx).ZAdd(neo.QUEUES_KILLMAIL_PROCESSING, &redis.Z{Score: float64(msg.ID), Member: string(payload)}).Result()
	if err != nil {
		txn.NoticeError(err)
		s.logger.WithContext(ctx).WithError(err).WithField("payload", string(payload)).Error("unable to push killmail to processing queue")
		return
	}

	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"id":   msg.ID,
		"hash": msg.Hash,
	}).Info("payload dispatched successfully")
}

func (s *service) connect(ctx context.Context) (*websocket.Conn, error) {
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
		newrelic.FromContext(ctx).NoticeError(err)
		s.logger.WithContext(ctx).WithError(err).WithField("request", body).Error("Encoutered Error Attempting marshal sub message")
		return nil, err
	}

	s.logger.WithContext(ctx).WithField("address", address.String()).Info("attempting to connect to websocket")
	c, _, err := websocket.DefaultDialer.DialContext(ctx, address.String(), nil)
	if err != nil {
		newrelic.FromContext(ctx).NoticeError(err)
		return nil, err
	}

	s.logger.WithContext(ctx).Info("successfully connected to websocket. Sending Initial Msg")

	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		newrelic.FromContext(ctx).NoticeError(err)
		return nil, errors.Wrap(err, "failed to send initial message")
	}

	return c, err
}
