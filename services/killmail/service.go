package killmail

import (
	"context"
	"net/http"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/backup"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/market"
	"github.com/eveisesi/neo/services/tracker"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type (
	Service interface {
		ProcessMessage(ctx context.Context, entry *logrus.Entry, message []byte) (*neo.Killmail, error)
		// Business Appliances
		HistoryExporter(mindate, maxdate string, datehold bool, threshold int64) error
		Importer(gLimit, gSleep int64) error
		Websocket() error
		// Recalculator(gLimit int64)
		// RecalculatorDispatcher(limit, trigger int64, after uint64)
		DispatchPayload(msg *neo.Message)

		// Killmails
		Killmail(ctx context.Context, id uint) (*neo.Killmail, error)
		// Killmails(ctx context.Context, coreMods []neo.Modifier, vicMods []neo.Modifier, attMods []neo.Modifier) ([]*neo.Killmail, error)
		FullKillmail(ctx context.Context, id uint, withNames bool) (*neo.Killmail, error)
		RecentKillmails(ctx context.Context, page int) ([]*neo.Killmail, error)
		KillmailsByCharacterID(ctx context.Context, id uint64, after uint) ([]*neo.Killmail, error)
		KillmailsByCorporationID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error)
		KillmailsByAllianceID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error)
		KillmailsByShipID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error)
		KillmailsByShipGroupID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error)
		KillmailsBySystemID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error)
		KillmailsByConstellationID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error)
		KillmailsByRegionID(ctx context.Context, id uint, after uint) ([]*neo.Killmail, error)

		// Attackers
		AttackersByKillmailID(ctx context.Context, id uint) ([]*neo.KillmailAttacker, error)
		AttackersByKillmailIDs(ctx context.Context, ids []uint) ([]*neo.KillmailAttacker, error)

		// Items
		ItemsByKillmailIDs(ctx context.Context, ids []uint) ([]*neo.KillmailItem, error)

		// Victim
		VictimByKillmailID(ctx context.Context, id uint) (*neo.KillmailVictim, error)
		VictimsByKillmailIDs(ctx context.Context, ids []uint) ([]*neo.KillmailVictim, error)

		MostValuable(ctx context.Context, column string, id, age, limit uint) ([]*neo.Killmail, error)
		MostValuableKills(ctx context.Context, column string, id uint64, age, limit uint) ([]*neo.Killmail, error)
		MostValuableLosses(ctx context.Context, column string, id uint64, age, limit uint) ([]*neo.Killmail, error)
	}

	WSPayload struct {
		Action        string `json:"action"`
		KillID        uint64 `json:"killID"`
		CharacterID   uint64 `json:"character_id"`
		CorporationID uint64 `json:"corporation_id"`
		AllianceID    uint64 `json:"alliance_id"`
		ShipTypeID    uint64 `json:"ship_type_id"`
		URL           string `json:"url"`
		Hash          string `json:"hash"`
	}

	service struct {
		client      *http.Client
		redis       *redis.Client
		newrelic    *newrelic.Application
		esi         esi.Service
		logger      *logrus.Logger
		config      *neo.Config
		backup      backup.Service
		character   character.Service
		corporation corporation.Service
		alliance    alliance.Service
		universe    universe.Service
		market      market.Service
		tracker     tracker.Service
		killmails   neo.KillmailRepository
	}
)

var (
	conn *websocket.Conn
	err  error
)

func NewService(
	client *http.Client,
	redis *redis.Client,
	nr *newrelic.Application,
	esi esi.Service,
	logger *logrus.Logger,
	config *neo.Config,
	backup backup.Service,

	// Services
	character character.Service,
	corporation corporation.Service,
	alliance alliance.Service,
	universe universe.Service,
	market market.Service,
	tracker tracker.Service,

	// Repositories
	killmails neo.KillmailRepository,
	// attackers neo.KillmailAttackerRepository,
	// items neo.KillmailItemRepository,
	// victim neo.KillmailVictimRepository,
) Service {
	return &service{
		client,
		redis,
		nr,
		esi,
		logger,
		config,
		backup,
		character,
		corporation,
		alliance,
		universe,
		market,
		tracker,
		killmails,
		// attackers,
		// items,
		// victim,
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
