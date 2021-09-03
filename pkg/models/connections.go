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

// Connection is an object representing the database table.
type Connection struct {
	ID           int         `boil:"id" json:"id" toml:"id" yaml:"id"`
	PeerID       string      `boil:"peer_id" json:"peer_id" toml:"peer_id" yaml:"peer_id"`
	MultiAddress null.String `boil:"multi_address" json:"multi_address,omitempty" toml:"multi_address" yaml:"multi_address,omitempty"`
	AgentVersion null.String `boil:"agent_version" json:"agent_version,omitempty" toml:"agent_version" yaml:"agent_version,omitempty"`
	DialAttempt  null.Time   `boil:"dial_attempt" json:"dial_attempt,omitempty" toml:"dial_attempt" yaml:"dial_attempt,omitempty"`
	Latency      null.String `boil:"latency" json:"latency,omitempty" toml:"latency" yaml:"latency,omitempty"`
	IsSucceed    null.Bool   `boil:"is_succeed" json:"is_succeed,omitempty" toml:"is_succeed" yaml:"is_succeed,omitempty"`

	R *connectionR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L connectionL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ConnectionColumns = struct {
	ID           string
	PeerID       string
	MultiAddress string
	AgentVersion string
	DialAttempt  string
	Latency      string
	IsSucceed    string
}{
	ID:           "id",
	PeerID:       "peer_id",
	MultiAddress: "multi_address",
	AgentVersion: "agent_version",
	DialAttempt:  "dial_attempt",
	Latency:      "latency",
	IsSucceed:    "is_succeed",
}

var ConnectionTableColumns = struct {
	ID           string
	PeerID       string
	MultiAddress string
	AgentVersion string
	DialAttempt  string
	Latency      string
	IsSucceed    string
}{
	ID:           "connections.id",
	PeerID:       "connections.peer_id",
	MultiAddress: "connections.multi_address",
	AgentVersion: "connections.agent_version",
	DialAttempt:  "connections.dial_attempt",
	Latency:      "connections.latency",
	IsSucceed:    "connections.is_succeed",
}

// Generated where

type whereHelperint struct{ field string }

func (w whereHelperint) EQ(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint) NEQ(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint) LT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint) LTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint) GT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint) GTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint) IN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint) NIN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperstring) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperstring) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpernull_String struct{ field string }

func (w whereHelpernull_String) EQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_String) NEQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_String) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_String) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }
func (w whereHelpernull_String) LT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_String) LTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_String) GT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_String) GTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

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

type whereHelpernull_Bool struct{ field string }

func (w whereHelpernull_Bool) EQ(x null.Bool) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Bool) NEQ(x null.Bool) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Bool) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Bool) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }
func (w whereHelpernull_Bool) LT(x null.Bool) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Bool) LTE(x null.Bool) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Bool) GT(x null.Bool) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Bool) GTE(x null.Bool) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var ConnectionWhere = struct {
	ID           whereHelperint
	PeerID       whereHelperstring
	MultiAddress whereHelpernull_String
	AgentVersion whereHelpernull_String
	DialAttempt  whereHelpernull_Time
	Latency      whereHelpernull_String
	IsSucceed    whereHelpernull_Bool
}{
	ID:           whereHelperint{field: "\"connections\".\"id\""},
	PeerID:       whereHelperstring{field: "\"connections\".\"peer_id\""},
	MultiAddress: whereHelpernull_String{field: "\"connections\".\"multi_address\""},
	AgentVersion: whereHelpernull_String{field: "\"connections\".\"agent_version\""},
	DialAttempt:  whereHelpernull_Time{field: "\"connections\".\"dial_attempt\""},
	Latency:      whereHelpernull_String{field: "\"connections\".\"latency\""},
	IsSucceed:    whereHelpernull_Bool{field: "\"connections\".\"is_succeed\""},
}

// ConnectionRels is where relationship names are stored.
var ConnectionRels = struct {
}{}

// connectionR is where relationships are stored.
type connectionR struct {
}

// NewStruct creates a new relationship struct
func (*connectionR) NewStruct() *connectionR {
	return &connectionR{}
}

// connectionL is where Load methods for each relationship are stored.
type connectionL struct{}

var (
	connectionAllColumns            = []string{"id", "peer_id", "multi_address", "agent_version", "dial_attempt", "latency", "is_succeed"}
	connectionColumnsWithoutDefault = []string{"peer_id", "multi_address", "agent_version", "dial_attempt", "latency", "is_succeed"}
	connectionColumnsWithDefault    = []string{"id"}
	connectionPrimaryKeyColumns     = []string{"id"}
)

