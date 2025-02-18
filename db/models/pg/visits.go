// Code generated by SQLBoiler 4.14.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package pg

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
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// Visit is an object representing the database table.
type Visit struct {
	ID              int              `boil:"id" json:"id" toml:"id" yaml:"id"`
	PeerID          int              `boil:"peer_id" json:"peer_id" toml:"peer_id" yaml:"peer_id"`
	CrawlID         null.Int         `boil:"crawl_id" json:"crawl_id,omitempty" toml:"crawl_id" yaml:"crawl_id,omitempty"`
	SessionID       null.Int         `boil:"session_id" json:"session_id,omitempty" toml:"session_id" yaml:"session_id,omitempty"`
	AgentVersionID  null.Int         `boil:"agent_version_id" json:"agent_version_id,omitempty" toml:"agent_version_id" yaml:"agent_version_id,omitempty"`
	ProtocolsSetID  null.Int         `boil:"protocols_set_id" json:"protocols_set_id,omitempty" toml:"protocols_set_id" yaml:"protocols_set_id,omitempty"`
	Type            string           `boil:"type" json:"type" toml:"type" yaml:"type"`
	ConnectError    null.String      `boil:"connect_error" json:"connect_error,omitempty" toml:"connect_error" yaml:"connect_error,omitempty"`
	CrawlError      null.String      `boil:"crawl_error" json:"crawl_error,omitempty" toml:"crawl_error" yaml:"crawl_error,omitempty"`
	VisitStartedAt  time.Time        `boil:"visit_started_at" json:"visit_started_at" toml:"visit_started_at" yaml:"visit_started_at"`
	VisitEndedAt    time.Time        `boil:"visit_ended_at" json:"visit_ended_at" toml:"visit_ended_at" yaml:"visit_ended_at"`
	CreatedAt       time.Time        `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	DialDuration    null.String      `boil:"dial_duration" json:"dial_duration,omitempty" toml:"dial_duration" yaml:"dial_duration,omitempty"`
	ConnectDuration null.String      `boil:"connect_duration" json:"connect_duration,omitempty" toml:"connect_duration" yaml:"connect_duration,omitempty"`
	CrawlDuration   null.String      `boil:"crawl_duration" json:"crawl_duration,omitempty" toml:"crawl_duration" yaml:"crawl_duration,omitempty"`
	MultiAddressIds types.Int64Array `boil:"multi_address_ids" json:"multi_address_ids,omitempty" toml:"multi_address_ids" yaml:"multi_address_ids,omitempty"`
	PeerProperties  null.JSON        `boil:"peer_properties" json:"peer_properties,omitempty" toml:"peer_properties" yaml:"peer_properties,omitempty"`

	R *visitR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L visitL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var VisitColumns = struct {
	ID              string
	PeerID          string
	CrawlID         string
	SessionID       string
	AgentVersionID  string
	ProtocolsSetID  string
	Type            string
	ConnectError    string
	CrawlError      string
	VisitStartedAt  string
	VisitEndedAt    string
	CreatedAt       string
	DialDuration    string
	ConnectDuration string
	CrawlDuration   string
	MultiAddressIds string
	PeerProperties  string
}{
	ID:              "id",
	PeerID:          "peer_id",
	CrawlID:         "crawl_id",
	SessionID:       "session_id",
	AgentVersionID:  "agent_version_id",
	ProtocolsSetID:  "protocols_set_id",
	Type:            "type",
	ConnectError:    "connect_error",
	CrawlError:      "crawl_error",
	VisitStartedAt:  "visit_started_at",
	VisitEndedAt:    "visit_ended_at",
	CreatedAt:       "created_at",
	DialDuration:    "dial_duration",
	ConnectDuration: "connect_duration",
	CrawlDuration:   "crawl_duration",
	MultiAddressIds: "multi_address_ids",
	PeerProperties:  "peer_properties",
}

var VisitTableColumns = struct {
	ID              string
	PeerID          string
	CrawlID         string
	SessionID       string
	AgentVersionID  string
	ProtocolsSetID  string
	Type            string
	ConnectError    string
	CrawlError      string
	VisitStartedAt  string
	VisitEndedAt    string
	CreatedAt       string
	DialDuration    string
	ConnectDuration string
	CrawlDuration   string
	MultiAddressIds string
	PeerProperties  string
}{
	ID:              "visits.id",
	PeerID:          "visits.peer_id",
	CrawlID:         "visits.crawl_id",
	SessionID:       "visits.session_id",
	AgentVersionID:  "visits.agent_version_id",
	ProtocolsSetID:  "visits.protocols_set_id",
	Type:            "visits.type",
	ConnectError:    "visits.connect_error",
	CrawlError:      "visits.crawl_error",
	VisitStartedAt:  "visits.visit_started_at",
	VisitEndedAt:    "visits.visit_ended_at",
	CreatedAt:       "visits.created_at",
	DialDuration:    "visits.dial_duration",
	ConnectDuration: "visits.connect_duration",
	CrawlDuration:   "visits.crawl_duration",
	MultiAddressIds: "visits.multi_address_ids",
	PeerProperties:  "visits.peer_properties",
}

// Generated where

var VisitWhere = struct {
	ID              whereHelperint
	PeerID          whereHelperint
	CrawlID         whereHelpernull_Int
	SessionID       whereHelpernull_Int
	AgentVersionID  whereHelpernull_Int
	ProtocolsSetID  whereHelpernull_Int
	Type            whereHelperstring
	ConnectError    whereHelpernull_String
	CrawlError      whereHelpernull_String
	VisitStartedAt  whereHelpertime_Time
	VisitEndedAt    whereHelpertime_Time
	CreatedAt       whereHelpertime_Time
	DialDuration    whereHelpernull_String
	ConnectDuration whereHelpernull_String
	CrawlDuration   whereHelpernull_String
	MultiAddressIds whereHelpertypes_Int64Array
	PeerProperties  whereHelpernull_JSON
}{
	ID:              whereHelperint{field: "\"visits\".\"id\""},
	PeerID:          whereHelperint{field: "\"visits\".\"peer_id\""},
	CrawlID:         whereHelpernull_Int{field: "\"visits\".\"crawl_id\""},
	SessionID:       whereHelpernull_Int{field: "\"visits\".\"session_id\""},
	AgentVersionID:  whereHelpernull_Int{field: "\"visits\".\"agent_version_id\""},
	ProtocolsSetID:  whereHelpernull_Int{field: "\"visits\".\"protocols_set_id\""},
	Type:            whereHelperstring{field: "\"visits\".\"type\""},
	ConnectError:    whereHelpernull_String{field: "\"visits\".\"connect_error\""},
	CrawlError:      whereHelpernull_String{field: "\"visits\".\"crawl_error\""},
	VisitStartedAt:  whereHelpertime_Time{field: "\"visits\".\"visit_started_at\""},
	VisitEndedAt:    whereHelpertime_Time{field: "\"visits\".\"visit_ended_at\""},
	CreatedAt:       whereHelpertime_Time{field: "\"visits\".\"created_at\""},
	DialDuration:    whereHelpernull_String{field: "\"visits\".\"dial_duration\""},
	ConnectDuration: whereHelpernull_String{field: "\"visits\".\"connect_duration\""},
	CrawlDuration:   whereHelpernull_String{field: "\"visits\".\"crawl_duration\""},
	MultiAddressIds: whereHelpertypes_Int64Array{field: "\"visits\".\"multi_address_ids\""},
	PeerProperties:  whereHelpernull_JSON{field: "\"visits\".\"peer_properties\""},
}

// VisitRels is where relationship names are stored.
var VisitRels = struct {
}{}

// visitR is where relationships are stored.
type visitR struct {
}

// NewStruct creates a new relationship struct
func (*visitR) NewStruct() *visitR {
	return &visitR{}
}

// visitL is where Load methods for each relationship are stored.
type visitL struct{}

var (
	visitAllColumns            = []string{"id", "peer_id", "crawl_id", "session_id", "agent_version_id", "protocols_set_id", "type", "connect_error", "crawl_error", "visit_started_at", "visit_ended_at", "created_at", "dial_duration", "connect_duration", "crawl_duration", "multi_address_ids", "peer_properties"}
	visitColumnsWithoutDefault = []string{"peer_id", "type", "visit_started_at", "visit_ended_at", "created_at"}
	visitColumnsWithDefault    = []string{"id", "crawl_id", "session_id", "agent_version_id", "protocols_set_id", "connect_error", "crawl_error", "dial_duration", "connect_duration", "crawl_duration", "multi_address_ids", "peer_properties"}
	visitPrimaryKeyColumns     = []string{"id", "visit_started_at"}
	visitGeneratedColumns      = []string{"id"}
)

type (
	// VisitSlice is an alias for a slice of pointers to Visit.
	// This should almost always be used instead of []Visit.
	VisitSlice []*Visit
	// VisitHook is the signature for custom Visit hook methods
	VisitHook func(context.Context, boil.ContextExecutor, *Visit) error

	visitQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	visitType                 = reflect.TypeOf(&Visit{})
	visitMapping              = queries.MakeStructMapping(visitType)
	visitPrimaryKeyMapping, _ = queries.BindMapping(visitType, visitMapping, visitPrimaryKeyColumns)
	visitInsertCacheMut       sync.RWMutex
	visitInsertCache          = make(map[string]insertCache)
	visitUpdateCacheMut       sync.RWMutex
	visitUpdateCache          = make(map[string]updateCache)
	visitUpsertCacheMut       sync.RWMutex
	visitUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var visitAfterSelectHooks []VisitHook

var visitBeforeInsertHooks []VisitHook
var visitAfterInsertHooks []VisitHook

var visitBeforeUpdateHooks []VisitHook
var visitAfterUpdateHooks []VisitHook

var visitBeforeDeleteHooks []VisitHook
var visitAfterDeleteHooks []VisitHook

var visitBeforeUpsertHooks []VisitHook
var visitAfterUpsertHooks []VisitHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Visit) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Visit) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Visit) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Visit) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Visit) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Visit) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Visit) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Visit) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Visit) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range visitAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddVisitHook registers your hook function for all future operations.
func AddVisitHook(hookPoint boil.HookPoint, visitHook VisitHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		visitAfterSelectHooks = append(visitAfterSelectHooks, visitHook)
	case boil.BeforeInsertHook:
		visitBeforeInsertHooks = append(visitBeforeInsertHooks, visitHook)
	case boil.AfterInsertHook:
		visitAfterInsertHooks = append(visitAfterInsertHooks, visitHook)
	case boil.BeforeUpdateHook:
		visitBeforeUpdateHooks = append(visitBeforeUpdateHooks, visitHook)
	case boil.AfterUpdateHook:
		visitAfterUpdateHooks = append(visitAfterUpdateHooks, visitHook)
	case boil.BeforeDeleteHook:
		visitBeforeDeleteHooks = append(visitBeforeDeleteHooks, visitHook)
	case boil.AfterDeleteHook:
		visitAfterDeleteHooks = append(visitAfterDeleteHooks, visitHook)
	case boil.BeforeUpsertHook:
		visitBeforeUpsertHooks = append(visitBeforeUpsertHooks, visitHook)
	case boil.AfterUpsertHook:
		visitAfterUpsertHooks = append(visitAfterUpsertHooks, visitHook)
	}
}

// One returns a single visit record from the query.
func (q visitQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Visit, error) {
	o := &Visit{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "pg: failed to execute a one query for visits")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Visit records from the query.
func (q visitQuery) All(ctx context.Context, exec boil.ContextExecutor) (VisitSlice, error) {
	var o []*Visit

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "pg: failed to assign all query results to Visit slice")
	}

	if len(visitAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Visit records in the query.
func (q visitQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to count visits rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q visitQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "pg: failed to check if visits exists")
	}

	return count > 0, nil
}

// Visits retrieves all the records using an executor.
func Visits(mods ...qm.QueryMod) visitQuery {
	mods = append(mods, qm.From("\"visits\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"visits\".*"})
	}

	return visitQuery{q}
}

// FindVisit retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindVisit(ctx context.Context, exec boil.ContextExecutor, iD int, visitStartedAt time.Time, selectCols ...string) (*Visit, error) {
	visitObj := &Visit{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"visits\" where \"id\"=$1 AND \"visit_started_at\"=$2", sel,
	)

	q := queries.Raw(query, iD, visitStartedAt)

	err := q.Bind(ctx, exec, visitObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "pg: unable to select from visits")
	}

	if err = visitObj.doAfterSelectHooks(ctx, exec); err != nil {
		return visitObj, err
	}

	return visitObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Visit) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("pg: no visits provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(visitColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	visitInsertCacheMut.RLock()
	cache, cached := visitInsertCache[key]
	visitInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			visitAllColumns,
			visitColumnsWithDefault,
			visitColumnsWithoutDefault,
			nzDefaults,
		)
		wl = strmangle.SetComplement(wl, visitGeneratedColumns)

		cache.valueMapping, err = queries.BindMapping(visitType, visitMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(visitType, visitMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"visits\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"visits\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "pg: unable to insert into visits")
	}

	if !cached {
		visitInsertCacheMut.Lock()
		visitInsertCache[key] = cache
		visitInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Visit.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Visit) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	visitUpdateCacheMut.RLock()
	cache, cached := visitUpdateCache[key]
	visitUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			visitAllColumns,
			visitPrimaryKeyColumns,
		)
		wl = strmangle.SetComplement(wl, visitGeneratedColumns)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("pg: unable to update visits, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"visits\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, visitPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(visitType, visitMapping, append(wl, visitPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "pg: unable to update visits row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to get rows affected by update for visits")
	}

	if !cached {
		visitUpdateCacheMut.Lock()
		visitUpdateCache[key] = cache
		visitUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q visitQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to update all for visits")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to retrieve rows affected for visits")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o VisitSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("pg: update all requires at least one column argument")
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), visitPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"visits\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, visitPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to update all in visit slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to retrieve rows affected all in update all visit")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Visit) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("pg: no visits provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(visitColumnsWithDefault, o)

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

	visitUpsertCacheMut.RLock()
	cache, cached := visitUpsertCache[key]
	visitUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			visitAllColumns,
			visitColumnsWithDefault,
			visitColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			visitAllColumns,
			visitPrimaryKeyColumns,
		)

		insert = strmangle.SetComplement(insert, visitGeneratedColumns)
		update = strmangle.SetComplement(update, visitGeneratedColumns)

		if updateOnConflict && len(update) == 0 {
			return errors.New("pg: unable to upsert visits, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(visitPrimaryKeyColumns))
			copy(conflict, visitPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"visits\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(visitType, visitMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(visitType, visitMapping, ret)
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
		return errors.Wrap(err, "pg: unable to upsert visits")
	}

	if !cached {
		visitUpsertCacheMut.Lock()
		visitUpsertCache[key] = cache
		visitUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Visit record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Visit) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("pg: no Visit provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), visitPrimaryKeyMapping)
	sql := "DELETE FROM \"visits\" WHERE \"id\"=$1 AND \"visit_started_at\"=$2"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to delete from visits")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to get rows affected by delete for visits")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q visitQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("pg: no visitQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to delete all from visits")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to get rows affected by deleteall for visits")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o VisitSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(visitBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), visitPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"visits\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, visitPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to delete all from visit slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to get rows affected by deleteall for visits")
	}

	if len(visitAfterDeleteHooks) != 0 {
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
func (o *Visit) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindVisit(ctx, exec, o.ID, o.VisitStartedAt)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *VisitSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := VisitSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), visitPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"visits\".* FROM \"visits\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, visitPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "pg: unable to reload all in VisitSlice")
	}

	*o = slice

	return nil
}

// VisitExists checks if the Visit row exists.
func VisitExists(ctx context.Context, exec boil.ContextExecutor, iD int, visitStartedAt time.Time) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"visits\" where \"id\"=$1 AND \"visit_started_at\"=$2 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD, visitStartedAt)
	}
	row := exec.QueryRowContext(ctx, sql, iD, visitStartedAt)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "pg: unable to check if visits exists")
	}

	return exists, nil
}

// Exists checks if the Visit row exists.
func (o *Visit) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return VisitExists(ctx, exec, o.ID, o.VisitStartedAt)
}
