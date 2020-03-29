package server

import (
	"context"
	"net/http"

	"github.com/eveisesi/neo/graphql/dataloaders"
)

type ctxKeyType struct{ string }

var ctxKey = ctxKeyType{"loaders"}

func (s *Server) Dataloaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		loaders := dataloaders.Loaders{
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

func CtxLoaders(ctx context.Context) dataloaders.Loaders {
	return ctx.Value(ctxKey).(dataloaders.Loaders)
}
