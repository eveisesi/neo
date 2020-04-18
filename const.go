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
