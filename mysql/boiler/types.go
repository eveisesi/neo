// Code generated by SQLBoiler 3.6.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package boiler

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/strmangle"
)

// Type is an object representing the database table.
type Type struct {
	ID            uint64       `boil:"id" json:"id" toml:"id" yaml:"id"`
	GroupID       uint64       `boil:"group_id" json:"group_id" toml:"group_id" yaml:"group_id"`
	Name          string       `boil:"name" json:"name" toml:"name" yaml:"name"`
	Description   string       `boil:"description" json:"description" toml:"description" yaml:"description"`
	Volume        float64      `boil:"volume" json:"volume" toml:"volume" yaml:"volume"`
	RaceID        null.Uint64  `boil:"race_id" json:"race_id,omitempty" toml:"race_id" yaml:"race_id,omitempty"`
	BasePrice     null.Float64 `boil:"base_price" json:"base_price,omitempty" toml:"base_price" yaml:"base_price,omitempty"`
	Published     bool         `boil:"published" json:"published" toml:"published" yaml:"published"`
	MarketGroupID null.Uint64  `boil:"market_group_id" json:"market_group_id,omitempty" toml:"market_group_id" yaml:"market_group_id,omitempty"`
	CreatedAt     null.Time    `boil:"created_at" json:"created_at,omitempty" toml:"created_at" yaml:"created_at,omitempty"`
	UpdatedAt     null.Time    `boil:"updated_at" json:"updated_at,omitempty" toml:"updated_at" yaml:"updated_at,omitempty"`

	R *typeR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L typeL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var TypeColumns = struct {
	ID            string
	GroupID       string
	Name          string
	Description   string
	Volume        string
	RaceID        string
	BasePrice     string
	Published     string
	MarketGroupID string
	CreatedAt     string
	UpdatedAt     string
}{
	ID:            "id",
	GroupID:       "group_id",
	Name:          "name",
	Description:   "description",
	Volume:        "volume",
	RaceID:        "race_id",
	BasePrice:     "base_price",
	Published:     "published",
	MarketGroupID: "market_group_id",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// Generated where

type whereHelpernull_Time struct{ field string }

func (w whereHelpernull_Time) EQ(x null.Time) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Time) NEQ(x null.Time) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Time) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Time) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }
func (w whereHelpernull_Time) LT(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Time) LTE(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Time) GT(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Time) GTE(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var TypeWhere = struct {
	ID            whereHelperuint64
	GroupID       whereHelperuint64
	Name          whereHelperstring
	Description   whereHelperstring
	Volume        whereHelperfloat64
	RaceID        whereHelpernull_Uint64
	BasePrice     whereHelpernull_Float64
	Published     whereHelperbool
	MarketGroupID whereHelpernull_Uint64
	CreatedAt     whereHelpernull_Time
	UpdatedAt     whereHelpernull_Time
}{
	ID:            whereHelperuint64{field: "`types`.`id`"},
	GroupID:       whereHelperuint64{field: "`types`.`group_id`"},
	Name:          whereHelperstring{field: "`types`.`name`"},
	Description:   whereHelperstring{field: "`types`.`description`"},
	Volume:        whereHelperfloat64{field: "`types`.`volume`"},
	RaceID:        whereHelpernull_Uint64{field: "`types`.`race_id`"},
	BasePrice:     whereHelpernull_Float64{field: "`types`.`base_price`"},
	Published:     whereHelperbool{field: "`types`.`published`"},
	MarketGroupID: whereHelpernull_Uint64{field: "`types`.`market_group_id`"},
	CreatedAt:     whereHelpernull_Time{field: "`types`.`created_at`"},
	UpdatedAt:     whereHelpernull_Time{field: "`types`.`updated_at`"},
}

// TypeRels is where relationship names are stored.
var TypeRels = struct {
}{}

// typeR is where relationships are stored.
type typeR struct {
}

// NewStruct creates a new relationship struct
func (*typeR) NewStruct() *typeR {
	return &typeR{}
}

// typeL is where Load methods for each relationship are stored.
type typeL struct{}

var (
	typeAllColumns            = []string{"id", "group_id", "name", "description", "volume", "race_id", "base_price", "published", "market_group_id", "created_at", "updated_at"}
	typeColumnsWithoutDefault = []string{"id", "group_id", "name", "description", "volume", "race_id", "base_price", "published", "market_group_id"}
	typeColumnsWithDefault    = []string{"created_at", "updated_at"}
	typePrimaryKeyColumns     = []string{"id"}
)

type (
	// TypeSlice is an alias for a slice of pointers to Type.
	// This should generally be used opposed to []Type.
	TypeSlice []*Type

	typeQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	typeType                 = reflect.TypeOf(&Type{})
	typeMapping              = queries.MakeStructMapping(typeType)
	typePrimaryKeyMapping, _ = queries.BindMapping(typeType, typeMapping, typePrimaryKeyColumns)
	typeInsertCacheMut       sync.RWMutex
	typeInsertCache          = make(map[string]insertCache)
	typeUpdateCacheMut       sync.RWMutex
	typeUpdateCache          = make(map[string]updateCache)
	typeUpsertCacheMut       sync.RWMutex
	typeUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single type record from the query.
func (q typeQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Type, error) {
	o := &Type{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for types")
	}

	return o, nil
}

// All returns all Type records from the query.
func (q typeQuery) All(ctx context.Context, exec boil.ContextExecutor) (TypeSlice, error) {
	var o []*Type

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Type slice")
	}

	return o, nil
}

// Count returns the count of all Type records in the query.
func (q typeQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count types rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q typeQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if types exists")
	}

	return count > 0, nil
}

