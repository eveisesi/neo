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
			AllianceLoader:      dataloaders.AllianceLoader(ctx, s.alliance),
			CharacterLoader:     dataloaders.CharacterLoader(ctx, s.character),
			ConstellationLoader: dataloaders.ConstellationLoader(ctx, s.universe),
			CorporationLoader:   dataloaders.CorporationLoader(ctx, s.corporation),
			RegionLoader:        dataloaders.RegionLoader(ctx, s.universe),
			SolarSystemLoader:   dataloaders.SolarSystemLoader(ctx, s.universe),
			TypeLoader:          dataloaders.TypeLoader(ctx, s.universe),
			TypeAttributeLoader: dataloaders.TypeAttributeLoader(ctx, s.universe),
			TypeCategoryLoader:  dataloaders.TypeCategoryLoader(ctx, s.universe),
			TypeFlagLoader:      dataloaders.TypeFlagLoader(ctx, s.universe),
			TypeGroupLoader:     dataloaders.TypeGroupLoader(ctx, s.universe),
		}

		ctx = context.WithValue(ctx, ctxKey, loaders)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CtxLoaders(ctx context.Context) dataloaders.Loaders {
	return ctx.Value(ctxKey).(dataloaders.Loaders)
}
