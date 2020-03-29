//go:generate go run github.com/vektah/dataloaden AllianceLoader uint64 *github.com/eveisesi/neo.Alliance
//go:generate go run github.com/vektah/dataloaden CharacterLoader uint64 *github.com/eveisesi/neo.Character
//go:generate go run github.com/vektah/dataloaden CorporationLoader uint64 *github.com/eveisesi/neo.Corporation
//go:generate go run github.com/vektah/dataloaden KillmailAttackersLoader uint64 []*github.com/eveisesi/neo.KillmailAttacker
//go:generate go run github.com/vektah/dataloaden KillmailItemsLoader *github.com/eveisesi/neo.KillmailItemLoader []*github.com/eveisesi/neo.KillmailItem
//go:generate go run github.com/vektah/dataloaden KillmailVictimLoader uint64 *github.com/eveisesi/neo.KillmailVictim
//go:generate go run github.com/vektah/dataloaden TypeLoader uint64 *github.com/eveisesi/neo.Type
//go:generate go run github.com/vektah/dataloaden SolarSystemLoader uint64 *github.com/eveisesi/neo.SolarSystem

package generated
