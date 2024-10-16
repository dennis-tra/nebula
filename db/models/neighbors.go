// Code generated by SQLBoiler 4.14.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// Neighbor is an object representing the database table.
type Neighbor struct {
	CrawlID     int              `boil:"crawl_id" json:"crawl_id" toml:"crawl_id" yaml:"crawl_id"`
	PeerID      int              `boil:"peer_id" json:"peer_id" toml:"peer_id" yaml:"peer_id"`
	NeighborIds types.Int64Array `boil:"neighbor_ids" json:"neighbor_ids,omitempty" toml:"neighbor_ids" yaml:"neighbor_ids,omitempty"`
	ErrorBits   int16            `boil:"error_bits" json:"error_bits" toml:"error_bits" yaml:"error_bits"`

	R *neighborR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L neighborL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var NeighborColumns = struct {
	CrawlID     string
	PeerID      string
	NeighborIds string
	ErrorBits   string
}{
	CrawlID:     "crawl_id",
	PeerID:      "peer_id",
	NeighborIds: "neighbor_ids",
	ErrorBits:   "error_bits",
}

var NeighborTableColumns = struct {
	CrawlID     string
	PeerID      string
	NeighborIds string
	ErrorBits   string
}{
	CrawlID:     "neighbors.crawl_id",
	PeerID:      "neighbors.peer_id",
	NeighborIds: "neighbors.neighbor_ids",
	ErrorBits:   "neighbors.error_bits",
}

// Generated where

type whereHelpertypes_Int64Array struct{ field string }

func (w whereHelpertypes_Int64Array) EQ(x types.Int64Array) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpertypes_Int64Array) NEQ(x types.Int64Array) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpertypes_Int64Array) LT(x types.Int64Array) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_Int64Array) LTE(x types.Int64Array) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_Int64Array) GT(x types.Int64Array) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_Int64Array) GTE(x types.Int64Array) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpertypes_Int64Array) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpertypes_Int64Array) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

type whereHelperint16 struct{ field string }

func (w whereHelperint16) EQ(x int16) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint16) NEQ(x int16) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint16) LT(x int16) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint16) LTE(x int16) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint16) GT(x int16) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint16) GTE(x int16) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint16) IN(slice []int16) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint16) NIN(slice []int16) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

var NeighborWhere = struct {
	CrawlID     whereHelperint
	PeerID      whereHelperint
	NeighborIds whereHelpertypes_Int64Array
	ErrorBits   whereHelperint16
}{
	CrawlID:     whereHelperint{field: "\"neighbors\".\"crawl_id\""},
	PeerID:      whereHelperint{field: "\"neighbors\".\"peer_id\""},
	NeighborIds: whereHelpertypes_Int64Array{field: "\"neighbors\".\"neighbor_ids\""},
	ErrorBits:   whereHelperint16{field: "\"neighbors\".\"error_bits\""},
}

// NeighborRels is where relationship names are stored.
var NeighborRels = struct {
}{}

// neighborR is where relationships are stored.
type neighborR struct {
}

// NewStruct creates a new relationship struct
func (*neighborR) NewStruct() *neighborR {
	return &neighborR{}
}

// neighborL is where Load methods for each relationship are stored.
type neighborL struct{}

var (
	neighborAllColumns            = []string{"crawl_id", "peer_id", "neighbor_ids", "error_bits"}
	neighborColumnsWithoutDefault = []string{"crawl_id", "peer_id"}
	neighborColumnsWithDefault    = []string{"neighbor_ids", "error_bits"}
	neighborPrimaryKeyColumns     = []string{"crawl_id", "peer_id"}
	neighborGeneratedColumns      = []string{}
)

type (
	// NeighborSlice is an alias for a slice of pointers to Neighbor.
	// This should almost always be used instead of []Neighbor.
	NeighborSlice []*Neighbor
	// NeighborHook is the signature for custom Neighbor hook methods
	NeighborHook func(context.Context, boil.ContextExecutor, *Neighbor) error

	neighborQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	neighborType                 = reflect.TypeOf(&Neighbor{})
	neighborMapping              = queries.MakeStructMapping(neighborType)
	neighborPrimaryKeyMapping, _ = queries.BindMapping(neighborType, neighborMapping, neighborPrimaryKeyColumns)
	neighborInsertCacheMut       sync.RWMutex
	neighborInsertCache          = make(map[string]insertCache)
	neighborUpdateCacheMut       sync.RWMutex
	neighborUpdateCache          = make(map[string]updateCache)
	neighborUpsertCacheMut       sync.RWMutex
	neighborUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var neighborAfterSelectHooks []NeighborHook

var neighborBeforeInsertHooks []NeighborHook
var neighborAfterInsertHooks []NeighborHook

var neighborBeforeUpdateHooks []NeighborHook
var neighborAfterUpdateHooks []NeighborHook

var neighborBeforeDeleteHooks []NeighborHook
var neighborAfterDeleteHooks []NeighborHook

var neighborBeforeUpsertHooks []NeighborHook
var neighborAfterUpsertHooks []NeighborHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Neighbor) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Neighbor) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Neighbor) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Neighbor) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Neighbor) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Neighbor) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Neighbor) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Neighbor) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Neighbor) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range neighborAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddNeighborHook registers your hook function for all future operations.
func AddNeighborHook(hookPoint boil.HookPoint, neighborHook NeighborHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		neighborAfterSelectHooks = append(neighborAfterSelectHooks, neighborHook)
	case boil.BeforeInsertHook:
		neighborBeforeInsertHooks = append(neighborBeforeInsertHooks, neighborHook)
	case boil.AfterInsertHook:
		neighborAfterInsertHooks = append(neighborAfterInsertHooks, neighborHook)
	case boil.BeforeUpdateHook:
		neighborBeforeUpdateHooks = append(neighborBeforeUpdateHooks, neighborHook)
	case boil.AfterUpdateHook:
		neighborAfterUpdateHooks = append(neighborAfterUpdateHooks, neighborHook)
	case boil.BeforeDeleteHook:
		neighborBeforeDeleteHooks = append(neighborBeforeDeleteHooks, neighborHook)
	case boil.AfterDeleteHook:
		neighborAfterDeleteHooks = append(neighborAfterDeleteHooks, neighborHook)
	case boil.BeforeUpsertHook:
		neighborBeforeUpsertHooks = append(neighborBeforeUpsertHooks, neighborHook)
	case boil.AfterUpsertHook:
		neighborAfterUpsertHooks = append(neighborAfterUpsertHooks, neighborHook)
	}
}

