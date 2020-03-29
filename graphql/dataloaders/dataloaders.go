package dataloaders

import (
	"time"

	"github.com/eveisesi/neo/graphql/dataloaders/generated"
)

const defaultWait = time.Millisecond * 10
const defaultMaxBatch = 100

type Loaders struct {
	AllianceLoader          *generated.AllianceLoader
	CharacterLoader         *generated.CharacterLoader
	CorporationLoader       *generated.CorporationLoader
	KillmailAttackersLoader *generated.KillmailAttackersLoader
	KillmailItemsLoader     *generated.KillmailItemsLoader
	KillmailVictimLoader    *generated.KillmailVictimLoader
	TypeLoader              *generated.TypeLoader
	SolarSystemLoader       *generated.SolarSystemLoader
}
