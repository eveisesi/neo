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
const REDIS_ESI_TRACKING_FAILED = "esi:tracking:failed"
const REDIS_ESI_TRACKING_SUCCESS = "esi:tracking:success"

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
