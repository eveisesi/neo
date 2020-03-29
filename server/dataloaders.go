package server

import (
	"context"
	"net/http"

	"github.com/ddouglas/killboard/graphql/dataloaders"

	"github.com/ddouglas/killboard/graphql/dataloaders/generated"
)

type ctxKeyType struct{ string }

var ctxKey = ctxKeyType{"loaders"}

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

func (s *Server) Dataloaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		loaders := Loaders{
			AllianceLoader:          dataloaders.AllianceLoader(ctx, s.alliance),
			CharacterLoader:         dataloaders.CharacterLoader(ctx, s.character),
			CorporationLoader:       dataloaders.CorporationLoader(ctx, s.corporation),
			KillmailAttackersLoader: dataloaders.KillmailAttackersLoader(ctx, s.killmail),
			KillmailItemsLoader:     dataloaders.KillmailItemsLoader(ctx, s.killmail),
			KillmailVictimLoader:    dataloaders.KillmailVictimLoader(ctx, s.killmail),
			TypeLoader:              dataloaders.TypeLoader(ctx, s.universe),
			SolarSystemLoader:       dataloaders.SolarSystemLoader(ctx, s.universe),
		}

		ctx = context.WithValue(ctx, ctxKey, loaders)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