// Types retrieves all the records using an executor.
func Types(mods ...qm.QueryMod) typeQuery {
	mods = append(mods, qm.From("`types`"))
	return typeQuery{NewQuery(mods...)}
}

// FindType retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindType(ctx context.Context, exec boil.ContextExecutor, iD uint64, selectCols ...string) (*Type, error) {
	typeObj := &Type{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `types` where `id`=?", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, typeObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from types")
	}

	return typeObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Type) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no types provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if queries.MustTime(o.CreatedAt).IsZero() {
			queries.SetScanner(&o.CreatedAt, currTime)
		}
		if queries.MustTime(o.UpdatedAt).IsZero() {
			queries.SetScanner(&o.UpdatedAt, currTime)
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(typeColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	typeInsertCacheMut.RLock()
	cache, cached := typeInsertCache[key]
	typeInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			typeAllColumns,
			typeColumnsWithDefault,
			typeColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(typeType, typeMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(typeType, typeMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `types` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `types` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `types` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, typePrimaryKeyColumns))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	_, err = exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "boiler: unable to insert into types")
	}

	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.ID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to populate default values for types")
	}

CacheNoHooks:
	if !cached {
		typeInsertCacheMut.Lock()
		typeInsertCache[key] = cache
		typeInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the Type.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Type) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		queries.SetScanner(&o.UpdatedAt, currTime)
	}

	var err error
	key := makeCacheKey(columns, nil)
	typeUpdateCacheMut.RLock()
	cache, cached := typeUpdateCache[key]
	typeUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			typeAllColumns,
			typePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update types, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `types` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, typePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(typeType, typeMapping, append(wl, typePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update types row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for types")
	}

	if !cached {
		typeUpdateCacheMut.Lock()
		typeUpdateCache[key] = cache
		typeUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q typeQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for types")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for types")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o TypeSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("boiler: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), typePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `types` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, typePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in type slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all type")
	}
	return rowsAff, nil
}

var mySQLTypeUniqueColumns = []string{
	"id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Type) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no types provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if queries.MustTime(o.CreatedAt).IsZero() {
			queries.SetScanner(&o.CreatedAt, currTime)
		}
		queries.SetScanner(&o.UpdatedAt, currTime)
	}

	nzDefaults := queries.NonZeroDefaultSet(typeColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLTypeUniqueColumns, o)

	if len(nzUniques) == 0 {
		return errors.New("cannot upsert with a table that cannot conflict on a unique column")
	}

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzUniques {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	typeUpsertCacheMut.RLock()
	cache, cached := typeUpsertCache[key]
	typeUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			typeAllColumns,
			typeColumnsWithDefault,
			typeColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			typeAllColumns,
			typePrimaryKeyColumns,
		)

		if len(update) == 0 {
			return errors.New("boiler: unable to upsert types, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "types", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `types` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(typeType, typeMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(typeType, typeMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	_, err = exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "boiler: unable to upsert for types")
	}

	var uniqueMap []uint64
	var nzUniqueCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(typeType, typeMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to retrieve unique values for types")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to populate default values for types")
	}

CacheNoHooks:
	if !cached {
		typeUpsertCacheMut.Lock()
		typeUpsertCache[key] = cache
		typeUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single Type record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Type) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Type provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), typePrimaryKeyMapping)
	sql := "DELETE FROM `types` WHERE `id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from types")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for types")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q typeQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no typeQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from types")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for types")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o TypeSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), typePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `types` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, typePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from type slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for types")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Type) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindType(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *TypeSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := TypeSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), typePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `types`.* FROM `types` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, typePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in TypeSlice")
	}

	*o = slice

	return nil
}

// TypeExists checks if the Type row exists.
func TypeExists(ctx context.Context, exec boil.ContextExecutor, iD uint64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `types` where `id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if types exists")
	}

	return exists, nil
}
