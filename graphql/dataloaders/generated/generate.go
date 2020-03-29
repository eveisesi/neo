//go:generate go run github.com/vektah/dataloaden AllianceLoader uint64 *github.com/ddouglas/neo.Alliance
//go:generate go run github.com/vektah/dataloaden CharacterLoader uint64 *github.com/ddouglas/neo.Character
//go:generate go run github.com/vektah/dataloaden CorporationLoader uint64 *github.com/ddouglas/neo.Corporation
//go:generate go run github.com/vektah/dataloaden KillmailAttackersLoader uint64 []*github.com/ddouglas/neo.KillmailAttacker
//go:generate go run github.com/vektah/dataloaden KillmailItemsLoader *github.com/ddouglas/neo.KillmailItemLoader []*github.com/ddouglas/neo.KillmailItem
//go:generate go run github.com/vektah/dataloaden KillmailVictimLoader uint64 *github.com/ddouglas/neo.KillmailVictim
//go:generate go run github.com/vektah/dataloaden TypeLoader uint64 *github.com/ddouglas/neo.Type
//go:generate go run github.com/vektah/dataloaden SolarSystemLoader uint64 *github.com/ddouglas/neo.SolarSystem

package generated
