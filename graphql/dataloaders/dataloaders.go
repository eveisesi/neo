package dataloaders

import (
	"time"

	"github.com/eveisesi/neo/graphql/dataloaders/generated"
)

const defaultWait = time.Millisecond * 10
const defaultMaxBatch = 100

type Loaders struct {
	AllianceLoader      *generated.AllianceLoader
	CharacterLoader     *generated.CharacterLoader
	ConstellationLoader *generated.ConstellationLoader
	CorporationLoader   *generated.CorporationLoader
	RegionLoader        *generated.RegionLoader
	SolarSystemLoader   *generated.SolarSystemLoader
	TypeLoader          *generated.TypeLoader
	TypeAttributeLoader *generated.TypeAttributeLoader
	TypeCategoryLoader  *generated.TypeCategoryLoader
	TypeFlagLoader      *generated.TypeFlagLoader
	TypeGroupLoader     *generated.TypeGroupLoader
}
