package killmail

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/korovkin/limiter"

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
		Recalculate(ctx context.Context, db *sqlx.DB)
		HistoryExporter(mindate, maxdate string) error
		Importer(gLimit, gSleep int64) error
		Websocket() error
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
	conn *websocket.Conn
	err  error
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

func (s *service) Recalculate(ctx context.Context, db *sqlx.DB) {

	nextID := null.NewUint64(83288286, true)
	limiter := limiter.NewConcurrencyLimiter(40)
	var count int
	err = db.Get(&count, `SELECT COUNT(id) FROM killmails where id > ?`, nextID.Uint64)

	s.logger.WithField("remaining", count).Println()

	for {

		killmails, err := s.KillmailGTID(ctx, nextID)
		if err != nil {
			s.logger.WithError(err).Fatal("failed to fetch killmails")
		}

		if len(killmails) == 0 {
			break
		}

		for _, killmail := range killmails {
			limiter.ExecuteWithTicket(func(workerID int) {
				s.processKillmailRecalc(killmail, workerID)
			})
		}
		s.logger.WithField("currentID", nextID.Uint64).Info("batch update successful")

		nextID = null.NewUint64(killmails[len(killmails)-1].ID, true)

	}

	s.logger.Info("done updating killmails")

}
