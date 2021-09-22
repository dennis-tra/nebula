// Code generated by SQLBoiler 4.6.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// PegasysNeighbour is an object representing the database table.
type PegasysNeighbour struct {
	ID              int       `boil:"id" json:"id" toml:"id" yaml:"id"`
	PeerID          string    `boil:"peer_id" json:"peer_id" toml:"peer_id" yaml:"peer_id"`
	NeighbourPeerID string    `boil:"neighbour_peer_id" json:"neighbour_peer_id" toml:"neighbour_peer_id" yaml:"neighbour_peer_id"`
	CreatedAt       null.Time `boil:"created_at" json:"created_at,omitempty" toml:"created_at" yaml:"created_at,omitempty"`
	CrawlStartAt    null.Time `boil:"crawl_start_at" json:"crawl_start_at,omitempty" toml:"crawl_start_at" yaml:"crawl_start_at,omitempty"`

	R *pegasysNeighbourR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L pegasysNeighbourL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PegasysNeighbourColumns = struct {
	ID              string
	PeerID          string
	NeighbourPeerID string
	CreatedAt       string
	CrawlStartAt    string
}{
	ID:              "id",
	PeerID:          "peer_id",
	NeighbourPeerID: "neighbour_peer_id",
	CreatedAt:       "created_at",
	CrawlStartAt:    "crawl_start_at",
}

var PegasysNeighbourTableColumns = struct {
	ID              string
	PeerID          string
	NeighbourPeerID string
	CreatedAt       string
	CrawlStartAt    string
}{
	ID:              "pegasys_neighbours.id",
	PeerID:          "pegasys_neighbours.peer_id",
	NeighbourPeerID: "pegasys_neighbours.neighbour_peer_id",
	CreatedAt:       "pegasys_neighbours.created_at",
	CrawlStartAt:    "pegasys_neighbours.crawl_start_at",
}

// Generated where

var PegasysNeighbourWhere = struct {
	ID              whereHelperint
	PeerID          whereHelperstring
	NeighbourPeerID whereHelperstring
	CreatedAt       whereHelpernull_Time
	CrawlStartAt    whereHelpernull_Time
}{
	ID:              whereHelperint{field: "\"pegasys_neighbours\".\"id\""},
	PeerID:          whereHelperstring{field: "\"pegasys_neighbours\".\"peer_id\""},
	NeighbourPeerID: whereHelperstring{field: "\"pegasys_neighbours\".\"neighbour_peer_id\""},
	CreatedAt:       whereHelpernull_Time{field: "\"pegasys_neighbours\".\"created_at\""},
	CrawlStartAt:    whereHelpernull_Time{field: "\"pegasys_neighbours\".\"crawl_start_at\""},
}

// PegasysNeighbourRels is where relationship names are stored.
var PegasysNeighbourRels = struct {
}{}

// pegasysNeighbourR is where relationships are stored.
type pegasysNeighbourR struct {
}

// NewStruct creates a new relationship struct
func (*pegasysNeighbourR) NewStruct() *pegasysNeighbourR {
	return &pegasysNeighbourR{}
}

// pegasysNeighbourL is where Load methods for each relationship are stored.
type pegasysNeighbourL struct{}

var (
	pegasysNeighbourAllColumns            = []string{"id", "peer_id", "neighbour_peer_id", "created_at", "crawl_start_at"}
	pegasysNeighbourColumnsWithoutDefault = []string{"peer_id", "neighbour_peer_id", "created_at", "crawl_start_at"}
	pegasysNeighbourColumnsWithDefault    = []string{"id"}
	pegasysNeighbourPrimaryKeyColumns     = []string{"id"}
)

type (
	// PegasysNeighbourSlice is an alias for a slice of pointers to PegasysNeighbour.
	// This should almost always be used instead of []PegasysNeighbour.
	PegasysNeighbourSlice []*PegasysNeighbour
	// PegasysNeighbourHook is the signature for custom PegasysNeighbour hook methods
	PegasysNeighbourHook func(context.Context, boil.ContextExecutor, *PegasysNeighbour) error

	pegasysNeighbourQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	pegasysNeighbourType                 = reflect.TypeOf(&PegasysNeighbour{})
	pegasysNeighbourMapping              = queries.MakeStructMapping(pegasysNeighbourType)
	pegasysNeighbourPrimaryKeyMapping, _ = queries.BindMapping(pegasysNeighbourType, pegasysNeighbourMapping, pegasysNeighbourPrimaryKeyColumns)
	pegasysNeighbourInsertCacheMut       sync.RWMutex
	pegasysNeighbourInsertCache          = make(map[string]insertCache)
	pegasysNeighbourUpdateCacheMut       sync.RWMutex
	pegasysNeighbourUpdateCache          = make(map[string]updateCache)
	pegasysNeighbourUpsertCacheMut       sync.RWMutex
	pegasysNeighbourUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var pegasysNeighbourBeforeInsertHooks []PegasysNeighbourHook
var pegasysNeighbourBeforeUpdateHooks []PegasysNeighbourHook
var pegasysNeighbourBeforeDeleteHooks []PegasysNeighbourHook
var pegasysNeighbourBeforeUpsertHooks []PegasysNeighbourHook

var pegasysNeighbourAfterInsertHooks []PegasysNeighbourHook
var pegasysNeighbourAfterSelectHooks []PegasysNeighbourHook
var pegasysNeighbourAfterUpdateHooks []PegasysNeighbourHook
var pegasysNeighbourAfterDeleteHooks []PegasysNeighbourHook
var pegasysNeighbourAfterUpsertHooks []PegasysNeighbourHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *PegasysNeighbour) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *PegasysNeighbour) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *PegasysNeighbour) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *PegasysNeighbour) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *PegasysNeighbour) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *PegasysNeighbour) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *PegasysNeighbour) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *PegasysNeighbour) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *PegasysNeighbour) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range pegasysNeighbourAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPegasysNeighbourHook registers your hook function for all future operations.
func AddPegasysNeighbourHook(hookPoint boil.HookPoint, pegasysNeighbourHook PegasysNeighbourHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		pegasysNeighbourBeforeInsertHooks = append(pegasysNeighbourBeforeInsertHooks, pegasysNeighbourHook)
	case boil.BeforeUpdateHook:
		pegasysNeighbourBeforeUpdateHooks = append(pegasysNeighbourBeforeUpdateHooks, pegasysNeighbourHook)
	case boil.BeforeDeleteHook:
		pegasysNeighbourBeforeDeleteHooks = append(pegasysNeighbourBeforeDeleteHooks, pegasysNeighbourHook)
	case boil.BeforeUpsertHook:
		pegasysNeighbourBeforeUpsertHooks = append(pegasysNeighbourBeforeUpsertHooks, pegasysNeighbourHook)
	case boil.AfterInsertHook:
		pegasysNeighbourAfterInsertHooks = append(pegasysNeighbourAfterInsertHooks, pegasysNeighbourHook)
	case boil.AfterSelectHook:
		pegasysNeighbourAfterSelectHooks = append(pegasysNeighbourAfterSelectHooks, pegasysNeighbourHook)
	case boil.AfterUpdateHook:
		pegasysNeighbourAfterUpdateHooks = append(pegasysNeighbourAfterUpdateHooks, pegasysNeighbourHook)
	case boil.AfterDeleteHook:
		pegasysNeighbourAfterDeleteHooks = append(pegasysNeighbourAfterDeleteHooks, pegasysNeighbourHook)
	case boil.AfterUpsertHook:
		pegasysNeighbourAfterUpsertHooks = append(pegasysNeighbourAfterUpsertHooks, pegasysNeighbourHook)
	}
}

