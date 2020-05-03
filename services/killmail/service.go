package killmail

import (
	"context"
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
)

type (
	Service interface {
		// Business Appliances
		// Recalculate(ctx context.Context, db *sqlx.DB)
		HistoryExporter(mindate, maxdate string) error
		Importer(gLimit, gSleep int64) error
		Websocket() error

		// Killmails
		Killmail(ctx context.Context, id uint64, hash string) (*neo.Killmail, error)
		RecentKillmails(ctx context.Context, page int) ([]*neo.Killmail, error)
		KillmailsByCharacterID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByCorporationID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByAllianceID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByShipID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)

		// Attackers
		AttackersByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailAttacker, error)

		// Items
		ItemsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailItem, error)

		// Victim
		VictimsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailVictim, error)

		// MVKs/MVLs
		MVKAll(ctx context.Context, age, limit int) ([]*neo.Killmail, error)
		MVKByCharacterID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByCharacterID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByCorporationID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByCorporationID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByAllianceID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByAllianceID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByShipID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByShipID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
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
		killmails   neo.KillmailRepository
		attackers   neo.KillmailAttackerRepository
		items       neo.KillmailItemRepository
		victim      neo.KillmailVictimRepository
		mvks        neo.MVRepository
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
	killmails neo.KillmailRepository,
	attackers neo.KillmailAttackerRepository,
	items neo.KillmailItemRepository,
	victim neo.KillmailVictimRepository,
	mvks neo.MVRepository,
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
		killmails,
		attackers,
		items,
		victim,
		mvks,
	}
}

// func (s *service) Recalculate(ctx context.Context, db *sqlx.DB) {

// 	nextID := null.NewUint64(83288286, true)
// 	limiter := limiter.NewConcurrencyLimiter(40)
// 	var count int
// 	err = db.Get(&count, `SELECT COUNT(id) FROM killmails where id > ?`, nextID.Uint64)

// 	s.logger.WithField("remaining", count).Println()

// 	for {

// 		killmails, err := s.killmails.GTID(ctx, nextID)
// 		if err != nil {
// 			s.logger.WithError(err).Fatal("failed to fetch killmails")
// 		}

// 		if len(killmails) == 0 {
// 			break
// 		}

// 		for _, killmail := range killmails {
// 			limiter.ExecuteWithTicket(func(workerID int) {
// 				s.processKillmailRecalc(killmail, workerID)
// 			})
// 		}
// 		s.logger.WithField("currentID", nextID.Uint64).Info("batch update successful")

// 		nextID = null.NewUint64(killmails[len(killmails)-1].ID, true)

// 	}

// 	s.logger.Info("done updating killmails")

// }
