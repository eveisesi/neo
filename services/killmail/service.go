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
	"github.com/eveisesi/neo/services/tracker"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type (
	Service interface {
		// Business Appliances
		HistoryExporter(mindate, maxdate string, datehold bool, threshold int64) error
		Importer(gLimit, gSleep int64) error
		Websocket() error
		Recalculator(gLimit int64)
		RecalculatorDispatcher(limit, trigger int64, after uint64)

		// Killmails
		Killmail(ctx context.Context, id uint64, hash string) (*neo.Killmail, error)
		RecentKillmails(ctx context.Context, page int) ([]*neo.Killmail, error)
		KillmailsByCharacterID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByCorporationID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByAllianceID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByShipID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByShipGroupID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsBySystemID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByConstellationID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)
		KillmailsByRegionID(ctx context.Context, id uint64, page int) ([]*neo.Killmail, error)

		// Attackers
		AttackersByKillmailID(ctx context.Context, id uint64, hash string) ([]*neo.KillmailAttacker, error)
		AttackersByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailAttacker, error)

		// Items
		ItemsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailItem, error)

		// Victim
		VictimByKillmailID(ctx context.Context, id uint64, hash string) (*neo.KillmailVictim, error)
		VictimsByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailVictim, error)

		// MVKs/MVLs
		MVKAll(ctx context.Context, age, limit int) ([]*neo.Killmail, error)
		MVKByCharacterID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByCharacterID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByCorporationID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByCorporationID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByAllianceID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByAllianceID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByShipID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByShipID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVLByShipGroupID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByShipGroupID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)

		MVKBySystemID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByConstellationID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
		MVKByRegionID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error)
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
		tracker     tracker.Service
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

	// Services
	character character.Service,
	corporation corporation.Service,
	alliance alliance.Service,
	universe universe.Service,
	market market.Service,
	tracker tracker.Service,

	txn neo.Starter,

	// Repositories
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
		tracker,
		txn,
		killmails,
		attackers,
		items,
		victim,
		mvks,
	}
}

func ChunkSliceKillmails(slice []*neo.Killmail, size int) [][]*neo.Killmail {

	var chunk = make([][]*neo.Killmail, 0)
	if len(slice) <= size {
		chunk = append(chunk, slice)
		slice = nil
		return chunk
	}

	for x := 0; x < len(slice); x += size {
		end := x + size

		if end > len(slice) {
			end = len(slice)
		}

		chunk = append(chunk, slice[x:end])
	}

	slice = nil

	return chunk

}
