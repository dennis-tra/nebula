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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Latency is an object representing the database table.
type Latency struct {
	ID              int       `boil:"id" json:"id" toml:"id" yaml:"id"`
	PeerID          string    `boil:"peer_id" json:"peer_id" toml:"peer_id" yaml:"peer_id"`
	Address         string    `boil:"address" json:"address" toml:"address" yaml:"address"`
	PingLatencySAvg float64   `boil:"ping_latency_s_avg" json:"ping_latency_s_avg" toml:"ping_latency_s_avg" yaml:"ping_latency_s_avg"`
	PingLatencySSTD float64   `boil:"ping_latency_s_std" json:"ping_latency_s_std" toml:"ping_latency_s_std" yaml:"ping_latency_s_std"`
	PingLatencySMin float64   `boil:"ping_latency_s_min" json:"ping_latency_s_min" toml:"ping_latency_s_min" yaml:"ping_latency_s_min"`
	PingLatencySMax float64   `boil:"ping_latency_s_max" json:"ping_latency_s_max" toml:"ping_latency_s_max" yaml:"ping_latency_s_max"`
	PingPacketsSent int       `boil:"ping_packets_sent" json:"ping_packets_sent" toml:"ping_packets_sent" yaml:"ping_packets_sent"`
	PingPacketsRecv int       `boil:"ping_packets_recv" json:"ping_packets_recv" toml:"ping_packets_recv" yaml:"ping_packets_recv"`
	PingPacketsDupl int       `boil:"ping_packets_dupl" json:"ping_packets_dupl" toml:"ping_packets_dupl" yaml:"ping_packets_dupl"`
	PingPacketLoss  float64   `boil:"ping_packet_loss" json:"ping_packet_loss" toml:"ping_packet_loss" yaml:"ping_packet_loss"`
	UpdatedAt       time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt       time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *latencyR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L latencyL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var LatencyColumns = struct {
	ID              string
	PeerID          string
	Address         string
	PingLatencySAvg string
	PingLatencySSTD string
	PingLatencySMin string
	PingLatencySMax string
	PingPacketsSent string
	PingPacketsRecv string
	PingPacketsDupl string
	PingPacketLoss  string
	UpdatedAt       string
	CreatedAt       string
}{
	ID:              "id",
	PeerID:          "peer_id",
	Address:         "address",
	PingLatencySAvg: "ping_latency_s_avg",
	PingLatencySSTD: "ping_latency_s_std",
	PingLatencySMin: "ping_latency_s_min",
	PingLatencySMax: "ping_latency_s_max",
	PingPacketsSent: "ping_packets_sent",
	PingPacketsRecv: "ping_packets_recv",
	PingPacketsDupl: "ping_packets_dupl",
	PingPacketLoss:  "ping_packet_loss",
	UpdatedAt:       "updated_at",
	CreatedAt:       "created_at",
}

var LatencyTableColumns = struct {
	ID              string
	PeerID          string
	Address         string
	PingLatencySAvg string
	PingLatencySSTD string
	PingLatencySMin string
	PingLatencySMax string
	PingPacketsSent string
	PingPacketsRecv string
	PingPacketsDupl string
	PingPacketLoss  string
	UpdatedAt       string
	CreatedAt       string
}{
	ID:              "latencies.id",
	PeerID:          "latencies.peer_id",
	Address:         "latencies.address",
	PingLatencySAvg: "latencies.ping_latency_s_avg",
	PingLatencySSTD: "latencies.ping_latency_s_std",
	PingLatencySMin: "latencies.ping_latency_s_min",
	PingLatencySMax: "latencies.ping_latency_s_max",
	PingPacketsSent: "latencies.ping_packets_sent",
	PingPacketsRecv: "latencies.ping_packets_recv",
	PingPacketsDupl: "latencies.ping_packets_dupl",
	PingPacketLoss:  "latencies.ping_packet_loss",
	UpdatedAt:       "latencies.updated_at",
	CreatedAt:       "latencies.created_at",
}

// Generated where

type whereHelperfloat64 struct{ field string }

func (w whereHelperfloat64) EQ(x float64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperfloat64) NEQ(x float64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelperfloat64) LT(x float64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperfloat64) LTE(x float64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelperfloat64) GT(x float64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperfloat64) GTE(x float64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}
func (w whereHelperfloat64) IN(slice []float64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperfloat64) NIN(slice []float64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

var LatencyWhere = struct {
	ID              whereHelperint
	PeerID          whereHelperstring
	Address         whereHelperstring
	PingLatencySAvg whereHelperfloat64
	PingLatencySSTD whereHelperfloat64
	PingLatencySMin whereHelperfloat64
	PingLatencySMax whereHelperfloat64
	PingPacketsSent whereHelperint
	PingPacketsRecv whereHelperint
	PingPacketsDupl whereHelperint
	PingPacketLoss  whereHelperfloat64
	UpdatedAt       whereHelpertime_Time
	CreatedAt       whereHelpertime_Time
}{
	ID:              whereHelperint{field: "\"latencies\".\"id\""},
	PeerID:          whereHelperstring{field: "\"latencies\".\"peer_id\""},
	Address:         whereHelperstring{field: "\"latencies\".\"address\""},
	PingLatencySAvg: whereHelperfloat64{field: "\"latencies\".\"ping_latency_s_avg\""},
	PingLatencySSTD: whereHelperfloat64{field: "\"latencies\".\"ping_latency_s_std\""},
	PingLatencySMin: whereHelperfloat64{field: "\"latencies\".\"ping_latency_s_min\""},
	PingLatencySMax: whereHelperfloat64{field: "\"latencies\".\"ping_latency_s_max\""},
	PingPacketsSent: whereHelperint{field: "\"latencies\".\"ping_packets_sent\""},
	PingPacketsRecv: whereHelperint{field: "\"latencies\".\"ping_packets_recv\""},
	PingPacketsDupl: whereHelperint{field: "\"latencies\".\"ping_packets_dupl\""},
	PingPacketLoss:  whereHelperfloat64{field: "\"latencies\".\"ping_packet_loss\""},
	UpdatedAt:       whereHelpertime_Time{field: "\"latencies\".\"updated_at\""},
	CreatedAt:       whereHelpertime_Time{field: "\"latencies\".\"created_at\""},
}

// LatencyRels is where relationship names are stored.
var LatencyRels = struct {
	Peer string
}{
	Peer: "Peer",
}

// latencyR is where relationships are stored.
type latencyR struct {
	Peer *Peer `boil:"Peer" json:"Peer" toml:"Peer" yaml:"Peer"`
}

// NewStruct creates a new relationship struct
func (*latencyR) NewStruct() *latencyR {
	return &latencyR{}
}

// latencyL is where Load methods for each relationship are stored.
type latencyL struct{}

var (
	latencyAllColumns            = []string{"id", "peer_id", "address", "ping_latency_s_avg", "ping_latency_s_std", "ping_latency_s_min", "ping_latency_s_max", "ping_packets_sent", "ping_packets_recv", "ping_packets_dupl", "ping_packet_loss", "updated_at", "created_at"}
	latencyColumnsWithoutDefault = []string{"peer_id", "address", "ping_latency_s_avg", "ping_latency_s_std", "ping_latency_s_min", "ping_latency_s_max", "ping_packets_sent", "ping_packets_recv", "ping_packets_dupl", "ping_packet_loss", "updated_at", "created_at"}
	latencyColumnsWithDefault    = []string{"id"}
	latencyPrimaryKeyColumns     = []string{"id"}
)

type (
	// LatencySlice is an alias for a slice of pointers to Latency.
	// This should almost always be used instead of []Latency.
	LatencySlice []*Latency
	// LatencyHook is the signature for custom Latency hook methods
	LatencyHook func(context.Context, boil.ContextExecutor, *Latency) error

	latencyQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	latencyType                 = reflect.TypeOf(&Latency{})
	latencyMapping              = queries.MakeStructMapping(latencyType)
	latencyPrimaryKeyMapping, _ = queries.BindMapping(latencyType, latencyMapping, latencyPrimaryKeyColumns)
	latencyInsertCacheMut       sync.RWMutex
	latencyInsertCache          = make(map[string]insertCache)
	latencyUpdateCacheMut       sync.RWMutex
	latencyUpdateCache          = make(map[string]updateCache)
	latencyUpsertCacheMut       sync.RWMutex
	latencyUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var latencyBeforeInsertHooks []LatencyHook
var latencyBeforeUpdateHooks []LatencyHook
var latencyBeforeDeleteHooks []LatencyHook
var latencyBeforeUpsertHooks []LatencyHook

var latencyAfterInsertHooks []LatencyHook
var latencyAfterSelectHooks []LatencyHook
var latencyAfterUpdateHooks []LatencyHook
var latencyAfterDeleteHooks []LatencyHook
var latencyAfterUpsertHooks []LatencyHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Latency) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Latency) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Latency) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Latency) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Latency) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Latency) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Latency) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Latency) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Latency) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range latencyAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddLatencyHook registers your hook function for all future operations.
func AddLatencyHook(hookPoint boil.HookPoint, latencyHook LatencyHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		latencyBeforeInsertHooks = append(latencyBeforeInsertHooks, latencyHook)
	case boil.BeforeUpdateHook:
		latencyBeforeUpdateHooks = append(latencyBeforeUpdateHooks, latencyHook)
	case boil.BeforeDeleteHook:
		latencyBeforeDeleteHooks = append(latencyBeforeDeleteHooks, latencyHook)
	case boil.BeforeUpsertHook:
		latencyBeforeUpsertHooks = append(latencyBeforeUpsertHooks, latencyHook)
	case boil.AfterInsertHook:
		latencyAfterInsertHooks = append(latencyAfterInsertHooks, latencyHook)
	case boil.AfterSelectHook:
		latencyAfterSelectHooks = append(latencyAfterSelectHooks, latencyHook)
	case boil.AfterUpdateHook:
		latencyAfterUpdateHooks = append(latencyAfterUpdateHooks, latencyHook)
	case boil.AfterDeleteHook:
		latencyAfterDeleteHooks = append(latencyAfterDeleteHooks, latencyHook)
	case boil.AfterUpsertHook:
		latencyAfterUpsertHooks = append(latencyAfterUpsertHooks, latencyHook)
	}
}

