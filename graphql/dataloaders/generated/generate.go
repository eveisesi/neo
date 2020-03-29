//go:generate go run github.com/vektah/dataloaden AllianceLoader uint64 *github.com/ddouglas/killboard.Alliance
//go:generate go run github.com/vektah/dataloaden CharacterLoader uint64 *github.com/ddouglas/killboard.Character
//go:generate go run github.com/vektah/dataloaden CorporationLoader uint64 *github.com/ddouglas/killboard.Corporation
//go:generate go run github.com/vektah/dataloaden KillmailAttackersLoader uint64 []*github.com/ddouglas/killboard.KillmailAttacker
//go:generate go run github.com/vektah/dataloaden KillmailItemsLoader *github.com/ddouglas/killboard.KillmailItemLoader []*github.com/ddouglas/killboard.KillmailItem
//go:generate go run github.com/vektah/dataloaden KillmailVictimLoader uint64 *github.com/ddouglas/killboard.KillmailVictim
//go:generate go run github.com/vektah/dataloaden TypeLoader uint64 *github.com/ddouglas/killboard.Type
//go:generate go run github.com/vektah/dataloaden SolarSystemLoader uint64 *github.com/ddouglas/killboard.SolarSystem

package generated
