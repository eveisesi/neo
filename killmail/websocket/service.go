package websocket

import (
	"encoding/json"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	core "github.com/eveisesi/neo/app"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli"
)

type (
	Listener struct {
		*core.App
	}
	Message struct {
		Action        string `json:"action"`
		KillID        uint   `json:"killID"`
		CharacterID   uint64 `json:"character_id"`
		CorporationID uint   `json:"corporation_id"`
		AllianceID    uint   `json:"alliance_id"`
		ShipTypeID    uint   `json:"ship_type_id"`
		URL           string `json:"url"`
		Hash          string `json:"hash"`
	}
)

var (
	wg      sync.WaitGroup
	channel string
)

func Action(c *cli.Context) error {

	listener := &Listener{
		core.New(),
	}
	// // Lets retrieve the channel from the cli
	channel = c.String("channel")

	listener.Logger.Info("Starting websocket listener")

	wg.Add(1)
	go func() {
		connected := make(chan bool, 1)
		disconnected := make(chan bool, 1)
		done := make(chan bool, 1)
		stream := make(chan []byte)

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		defer wg.Done()

		wg.Add(1)
		go listener.Listen(stream, interrupt, connected, disconnected, done)
		for {
			select {
			case <-done:
				listener.Logger.Info("Done in Action")
				listener.Logger.Infof("Number of Go Routines Remaining: %d", runtime.NumGoroutine())
				return
			case <-disconnected:
				listener.Logger.Error("Supervisor: Disconnected from Websocket. Attempting to reconnect")
				// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				// go func(msg string) {
				// 	_, _ = listener.DGO.ChannelMessageSend("394991263344230411", msg)
				// 	return
				// }(msg)
				time.Sleep(2 * time.Second)
				wg.Add(1)
				go listener.Listen(stream, interrupt, connected, disconnected, done)
			case <-connected:
				listener.Logger.Info("Supervisor: Connected to Websocket")
			}
		}

	}()

	listener.Logger.Info("Supervisor Routine Launched Successfully")
	wg.Wait()
	listener.Logger.Info("Bye")
	return nil

}

func (r *Listener) Listen(stream chan []byte, interrupt chan os.Signal, connected, disconnected, done chan bool) {
	defer wg.Done()

	c, err := r.connect()
	if err != nil {
		disconnected <- true
		r.Logger.WithError(err).Error("failed to establish connection to websocket")
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if err, ok := err.(*gorilla.CloseError); ok {
					// Error Code 1000 is the response to a close message.
					if err.Code == 1000 {
						done <- true
						return
					}

					r.Logger.WithError(err).Error("error enconnected. Attempting to disconnect and reconnect")
					eerr := c.Close()
					if eerr != nil {
						r.Logger.WithError(eerr).Error("unable to close connection after error")
					}

					disconnected <- true
					return
				}
				r.Logger.WithError(err).Fatal("unknown error encountered. Crashing container")
			}

			stream <- message
		}
	}()

	connected <- true

	for {
		select {
		case msg := <-stream:
			wg.Add(1)
			go r.processMessage(msg)
		case <-interrupt:
			r.Logger.Info("Interrupted")
			err := c.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, ""))
			if err != nil {
				r.Logger.WithError(err).Fatal("Failed to write close message")
			}
			done <- true
			return
		}
	}
}

func (r *Listener) connect() (*websocket.Conn, error) {
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
		r.Logger.WithError(err).WithField("request", body).Error("Encoutered Error Attempting marshal sub message")
		return nil, err
	}

	r.Logger.WithField("address", address.String()).Info("attempting to connect to websocket")

	c, _, err := gorilla.DefaultDialer.Dial(address.String(), nil)
	if err != nil {
		r.Logger.WithError(err).Fatal("failed to connect to websocket")
	}

	r.Logger.Info("successfully connected to websocket. Sending Initial Msg")

	err = c.WriteMessage(gorilla.TextMessage, msg)
	if err != nil {
		r.Logger.WithError(err).Fatal("failed to send initial message")
	}

	return c, err
}

func (r *Listener) processMessage(msg []byte) {
	defer wg.Done()

	var message Message
	err := json.Unmarshal(msg, &message)
	if err != nil {
		r.Logger.WithError(err).Fatal("failed to unmarhal message into message struct")
	}

	// Lets get the most recent record from the end of the set to determine the score to use
	results, err := r.Redis.ZRevRangeByScoreWithScores(channel, redis.ZRangeBy{Min: "-inf", Max: "+inf", Count: 1}).Result()
	if err != nil {
		r.Logger.WithError(err).Fatal("unable to get max score of redis z range")
	}

	// If we received more than one result, something is wrong and we need to bail
	if len(results) > 1 {
		r.Logger.WithError(err).Fatal("unable to determine score")
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

	_, err = r.Redis.ZAdd(channel, redis.Z{Score: score, Member: payload}).Result()
	if err != nil {
		r.Logger.WithError(err).Fatal("something is wrong")
	}

	r.Logger.WithFields(logrus.Fields{
		"id":   message.KillID,
		"hash": message.Hash,
	}).Info("message received and queued successfully")
}
