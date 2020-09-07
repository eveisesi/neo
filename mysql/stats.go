package mysql

// func (r *statRepository) CreateStats(ctx context.Context, stats []*neo.Stat) error {

// 	places := make([]string, 0)
// 	values := make([]interface{}, 0)

// 	for _, v := range stats {
// 		places = append(places, "(?, ?, ?, ?, ?, ?, NOW(), NOW())")
// 		values = append(values, v.EntityID, v.EntityType, v.Category, v.Frequency, v.Date, v.Value)
// 	}

// 	query := `
// 		INSERT INTO stats (entity_id, entity_type, category, frequency, date, value, created_at, updated_at)
// 		VALUES %s ON DUPLICATE KEY UPDATE value = VALUES(value) + value
// 	`

// 	query = fmt.Sprintf(query, strings.Join(places, ","))

// 	_, err := r.db.ExecContext(ctx, query, values...)

// 	return err

// }

// func (r *statRepository) DeleteStat(ctx context.Context, entityID int64, entityType neo.StatEntity, date time.Time) error {

// 	_, err := boiler.Stats(
// 		boiler.StatWhere.EntityID.EQ(uint64(entityID)),
// 		boiler.StatWhere.EntityType.EQ(entityType.String()),
// 		boiler.StatWhere.Date.GTE(date),
// 	).DeleteAll(ctx, r.db)

// 	return err

// }
