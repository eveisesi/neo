package killmail

import (
	"net/http"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/esi"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/market"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/null"
)

type (
	Service interface {
		// WebsocketExporter(channel string) error
		HistoryExporter(channel string, cDate null.String) error
		Importer(channel string, gLimit, gSleep int64) error
		neo.KillmailRespository
	}

	Message struct {
		ID   string `json:"id"`
		Hash string `json:"hash"`
	}

	service struct {
		client      *http.Client
		redis       *redis.Client
		esi         *esi.Client
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

func NewService(
	client *http.Client,
	redis *redis.Client,
	esi *esi.Client,
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
