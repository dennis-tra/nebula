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
	"github.com/volatiletech/strmangle"
)

// Session is an object representing the database table.
type Session struct {
	ID                    int         `boil:"id" json:"id" toml:"id" yaml:"id"`
	PeerID                int         `boil:"peer_id" json:"peer_id" toml:"peer_id" yaml:"peer_id"`
	FirstSuccessfulVisit  time.Time   `boil:"first_successful_visit" json:"first_successful_visit" toml:"first_successful_visit" yaml:"first_successful_visit"`
	LastSuccessfulVisit   time.Time   `boil:"last_successful_visit" json:"last_successful_visit" toml:"last_successful_visit" yaml:"last_successful_visit"`
	NextVisitDueAt        null.Time   `boil:"next_visit_due_at" json:"next_visit_due_at,omitempty" toml:"next_visit_due_at" yaml:"next_visit_due_at,omitempty"`
	FirstFailedVisit      null.Time   `boil:"first_failed_visit" json:"first_failed_visit,omitempty" toml:"first_failed_visit" yaml:"first_failed_visit,omitempty"`
	LastFailedVisit       null.Time   `boil:"last_failed_visit" json:"last_failed_visit,omitempty" toml:"last_failed_visit" yaml:"last_failed_visit,omitempty"`
	LastVisitedAt         time.Time   `boil:"last_visited_at" json:"last_visited_at" toml:"last_visited_at" yaml:"last_visited_at"`
	UpdatedAt             time.Time   `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt             time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	SuccessfulVisitsCount int         `boil:"successful_visits_count" json:"successful_visits_count" toml:"successful_visits_count" yaml:"successful_visits_count"`
	RecoveredCount        int         `boil:"recovered_count" json:"recovered_count" toml:"recovered_count" yaml:"recovered_count"`
	State                 string      `boil:"state" json:"state" toml:"state" yaml:"state"`
	FailedVisitsCount     int16       `boil:"failed_visits_count" json:"failed_visits_count" toml:"failed_visits_count" yaml:"failed_visits_count"`
	FinishReason          null.String `boil:"finish_reason" json:"finish_reason,omitempty" toml:"finish_reason" yaml:"finish_reason,omitempty"`
	Uptime                string      `boil:"uptime" json:"uptime" toml:"uptime" yaml:"uptime"`

	R *sessionR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L sessionL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var SessionColumns = struct {
	ID                    string
	PeerID                string
	FirstSuccessfulVisit  string
	LastSuccessfulVisit   string
	NextVisitDueAt        string
	FirstFailedVisit      string
	LastFailedVisit       string
	LastVisitedAt         string
	UpdatedAt             string
	CreatedAt             string
	SuccessfulVisitsCount string
	RecoveredCount        string
	State                 string
	FailedVisitsCount     string
	FinishReason          string
	Uptime                string
}{
	ID:                    "id",
	PeerID:                "peer_id",
	FirstSuccessfulVisit:  "first_successful_visit",
	LastSuccessfulVisit:   "last_successful_visit",
	NextVisitDueAt:        "next_visit_due_at",
	FirstFailedVisit:      "first_failed_visit",
	LastFailedVisit:       "last_failed_visit",
	LastVisitedAt:         "last_visited_at",
	UpdatedAt:             "updated_at",
	CreatedAt:             "created_at",
	SuccessfulVisitsCount: "successful_visits_count",
	RecoveredCount:        "recovered_count",
	State:                 "state",
	FailedVisitsCount:     "failed_visits_count",
	FinishReason:          "finish_reason",
	Uptime:                "uptime",
}

var SessionTableColumns = struct {
	ID                    string
	PeerID                string
	FirstSuccessfulVisit  string
	LastSuccessfulVisit   string
	NextVisitDueAt        string
	FirstFailedVisit      string
	LastFailedVisit       string
	LastVisitedAt         string
	UpdatedAt             string
	CreatedAt             string
	SuccessfulVisitsCount string
	RecoveredCount        string
	State                 string
	FailedVisitsCount     string
	FinishReason          string
	Uptime                string
}{
	ID:                    "sessions.id",
	PeerID:                "sessions.peer_id",
	FirstSuccessfulVisit:  "sessions.first_successful_visit",
	LastSuccessfulVisit:   "sessions.last_successful_visit",
	NextVisitDueAt:        "sessions.next_visit_due_at",
	FirstFailedVisit:      "sessions.first_failed_visit",
	LastFailedVisit:       "sessions.last_failed_visit",
	LastVisitedAt:         "sessions.last_visited_at",
	UpdatedAt:             "sessions.updated_at",
	CreatedAt:             "sessions.created_at",
	SuccessfulVisitsCount: "sessions.successful_visits_count",
	RecoveredCount:        "sessions.recovered_count",
	State:                 "sessions.state",
	FailedVisitsCount:     "sessions.failed_visits_count",
	FinishReason:          "sessions.finish_reason",
	Uptime:                "sessions.uptime",
}

// Generated where

var SessionWhere = struct {
	ID                    whereHelperint
	PeerID                whereHelperint
	FirstSuccessfulVisit  whereHelpertime_Time
	LastSuccessfulVisit   whereHelpertime_Time
	NextVisitDueAt        whereHelpernull_Time
	FirstFailedVisit      whereHelpernull_Time
	LastFailedVisit       whereHelpernull_Time
	LastVisitedAt         whereHelpertime_Time
	UpdatedAt             whereHelpertime_Time
	CreatedAt             whereHelpertime_Time
	SuccessfulVisitsCount whereHelperint
	RecoveredCount        whereHelperint
	State                 whereHelperstring
	FailedVisitsCount     whereHelperint16
	FinishReason          whereHelpernull_String
	Uptime                whereHelperstring
}{
	ID:                    whereHelperint{field: "\"sessions\".\"id\""},
	PeerID:                whereHelperint{field: "\"sessions\".\"peer_id\""},
	FirstSuccessfulVisit:  whereHelpertime_Time{field: "\"sessions\".\"first_successful_visit\""},
	LastSuccessfulVisit:   whereHelpertime_Time{field: "\"sessions\".\"last_successful_visit\""},
	NextVisitDueAt:        whereHelpernull_Time{field: "\"sessions\".\"next_visit_due_at\""},
	FirstFailedVisit:      whereHelpernull_Time{field: "\"sessions\".\"first_failed_visit\""},
	LastFailedVisit:       whereHelpernull_Time{field: "\"sessions\".\"last_failed_visit\""},
	LastVisitedAt:         whereHelpertime_Time{field: "\"sessions\".\"last_visited_at\""},
	UpdatedAt:             whereHelpertime_Time{field: "\"sessions\".\"updated_at\""},
	CreatedAt:             whereHelpertime_Time{field: "\"sessions\".\"created_at\""},
	SuccessfulVisitsCount: whereHelperint{field: "\"sessions\".\"successful_visits_count\""},
	RecoveredCount:        whereHelperint{field: "\"sessions\".\"recovered_count\""},
	State:                 whereHelperstring{field: "\"sessions\".\"state\""},
	FailedVisitsCount:     whereHelperint16{field: "\"sessions\".\"failed_visits_count\""},
	FinishReason:          whereHelpernull_String{field: "\"sessions\".\"finish_reason\""},
	Uptime:                whereHelperstring{field: "\"sessions\".\"uptime\""},
}

// SessionRels is where relationship names are stored.
var SessionRels = struct {
}{}

// sessionR is where relationships are stored.
type sessionR struct {
}

// NewStruct creates a new relationship struct
func (*sessionR) NewStruct() *sessionR {
	return &sessionR{}
}

// sessionL is where Load methods for each relationship are stored.
type sessionL struct{}

var (
	sessionAllColumns            = []string{"id", "peer_id", "first_successful_visit", "last_successful_visit", "next_visit_due_at", "first_failed_visit", "last_failed_visit", "last_visited_at", "updated_at", "created_at", "successful_visits_count", "recovered_count", "state", "failed_visits_count", "finish_reason", "uptime"}
	sessionColumnsWithoutDefault = []string{"peer_id", "first_successful_visit", "last_successful_visit", "last_visited_at", "updated_at", "created_at", "successful_visits_count", "recovered_count", "state", "failed_visits_count", "uptime"}
	sessionColumnsWithDefault    = []string{"id", "next_visit_due_at", "first_failed_visit", "last_failed_visit", "finish_reason"}
	sessionPrimaryKeyColumns     = []string{"id", "state", "last_visited_at"}
	sessionGeneratedColumns      = []string{"id"}
)

type (
	// SessionSlice is an alias for a slice of pointers to Session.
	// This should almost always be used instead of []Session.
	SessionSlice []*Session
	// SessionHook is the signature for custom Session hook methods
	SessionHook func(context.Context, boil.ContextExecutor, *Session) error

	sessionQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	sessionType                 = reflect.TypeOf(&Session{})
	sessionMapping              = queries.MakeStructMapping(sessionType)
	sessionPrimaryKeyMapping, _ = queries.BindMapping(sessionType, sessionMapping, sessionPrimaryKeyColumns)
	sessionInsertCacheMut       sync.RWMutex
	sessionInsertCache          = make(map[string]insertCache)
	sessionUpdateCacheMut       sync.RWMutex
	sessionUpdateCache          = make(map[string]updateCache)
	sessionUpsertCacheMut       sync.RWMutex
	sessionUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var sessionAfterSelectHooks []SessionHook

var sessionBeforeInsertHooks []SessionHook
var sessionAfterInsertHooks []SessionHook

var sessionBeforeUpdateHooks []SessionHook
var sessionAfterUpdateHooks []SessionHook

var sessionBeforeDeleteHooks []SessionHook
var sessionAfterDeleteHooks []SessionHook

var sessionBeforeUpsertHooks []SessionHook
var sessionAfterUpsertHooks []SessionHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Session) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Session) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Session) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Session) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Session) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Session) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Session) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Session) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Session) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sessionAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddSessionHook registers your hook function for all future operations.
func AddSessionHook(hookPoint boil.HookPoint, sessionHook SessionHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		sessionAfterSelectHooks = append(sessionAfterSelectHooks, sessionHook)
	case boil.BeforeInsertHook:
		sessionBeforeInsertHooks = append(sessionBeforeInsertHooks, sessionHook)
	case boil.AfterInsertHook:
		sessionAfterInsertHooks = append(sessionAfterInsertHooks, sessionHook)
	case boil.BeforeUpdateHook:
		sessionBeforeUpdateHooks = append(sessionBeforeUpdateHooks, sessionHook)
	case boil.AfterUpdateHook:
		sessionAfterUpdateHooks = append(sessionAfterUpdateHooks, sessionHook)
	case boil.BeforeDeleteHook:
		sessionBeforeDeleteHooks = append(sessionBeforeDeleteHooks, sessionHook)
	case boil.AfterDeleteHook:
		sessionAfterDeleteHooks = append(sessionAfterDeleteHooks, sessionHook)
	case boil.BeforeUpsertHook:
		sessionBeforeUpsertHooks = append(sessionBeforeUpsertHooks, sessionHook)
	case boil.AfterUpsertHook:
		sessionAfterUpsertHooks = append(sessionAfterUpsertHooks, sessionHook)
	}
}

// One returns a single session record from the query.
func (q sessionQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Session, error) {
	o := &Session{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "pg: failed to execute a one query for sessions")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Session records from the query.
func (q sessionQuery) All(ctx context.Context, exec boil.ContextExecutor) (SessionSlice, error) {
	var o []*Session

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "pg: failed to assign all query results to Session slice")
	}

	if len(sessionAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Session records in the query.
func (q sessionQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to count sessions rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q sessionQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "pg: failed to check if sessions exists")
	}

	return count > 0, nil
}

// Sessions retrieves all the records using an executor.
func Sessions(mods ...qm.QueryMod) sessionQuery {
	mods = append(mods, qm.From("\"sessions\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"sessions\".*"})
	}

	return sessionQuery{q}
}

// FindSession retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindSession(ctx context.Context, exec boil.ContextExecutor, iD int, state string, lastVisitedAt time.Time, selectCols ...string) (*Session, error) {
	sessionObj := &Session{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"sessions\" where \"id\"=$1 AND \"state\"=$2 AND \"last_visited_at\"=$3", sel,
	)

	q := queries.Raw(query, iD, state, lastVisitedAt)

	err := q.Bind(ctx, exec, sessionObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "pg: unable to select from sessions")
	}

	if err = sessionObj.doAfterSelectHooks(ctx, exec); err != nil {
		return sessionObj, err
	}

	return sessionObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Session) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("pg: no sessions provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = currTime
		}
		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(sessionColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	sessionInsertCacheMut.RLock()
	cache, cached := sessionInsertCache[key]
	sessionInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			sessionAllColumns,
			sessionColumnsWithDefault,
			sessionColumnsWithoutDefault,
			nzDefaults,
		)
		wl = strmangle.SetComplement(wl, sessionGeneratedColumns)

		cache.valueMapping, err = queries.BindMapping(sessionType, sessionMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(sessionType, sessionMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"sessions\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"sessions\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "pg: unable to insert into sessions")
	}

	if !cached {
		sessionInsertCacheMut.Lock()
		sessionInsertCache[key] = cache
		sessionInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Session.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Session) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	sessionUpdateCacheMut.RLock()
	cache, cached := sessionUpdateCache[key]
	sessionUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			sessionAllColumns,
			sessionPrimaryKeyColumns,
		)
		wl = strmangle.SetComplement(wl, sessionGeneratedColumns)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("pg: unable to update sessions, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"sessions\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, sessionPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(sessionType, sessionMapping, append(wl, sessionPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "pg: unable to update sessions row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to get rows affected by update for sessions")
	}

	if !cached {
		sessionUpdateCacheMut.Lock()
		sessionUpdateCache[key] = cache
		sessionUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q sessionQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to update all for sessions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to retrieve rows affected for sessions")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o SessionSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), sessionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"sessions\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, sessionPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to update all in session slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to retrieve rows affected all in update all session")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Session) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("pg: no sessions provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(sessionColumnsWithDefault, o)

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

	sessionUpsertCacheMut.RLock()
	cache, cached := sessionUpsertCache[key]
	sessionUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			sessionAllColumns,
			sessionColumnsWithDefault,
			sessionColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			sessionAllColumns,
			sessionPrimaryKeyColumns,
		)

		insert = strmangle.SetComplement(insert, sessionGeneratedColumns)
		update = strmangle.SetComplement(update, sessionGeneratedColumns)

		if updateOnConflict && len(update) == 0 {
			return errors.New("pg: unable to upsert sessions, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(sessionPrimaryKeyColumns))
			copy(conflict, sessionPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"sessions\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(sessionType, sessionMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(sessionType, sessionMapping, ret)
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
		return errors.Wrap(err, "pg: unable to upsert sessions")
	}

	if !cached {
		sessionUpsertCacheMut.Lock()
		sessionUpsertCache[key] = cache
		sessionUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Session record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Session) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("pg: no Session provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), sessionPrimaryKeyMapping)
	sql := "DELETE FROM \"sessions\" WHERE \"id\"=$1 AND \"state\"=$2 AND \"last_visited_at\"=$3"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to delete from sessions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to get rows affected by delete for sessions")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q sessionQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("pg: no sessionQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to delete all from sessions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to get rows affected by deleteall for sessions")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o SessionSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(sessionBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), sessionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"sessions\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, sessionPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "pg: unable to delete all from session slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "pg: failed to get rows affected by deleteall for sessions")
	}

	if len(sessionAfterDeleteHooks) != 0 {
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
func (o *Session) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindSession(ctx, exec, o.ID, o.State, o.LastVisitedAt)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *SessionSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := SessionSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), sessionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"sessions\".* FROM \"sessions\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, sessionPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "pg: unable to reload all in SessionSlice")
	}

	*o = slice

	return nil
}

// SessionExists checks if the Session row exists.
func SessionExists(ctx context.Context, exec boil.ContextExecutor, iD int, state string, lastVisitedAt time.Time) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"sessions\" where \"id\"=$1 AND \"state\"=$2 AND \"last_visited_at\"=$3 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD, state, lastVisitedAt)
	}
	row := exec.QueryRowContext(ctx, sql, iD, state, lastVisitedAt)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "pg: unable to check if sessions exists")
	}

	return exists, nil
}

// Exists checks if the Session row exists.
func (o *Session) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return SessionExists(ctx, exec, o.ID, o.State, o.LastVisitedAt)
}