// One returns a single latency record from the query.
func (q latencyQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Latency, error) {
	o := &Latency{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for latencies")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Latency records from the query.
func (q latencyQuery) All(ctx context.Context, exec boil.ContextExecutor) (LatencySlice, error) {
	var o []*Latency

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Latency slice")
	}

	if len(latencyAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Latency records in the query.
func (q latencyQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count latencies rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q latencyQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if latencies exists")
	}

	return count > 0, nil
}

// Peer pointed to by the foreign key.
func (o *Latency) Peer(mods ...qm.QueryMod) peerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PeerID),
	}

	queryMods = append(queryMods, mods...)

	query := Peers(queryMods...)
	queries.SetFrom(query.Query, "\"peers\"")

	return query
}

// LoadPeer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (latencyL) LoadPeer(ctx context.Context, e boil.ContextExecutor, singular bool, maybeLatency interface{}, mods queries.Applicator) error {
	var slice []*Latency
	var object *Latency

	if singular {
		object = maybeLatency.(*Latency)
	} else {
		slice = *maybeLatency.(*[]*Latency)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &latencyR{}
		}
		args = append(args, object.PeerID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &latencyR{}
			}

			for _, a := range args {
				if a == obj.PeerID {
					continue Outer
				}
			}

			args = append(args, obj.PeerID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`peers`),
		qm.WhereIn(`peers.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Peer")
	}

	var resultSlice []*Peer
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Peer")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for peers")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for peers")
	}

	if len(latencyAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Peer = foreign
		if foreign.R == nil {
			foreign.R = &peerR{}
		}
		foreign.R.Latencies = append(foreign.R.Latencies, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PeerID == foreign.ID {
				local.R.Peer = foreign
				if foreign.R == nil {
					foreign.R = &peerR{}
				}
				foreign.R.Latencies = append(foreign.R.Latencies, local)
				break
			}
		}
	}

	return nil
}

// SetPeer of the latency to the related item.
// Sets o.R.Peer to related.
// Adds o to related.R.Latencies.
func (o *Latency) SetPeer(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Peer) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"latencies\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"peer_id"}),
		strmangle.WhereClause("\"", "\"", 2, latencyPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.PeerID = related.ID
	if o.R == nil {
		o.R = &latencyR{
			Peer: related,
		}
	} else {
		o.R.Peer = related
	}

	if related.R == nil {
		related.R = &peerR{
			Latencies: LatencySlice{o},
		}
	} else {
		related.R.Latencies = append(related.R.Latencies, o)
	}

	return nil
}

// Latencies retrieves all the records using an executor.
func Latencies(mods ...qm.QueryMod) latencyQuery {
	mods = append(mods, qm.From("\"latencies\""))
	return latencyQuery{NewQuery(mods...)}
}

// FindLatency retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindLatency(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Latency, error) {
	latencyObj := &Latency{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"latencies\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, latencyObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from latencies")
	}

	if err = latencyObj.doAfterSelectHooks(ctx, exec); err != nil {
		return latencyObj, err
	}

	return latencyObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Latency) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no latencies provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(latencyColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	latencyInsertCacheMut.RLock()
	cache, cached := latencyInsertCache[key]
	latencyInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			latencyAllColumns,
			latencyColumnsWithDefault,
			latencyColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(latencyType, latencyMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(latencyType, latencyMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"latencies\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"latencies\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into latencies")
	}

	if !cached {
		latencyInsertCacheMut.Lock()
		latencyInsertCache[key] = cache
		latencyInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Latency.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Latency) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	latencyUpdateCacheMut.RLock()
	cache, cached := latencyUpdateCache[key]
	latencyUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			latencyAllColumns,
			latencyPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update latencies, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"latencies\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, latencyPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(latencyType, latencyMapping, append(wl, latencyPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update latencies row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for latencies")
	}

	if !cached {
		latencyUpdateCacheMut.Lock()
		latencyUpdateCache[key] = cache
		latencyUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q latencyQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for latencies")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for latencies")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o LatencySlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), latencyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"latencies\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, latencyPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in latency slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all latency")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Latency) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no latencies provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(latencyColumnsWithDefault, o)

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

	latencyUpsertCacheMut.RLock()
	cache, cached := latencyUpsertCache[key]
	latencyUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			latencyAllColumns,
			latencyColumnsWithDefault,
			latencyColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			latencyAllColumns,
			latencyPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert latencies, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(latencyPrimaryKeyColumns))
			copy(conflict, latencyPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"latencies\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(latencyType, latencyMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(latencyType, latencyMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert latencies")
	}

	if !cached {
		latencyUpsertCacheMut.Lock()
		latencyUpsertCache[key] = cache
		latencyUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Latency record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Latency) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Latency provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), latencyPrimaryKeyMapping)
	sql := "DELETE FROM \"latencies\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from latencies")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for latencies")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q latencyQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no latencyQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from latencies")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for latencies")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o LatencySlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(latencyBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), latencyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"latencies\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, latencyPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from latency slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for latencies")
	}

	if len(latencyAfterDeleteHooks) != 0 {
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
func (o *Latency) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindLatency(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *LatencySlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := LatencySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), latencyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"latencies\".* FROM \"latencies\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, latencyPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in LatencySlice")
	}

	*o = slice

	return nil
}

// LatencyExists checks if the Latency row exists.
func LatencyExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"latencies\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if latencies exists")
	}

	return exists, nil
}