// One returns a single pegasysNeighbour record from the query.
func (q pegasysNeighbourQuery) One(ctx context.Context, exec boil.ContextExecutor) (*PegasysNeighbour, error) {
	o := &PegasysNeighbour{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for pegasys_neighbours")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all PegasysNeighbour records from the query.
func (q pegasysNeighbourQuery) All(ctx context.Context, exec boil.ContextExecutor) (PegasysNeighbourSlice, error) {
	var o []*PegasysNeighbour

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to PegasysNeighbour slice")
	}

	if len(pegasysNeighbourAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all PegasysNeighbour records in the query.
func (q pegasysNeighbourQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count pegasys_neighbours rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q pegasysNeighbourQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if pegasys_neighbours exists")
	}

	return count > 0, nil
}

// PegasysNeighbours retrieves all the records using an executor.
func PegasysNeighbours(mods ...qm.QueryMod) pegasysNeighbourQuery {
	mods = append(mods, qm.From("\"pegasys_neighbours\""))
	return pegasysNeighbourQuery{NewQuery(mods...)}
}

// FindPegasysNeighbour retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPegasysNeighbour(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*PegasysNeighbour, error) {
	pegasysNeighbourObj := &PegasysNeighbour{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"pegasys_neighbours\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, pegasysNeighbourObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from pegasys_neighbours")
	}

	if err = pegasysNeighbourObj.doAfterSelectHooks(ctx, exec); err != nil {
		return pegasysNeighbourObj, err
	}

	return pegasysNeighbourObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PegasysNeighbour) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no pegasys_neighbours provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if queries.MustTime(o.CreatedAt).IsZero() {
			queries.SetScanner(&o.CreatedAt, currTime)
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(pegasysNeighbourColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	pegasysNeighbourInsertCacheMut.RLock()
	cache, cached := pegasysNeighbourInsertCache[key]
	pegasysNeighbourInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			pegasysNeighbourAllColumns,
			pegasysNeighbourColumnsWithDefault,
			pegasysNeighbourColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(pegasysNeighbourType, pegasysNeighbourMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(pegasysNeighbourType, pegasysNeighbourMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"pegasys_neighbours\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"pegasys_neighbours\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into pegasys_neighbours")
	}

	if !cached {
		pegasysNeighbourInsertCacheMut.Lock()
		pegasysNeighbourInsertCache[key] = cache
		pegasysNeighbourInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the PegasysNeighbour.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PegasysNeighbour) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	pegasysNeighbourUpdateCacheMut.RLock()
	cache, cached := pegasysNeighbourUpdateCache[key]
	pegasysNeighbourUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			pegasysNeighbourAllColumns,
			pegasysNeighbourPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update pegasys_neighbours, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"pegasys_neighbours\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, pegasysNeighbourPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(pegasysNeighbourType, pegasysNeighbourMapping, append(wl, pegasysNeighbourPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update pegasys_neighbours row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for pegasys_neighbours")
	}

	if !cached {
		pegasysNeighbourUpdateCacheMut.Lock()
		pegasysNeighbourUpdateCache[key] = cache
		pegasysNeighbourUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q pegasysNeighbourQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for pegasys_neighbours")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for pegasys_neighbours")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PegasysNeighbourSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), pegasysNeighbourPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"pegasys_neighbours\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, pegasysNeighbourPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in pegasysNeighbour slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all pegasysNeighbour")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PegasysNeighbour) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no pegasys_neighbours provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if queries.MustTime(o.CreatedAt).IsZero() {
			queries.SetScanner(&o.CreatedAt, currTime)
		}
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(pegasysNeighbourColumnsWithDefault, o)

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

	pegasysNeighbourUpsertCacheMut.RLock()
	cache, cached := pegasysNeighbourUpsertCache[key]
	pegasysNeighbourUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			pegasysNeighbourAllColumns,
			pegasysNeighbourColumnsWithDefault,
			pegasysNeighbourColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			pegasysNeighbourAllColumns,
			pegasysNeighbourPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert pegasys_neighbours, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(pegasysNeighbourPrimaryKeyColumns))
			copy(conflict, pegasysNeighbourPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"pegasys_neighbours\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(pegasysNeighbourType, pegasysNeighbourMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(pegasysNeighbourType, pegasysNeighbourMapping, ret)
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
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert pegasys_neighbours")
	}

	if !cached {
		pegasysNeighbourUpsertCacheMut.Lock()
		pegasysNeighbourUpsertCache[key] = cache
		pegasysNeighbourUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single PegasysNeighbour record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PegasysNeighbour) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no PegasysNeighbour provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), pegasysNeighbourPrimaryKeyMapping)
	sql := "DELETE FROM \"pegasys_neighbours\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from pegasys_neighbours")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for pegasys_neighbours")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q pegasysNeighbourQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no pegasysNeighbourQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from pegasys_neighbours")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for pegasys_neighbours")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PegasysNeighbourSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(pegasysNeighbourBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), pegasysNeighbourPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"pegasys_neighbours\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, pegasysNeighbourPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from pegasysNeighbour slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for pegasys_neighbours")
	}

	if len(pegasysNeighbourAfterDeleteHooks) != 0 {
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
func (o *PegasysNeighbour) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindPegasysNeighbour(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PegasysNeighbourSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PegasysNeighbourSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), pegasysNeighbourPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"pegasys_neighbours\".* FROM \"pegasys_neighbours\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, pegasysNeighbourPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in PegasysNeighbourSlice")
	}

	*o = slice

	return nil
}

// PegasysNeighbourExists checks if the PegasysNeighbour row exists.
func PegasysNeighbourExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"pegasys_neighbours\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if pegasys_neighbours exists")
	}

	return exists, nil
}