// One returns a single neighbor record from the query.
func (q neighborQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Neighbor, error) {
	o := &Neighbor{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for neighbors")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Neighbor records from the query.
func (q neighborQuery) All(ctx context.Context, exec boil.ContextExecutor) (NeighborSlice, error) {
	var o []*Neighbor

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Neighbor slice")
	}

	if len(neighborAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Neighbor records in the query.
func (q neighborQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count neighbors rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q neighborQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if neighbors exists")
	}

	return count > 0, nil
}

// Neighbors retrieves all the records using an executor.
func Neighbors(mods ...qm.QueryMod) neighborQuery {
	mods = append(mods, qm.From("\"neighbors\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"neighbors\".*"})
	}

	return neighborQuery{q}
}

// FindNeighbor retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindNeighbor(ctx context.Context, exec boil.ContextExecutor, crawlID int, peerID int, selectCols ...string) (*Neighbor, error) {
	neighborObj := &Neighbor{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"neighbors\" where \"crawl_id\"=$1 AND \"peer_id\"=$2", sel,
	)

	q := queries.Raw(query, crawlID, peerID)

	err := q.Bind(ctx, exec, neighborObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from neighbors")
	}

	if err = neighborObj.doAfterSelectHooks(ctx, exec); err != nil {
		return neighborObj, err
	}

	return neighborObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Neighbor) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no neighbors provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(neighborColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	neighborInsertCacheMut.RLock()
	cache, cached := neighborInsertCache[key]
	neighborInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			neighborAllColumns,
			neighborColumnsWithDefault,
			neighborColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(neighborType, neighborMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(neighborType, neighborMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"neighbors\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"neighbors\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
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

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into neighbors")
	}

	if !cached {
		neighborInsertCacheMut.Lock()
		neighborInsertCache[key] = cache
		neighborInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Neighbor.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Neighbor) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	neighborUpdateCacheMut.RLock()
	cache, cached := neighborUpdateCache[key]
	neighborUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			neighborAllColumns,
			neighborPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update neighbors, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"neighbors\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, neighborPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(neighborType, neighborMapping, append(wl, neighborPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update neighbors row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for neighbors")
	}

	if !cached {
		neighborUpdateCacheMut.Lock()
		neighborUpdateCache[key] = cache
		neighborUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q neighborQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for neighbors")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for neighbors")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o NeighborSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), neighborPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"neighbors\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, neighborPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in neighbor slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all neighbor")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Neighbor) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no neighbors provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(neighborColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
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
	key := buf.String()
	strmangle.PutBuffer(buf)

	neighborUpsertCacheMut.RLock()
	cache, cached := neighborUpsertCache[key]
	neighborUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			neighborAllColumns,
			neighborColumnsWithDefault,
			neighborColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			neighborAllColumns,
			neighborPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert neighbors, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(neighborPrimaryKeyColumns))
			copy(conflict, neighborPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"neighbors\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(neighborType, neighborMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(neighborType, neighborMapping, ret)
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
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert neighbors")
	}

	if !cached {
		neighborUpsertCacheMut.Lock()
		neighborUpsertCache[key] = cache
		neighborUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Neighbor record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Neighbor) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Neighbor provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), neighborPrimaryKeyMapping)
	sql := "DELETE FROM \"neighbors\" WHERE \"crawl_id\"=$1 AND \"peer_id\"=$2"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from neighbors")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for neighbors")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q neighborQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no neighborQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from neighbors")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for neighbors")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o NeighborSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(neighborBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), neighborPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"neighbors\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, neighborPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from neighbor slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for neighbors")
	}

	if len(neighborAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Neighbor) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindNeighbor(ctx, exec, o.CrawlID, o.PeerID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *NeighborSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := NeighborSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), neighborPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"neighbors\".* FROM \"neighbors\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, neighborPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in NeighborSlice")
	}

	*o = slice

	return nil
}

// NeighborExists checks if the Neighbor row exists.
func NeighborExists(ctx context.Context, exec boil.ContextExecutor, crawlID int, peerID int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"neighbors\" where \"crawl_id\"=$1 AND \"peer_id\"=$2 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, crawlID, peerID)
	}
	row := exec.QueryRowContext(ctx, sql, crawlID, peerID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if neighbors exists")
	}

	return exists, nil
}

// Exists checks if the Neighbor row exists.
func (o *Neighbor) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return NeighborExists(ctx, exec, o.CrawlID, o.PeerID)
}
