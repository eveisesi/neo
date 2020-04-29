package killmail

import (
	"net/http"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/market"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/null"
)

type (
	Service interface {
		// WebsocketExporter(channel string) error
		HistoryExporter(channel string, cDate null.String) error
		Importer(channel string, gLimit, gSleep int64) error
		Websocket(channel string) error
		neo.KillmailRespository
	}

	Message struct {
		ID   string `json:"id"`
		Hash string `json:"hash"`
	}

	WSPayload struct {
		Action        string `json:"action"`
		KillID        uint   `json:"killID"`
		CharacterID   uint64 `json:"character_id"`
		CorporationID uint   `json:"corporation_id"`
		AllianceID    uint   `json:"alliance_id"`
		ShipTypeID    uint   `json:"ship_type_id"`
		URL           string `json:"url"`
		Hash          string `json:"hash"`
	}

	service struct {
		client      *http.Client
		redis       *redis.Client
		esi         esi.Service
		logger      *logrus.Logger
		config      *neo.Config
		character   character.Service
		corporation corporation.Service
		alliance    alliance.Service
		universe    universe.Service
		market      market.Service
		txn         neo.Starter
		neo.KillmailRespository
	}
)

var (
	conn    *websocket.Conn
	err     error
	channel string
)

func NewService(
	client *http.Client,
	redis *redis.Client,
	esi esi.Service,
	logger *logrus.Logger,
	config *neo.Config,
	character character.Service,
	corporation corporation.Service,
	alliance alliance.Service,
	universe universe.Service,
	market market.Service,
	txn neo.Starter,
	killmail neo.KillmailRespository,
) Service {
	return &service{
		client,
		redis,
		esi,
		logger,
		config,
		character,
		corporation,
		alliance,
		universe,
		market,
		txn,
		killmail,
	}
}
