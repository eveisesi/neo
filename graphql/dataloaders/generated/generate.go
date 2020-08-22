//go:generate go run github.com/vektah/dataloaden AllianceLoader uint *github.com/eveisesi/neo.Alliance
//go:generate go run github.com/vektah/dataloaden CharacterLoader uint64 *github.com/eveisesi/neo.Character
//go:generate go run github.com/vektah/dataloaden CorporationLoader uint *github.com/eveisesi/neo.Corporation
//go:generate go run github.com/vektah/dataloaden KillmailAttackersLoader uint []*github.com/eveisesi/neo.KillmailAttacker
//go:generate go run github.com/vektah/dataloaden KillmailItemsLoader uint []*github.com/eveisesi/neo.KillmailItem
//go:generate go run github.com/vektah/dataloaden KillmailVictimLoader uint *github.com/eveisesi/neo.KillmailVictim
//go:generate go run github.com/vektah/dataloaden TypeLoader uint *github.com/eveisesi/neo.Type
//go:generate go run github.com/vektah/dataloaden SolarSystemLoader uint *github.com/eveisesi/neo.SolarSystem
//go:generate go run github.com/vektah/dataloaden TypeAttributeLoader uint []*github.com/eveisesi/neo.TypeAttribute
//go:generate go run github.com/vektah/dataloaden TypeFlagLoader uint *github.com/eveisesi/neo.TypeFlag
//go:generate go run github.com/vektah/dataloaden TypeGroupLoader uint *github.com/eveisesi/neo.TypeGroup
//go:generate go run github.com/vektah/dataloaden TypeCategoryLoader uint *github.com/eveisesi/neo.TypeCategory
//go:generate go run github.com/vektah/dataloaden ConstellationLoader uint *github.com/eveisesi/neo.Constellation
//go:generate go run github.com/vektah/dataloaden RegionLoader uint *github.com/eveisesi/neo.Region

package generated
