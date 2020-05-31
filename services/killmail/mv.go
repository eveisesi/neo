package killmail

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/tools"
	"github.com/pkg/errors"
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

func (s *service) MVKByShipGroupID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	allowed := tools.IsGroupAllowed(id)
	if !allowed {
		return nil, errors.New("invalid group id. Only published group ids are allowed")
	}

	return s.mvks.KillsByShipGroupID(ctx, id, age, limit)
}

func (s *service) MVLByShipGroupID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	allowed := tools.IsGroupAllowed(id)
	if !allowed {
		return nil, errors.New("invalid group id. Only published group ids are allowed")
	}

	return s.mvks.LossesByShipGroupID(ctx, id, age, limit)
}

func (s *service) MVKBySystemID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.KillsBySystemID(ctx, id, age, limit)
}

func (s *service) MVKByConstellationID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.KillsByConstellationID(ctx, id, age, limit)
}

func (s *service) MVKByRegionID(ctx context.Context, id uint64, age, limit int) ([]*neo.Killmail, error) {
	return s.mvks.KillsByRegionID(ctx, id, age, limit)
}