type (
	// ConnectionSlice is an alias for a slice of pointers to Connection.
	// This should almost always be used instead of []Connection.
	ConnectionSlice []*Connection
	// ConnectionHook is the signature for custom Connection hook methods
	ConnectionHook func(context.Context, boil.ContextExecutor, *Connection) error

	connectionQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	connectionType                 = reflect.TypeOf(&Connection{})
	connectionMapping              = queries.MakeStructMapping(connectionType)
	connectionPrimaryKeyMapping, _ = queries.BindMapping(connectionType, connectionMapping, connectionPrimaryKeyColumns)
	connectionInsertCacheMut       sync.RWMutex
	connectionInsertCache          = make(map[string]insertCache)
	connectionUpdateCacheMut       sync.RWMutex
	connectionUpdateCache          = make(map[string]updateCache)
	connectionUpsertCacheMut       sync.RWMutex
	connectionUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var connectionBeforeInsertHooks []ConnectionHook
var connectionBeforeUpdateHooks []ConnectionHook
var connectionBeforeDeleteHooks []ConnectionHook
var connectionBeforeUpsertHooks []ConnectionHook

var connectionAfterInsertHooks []ConnectionHook
var connectionAfterSelectHooks []ConnectionHook
var connectionAfterUpdateHooks []ConnectionHook
var connectionAfterDeleteHooks []ConnectionHook
var connectionAfterUpsertHooks []ConnectionHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Connection) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Connection) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Connection) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Connection) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Connection) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Connection) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Connection) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Connection) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Connection) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range connectionAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddConnectionHook registers your hook function for all future operations.
func AddConnectionHook(hookPoint boil.HookPoint, connectionHook ConnectionHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		connectionBeforeInsertHooks = append(connectionBeforeInsertHooks, connectionHook)
	case boil.BeforeUpdateHook:
		connectionBeforeUpdateHooks = append(connectionBeforeUpdateHooks, connectionHook)
	case boil.BeforeDeleteHook:
		connectionBeforeDeleteHooks = append(connectionBeforeDeleteHooks, connectionHook)
	case boil.BeforeUpsertHook:
		connectionBeforeUpsertHooks = append(connectionBeforeUpsertHooks, connectionHook)
	case boil.AfterInsertHook:
		connectionAfterInsertHooks = append(connectionAfterInsertHooks, connectionHook)
	case boil.AfterSelectHook:
		connectionAfterSelectHooks = append(connectionAfterSelectHooks, connectionHook)
	case boil.AfterUpdateHook:
		connectionAfterUpdateHooks = append(connectionAfterUpdateHooks, connectionHook)
	case boil.AfterDeleteHook:
		connectionAfterDeleteHooks = append(connectionAfterDeleteHooks, connectionHook)
	case boil.AfterUpsertHook:
		connectionAfterUpsertHooks = append(connectionAfterUpsertHooks, connectionHook)
	}
}

// One returns a single connection record from the query.
func (q connectionQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Connection, error) {
	o := &Connection{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for connections")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Connection records from the query.
func (q connectionQuery) All(ctx context.Context, exec boil.ContextExecutor) (ConnectionSlice, error) {
	var o []*Connection

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Connection slice")
	}

	if len(connectionAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Connection records in the query.
func (q connectionQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count connections rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q connectionQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if connections exists")
	}

	return count > 0, nil
}

// Connections retrieves all the records using an executor.
func Connections(mods ...qm.QueryMod) connectionQuery {
	mods = append(mods, qm.From("\"connections\""))
	return connectionQuery{NewQuery(mods...)}
}

// FindConnection retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindConnection(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Connection, error) {
	connectionObj := &Connection{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"connections\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, connectionObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from connections")
	}

	if err = connectionObj.doAfterSelectHooks(ctx, exec); err != nil {
		return connectionObj, err
	}

	return connectionObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Connection) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no connections provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(connectionColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	connectionInsertCacheMut.RLock()
	cache, cached := connectionInsertCache[key]
	connectionInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			connectionAllColumns,
			connectionColumnsWithDefault,
			connectionColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(connectionType, connectionMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(connectionType, connectionMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"connections\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"connections\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into connections")
	}

	if !cached {
		connectionInsertCacheMut.Lock()
		connectionInsertCache[key] = cache
		connectionInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Connection.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Connection) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	connectionUpdateCacheMut.RLock()
	cache, cached := connectionUpdateCache[key]
	connectionUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			connectionAllColumns,
			connectionPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update connections, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"connections\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, connectionPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(connectionType, connectionMapping, append(wl, connectionPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update connections row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for connections")
	}

	if !cached {
		connectionUpdateCacheMut.Lock()
		connectionUpdateCache[key] = cache
		connectionUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q connectionQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for connections")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for connections")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ConnectionSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), connectionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"connections\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, connectionPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in connection slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all connection")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Connection) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no connections provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(connectionColumnsWithDefault, o)

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

	connectionUpsertCacheMut.RLock()
	cache, cached := connectionUpsertCache[key]
	connectionUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			connectionAllColumns,
			connectionColumnsWithDefault,
			connectionColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			connectionAllColumns,
			connectionPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert connections, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(connectionPrimaryKeyColumns))
			copy(conflict, connectionPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"connections\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(connectionType, connectionMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(connectionType, connectionMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert connections")
	}

	if !cached {
		connectionUpsertCacheMut.Lock()
		connectionUpsertCache[key] = cache
		connectionUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Connection record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Connection) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Connection provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), connectionPrimaryKeyMapping)
	sql := "DELETE FROM \"connections\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from connections")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for connections")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q connectionQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no connectionQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from connections")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for connections")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ConnectionSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(connectionBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), connectionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"connections\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, connectionPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from connection slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for connections")
	}

	if len(connectionAfterDeleteHooks) != 0 {
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
func (o *Connection) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindConnection(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ConnectionSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ConnectionSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), connectionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"connections\".* FROM \"connections\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, connectionPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in ConnectionSlice")
	}

	*o = slice

	return nil
}

// ConnectionExists checks if the Connection row exists.
func ConnectionExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"connections\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if connections exists")
	}

	return exists, nil
}