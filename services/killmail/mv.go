package killmail

import (
	"context"

	"github.com/eveisesi/neo"
)

func (s *service) MVKAll(ctx context.Context, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.All(ctx, limit, age)
}

func (s *service) MVKByCharacterID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.LossesByCharacterID(ctx, id, limit, age)
}

func (s *service) MVLByCharacterID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.KillsByCharacterID(ctx, id, limit, age)
}

func (s *service) MVLByCorporationID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.LossesByCorporationID(ctx, id, limit, age)
}

func (s *service) MVKByCorporationID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.KillsByCorporationID(ctx, id, limit, age)
}

func (s *service) MVLByAllianceID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.LossesByAllianceID(ctx, id, limit, age)
}

func (s *service) MVKByAllianceID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.KillsByAllianceID(ctx, id, limit, age)
}

func (s *service) MVKByShipID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.KillsByShipID(ctx, id, age, limit)
}

func (s *service) MVLByShipID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.LossesByShipID(ctx, id, age, limit)
}
