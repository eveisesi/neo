package neo

import "errors"

const ZKILLBOARD_URL = "https://zkillboard.com"

// Intentionally leaving the %s token. This is meant to be run through Sprintf to replace the token with a date
const ZKILLBOARD_HISTORY_API = "https://zkillboard.com/api/history/%s.json"

const ESI_URL = "https://esi.evetech.net"
const SSO_URL = "https://login.eveonline.com"

// Errors
var ErrRedisNil = errors.New("redis: nil")
var ErrEsiMaxAttempts = errors.New("max attempts exceeded")
var ErrEsiTypeNotFound = errors.New("type not found")

// ESI Timestamp Format
const ESI_EXPIRES_HEADER_FORMAT = "Mon, 02 Jan 2006 15:04:05 MST"

// REDIS KEY
const REDIS_ESI_ERROR_COUNT = "esi:error:count"
const REDIS_ESI_ERROR_RESET = "esi:error:reset"
const REDIS_ESI_TRACKING_STATUS = "esi:tracking:status"

const REDIS_ESI_TRACKING_OK = "neo:esi:tracking:ok"                     // 200
const REDIS_ESI_TRACKING_NOT_MODIFIED = "neo:esi:tracking:not_modified" // 304
const REDIS_ESI_TRACKING_CALM_DOWN = "neo:esi:tracking:calm_down"       // 420
const REDIS_ESI_TRACKING_4XX = "neo:esi:tracking:4xx"                   // Does not include 420s. Those are in the calm down set
const REDIS_ESI_TRACKING_5XX = "neo:esi:tracking:5xx"

// Locked is something that one of our crons can set to signal to all the other jobs that no matter what, the current status is locked and cannot be unlocked with out that same cron removing the lock
const REDIS_ESI_TRACKING_STATUS_LOCKED = "esi:tracking:status:lock"

// Status Const
const COUNT_STATUS_DOWNTIME = 3
const COUNT_STATUS_RED = 2
const COUNT_STATUS_YELLOW = 1
const COUNT_STATUS_GREEN = 0

// TQ Const
const TQ_PLAYER_COUNT = "esi:tq:player_count"
const TQ_VIP_MODE = "esi:tq:vip"

// ZKB_HISTORY_DATE Const
const ZKB_HISTORY_DATE = "zkb:history:date"

// Actual NEO keys now lol
const REDIS_ALLIANCE = "neo:alliance:%d"
const REDIS_CHARACTER = "neo:character:%d"
const REDIS_CORPORATION = "neo:corporation:%d"
const REDIS_KILLMAIL = "neo:killmail:%d:%s"
const REDIS_KILLMAIL_ATTACKERS = "neo:killmail:%d:%s:attackers"
const REDIS_KILLMAIL_VICTIM = "neo:killmail:%d:%s:victim"
const REDIS_KILLMAIL_VICTIM_ITEMS = "neo:killmail:%d:%s:victim:items"
const REDIS_KILLMAILS_BY_ENTITY = "neo:killmails:${type}:${id}:${page}"
const REDIS_BLUEPRINT_MATERIALS = "neo:blueprint:materials:%d"
const REDIS_BLUEPRINT_PRODUCT = "neo:blueprint:product:%d"
const REDIS_BLUEPRINT_PRODUCTTYPEID = "neo:blueprint:producttypeid:%d"
const REDIS_CONSTELLATION = "neo:constellation:%d"
const REDIS_REGION = "neo:region:%d"
const REDIS_SYSTEM = "neo:system:%d"
const REDIS_TYPE = "neo:type:%d"
const REDIS_TYPE_CATEGORY = "neo:type:category:%d"
const REDIS_TYPE_ATTRIBUTES = "neo:type:attributes:%d"
const REDIS_TYPE_FLAG = "neo:type:flag:%d"
const REDIS_TYPE_GROUP = "neo:type:group:%d"
const REDIS_GRAPHQL_APQ_CACHE = "neo:graphql:apq"

// NEO Queues
const QUEUES_KILLMAIL_PROCESSING = "neo:killmails:processing"
const QUEUE_KILLMAIL_RECALCULATE = "neo:killmails:recalculate"

// NEO Notifications Redis PubSub Channel
const REDIS_NOTIFICATION_PUBSUB = "neo:notifications"

// FlagIDs to FittingSlots

var SLOT_TO_FLAGIDS = map[string]map[uint64]bool{
	"low": map[uint64]bool{
		11: true, 12: true, 13: true, 14: true, 15: true, 16: true, 17: true, 18: true,
	},

	"mid": map[uint64]bool{
		19: true, 20: true, 21: true, 22: true, 23: true, 24: true, 25: true, 26: true,
	},

	"hi": map[uint64]bool{
		27: true, 28: true, 29: true, 30: true, 31: true, 32: true, 33: true, 34: true,
	},

	"drone": map[uint64]bool{
		87: true,
	},

	"implants": map[uint64]bool{
		89: true,
	},

	"rigs": map[uint64]bool{
		92: true, 93: true, 94: true, 95: true, 96: true, 97: true, 98: true, 99: true,
	},

	"cargo": map[uint64]bool{
		5: true,
	},

	"subsystem": map[uint64]bool{
		125: true, 126: true, 127: true, 128: true, 129: true, 130: true, 131: true, 132: true,
	},

	"fighter_tubes": map[uint64]bool{
		159: true, 160: true, 161: true, 162: true, 163: true,
	},

	"structure_service": map[uint64]bool{
		164: true, 165: true, 166: true, 167: true, 168: true, 169: true, 170: true, 171: true,
	},
}

const KILLMAILS_PER_PAGE = 50

var ALLOWED_SHIP_GROUPS = []uint64{
	25,   // Frigate
	26,   // Cruiser
	27,   // Battleship
	28,   // Industrial
	30,   // Titan
	31,   // Shuttle
	237,  // Corvette
	324,  // Assault Frigate
	358,  // Heavy Assault Cruiser
	380,  // Deep Space Transport
	419,  // Combat Battlecruiser
	420,  // Destroyer
	463,  // Mining Barge
	485,  // Dreadnought
	513,  // Freighter
	540,  // Command Ship
	541,  // Interdictor
	543,  // Exhumer
	547,  // Carrier
	659,  // Supercarrier
	830,  // Covert Ops
	831,  // Interceptor
	832,  // Logistics
	833,  // Force Recon Ship
	834,  // Stealth Bomber
	883,  // Capital Industrial Ship
	893,  // Electronic Attack Ship
	894,  // Heavy Interdiction Cruiser
	898,  // Black Ops
	900,  // Marauder
	902,  // Jump Freighter
	906,  // Combat Recon Ship
	941,  // Industrial Command Ship
	963,  // Strategic Cruiser
	1022, // Prototype Exploration Ship
	1201, // Attack Battlecruiser
	1202, // Blockade Runner
	1283, // Expedition Frigate
	1305, // Tactical Destroyer
	1404, // Engineering Complex
	1406, // Refinery
	1408, // Upwell Jump Gate
	1527, // Logistics Frigate
	1534, // Command Destroyer
	1538, // Force Auxiliary
	1657, // Citadel
	1972, // Flag Cruiser
	2016, // Upwell Cyno Jammer
	2017, // Upwell Cyno Beacon
}
