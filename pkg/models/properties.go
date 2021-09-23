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

// Property is an object representing the database table.
type Property struct {
	ID        int       `boil:"id" json:"id" toml:"id" yaml:"id"`
	Property  string    `boil:"property" json:"property" toml:"property" yaml:"property"`
	Value     string    `boil:"value" json:"value" toml:"value" yaml:"value"`
	UpdatedAt time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *propertyR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L propertyL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PropertyColumns = struct {
	ID        string
	Property  string
	Value     string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "id",
	Property:  "property",
	Value:     "value",
	UpdatedAt: "updated_at",
	CreatedAt: "created_at",
}

var PropertyTableColumns = struct {
	ID        string
	Property  string
	Value     string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "properties.id",
	Property:  "properties.property",
	Value:     "properties.value",
	UpdatedAt: "properties.updated_at",
	CreatedAt: "properties.created_at",
}

// Generated where

var PropertyWhere = struct {
	ID        whereHelperint
	Property  whereHelperstring
	Value     whereHelperstring
	UpdatedAt whereHelpertime_Time
	CreatedAt whereHelpertime_Time
}{
	ID:        whereHelperint{field: "\"properties\".\"id\""},
	Property:  whereHelperstring{field: "\"properties\".\"property\""},
	Value:     whereHelperstring{field: "\"properties\".\"value\""},
	UpdatedAt: whereHelpertime_Time{field: "\"properties\".\"updated_at\""},
	CreatedAt: whereHelpertime_Time{field: "\"properties\".\"created_at\""},
}

// PropertyRels is where relationship names are stored.
var PropertyRels = struct {
	CrawlProperties string
	Peers           string
}{
	CrawlProperties: "CrawlProperties",
	Peers:           "Peers",
}

// propertyR is where relationships are stored.
type propertyR struct {
	CrawlProperties CrawlPropertySlice `boil:"CrawlProperties" json:"CrawlProperties" toml:"CrawlProperties" yaml:"CrawlProperties"`
	Peers           PeerSlice          `boil:"Peers" json:"Peers" toml:"Peers" yaml:"Peers"`
}

// NewStruct creates a new relationship struct
func (*propertyR) NewStruct() *propertyR {
	return &propertyR{}
}

// propertyL is where Load methods for each relationship are stored.
type propertyL struct{}

var (
	propertyAllColumns            = []string{"id", "property", "value", "updated_at", "created_at"}
	propertyColumnsWithoutDefault = []string{"property", "value", "updated_at", "created_at"}
	propertyColumnsWithDefault    = []string{"id"}
	propertyPrimaryKeyColumns     = []string{"id"}
)

type (
	// PropertySlice is an alias for a slice of pointers to Property.
	// This should almost always be used instead of []Property.
	PropertySlice []*Property
	// PropertyHook is the signature for custom Property hook methods
	PropertyHook func(context.Context, boil.ContextExecutor, *Property) error

	propertyQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	propertyType                 = reflect.TypeOf(&Property{})
	propertyMapping              = queries.MakeStructMapping(propertyType)
	propertyPrimaryKeyMapping, _ = queries.BindMapping(propertyType, propertyMapping, propertyPrimaryKeyColumns)
	propertyInsertCacheMut       sync.RWMutex
	propertyInsertCache          = make(map[string]insertCache)
	propertyUpdateCacheMut       sync.RWMutex
	propertyUpdateCache          = make(map[string]updateCache)
	propertyUpsertCacheMut       sync.RWMutex
	propertyUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var propertyBeforeInsertHooks []PropertyHook
var propertyBeforeUpdateHooks []PropertyHook
var propertyBeforeDeleteHooks []PropertyHook
var propertyBeforeUpsertHooks []PropertyHook

var propertyAfterInsertHooks []PropertyHook
var propertyAfterSelectHooks []PropertyHook
var propertyAfterUpdateHooks []PropertyHook
var propertyAfterDeleteHooks []PropertyHook
var propertyAfterUpsertHooks []PropertyHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Property) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Property) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Property) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Property) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Property) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Property) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Property) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Property) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Property) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range propertyAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPropertyHook registers your hook function for all future operations.
func AddPropertyHook(hookPoint boil.HookPoint, propertyHook PropertyHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		propertyBeforeInsertHooks = append(propertyBeforeInsertHooks, propertyHook)
	case boil.BeforeUpdateHook:
		propertyBeforeUpdateHooks = append(propertyBeforeUpdateHooks, propertyHook)
	case boil.BeforeDeleteHook:
		propertyBeforeDeleteHooks = append(propertyBeforeDeleteHooks, propertyHook)
	case boil.BeforeUpsertHook:
		propertyBeforeUpsertHooks = append(propertyBeforeUpsertHooks, propertyHook)
	case boil.AfterInsertHook:
		propertyAfterInsertHooks = append(propertyAfterInsertHooks, propertyHook)
	case boil.AfterSelectHook:
		propertyAfterSelectHooks = append(propertyAfterSelectHooks, propertyHook)
	case boil.AfterUpdateHook:
		propertyAfterUpdateHooks = append(propertyAfterUpdateHooks, propertyHook)
	case boil.AfterDeleteHook:
		propertyAfterDeleteHooks = append(propertyAfterDeleteHooks, propertyHook)
	case boil.AfterUpsertHook:
		propertyAfterUpsertHooks = append(propertyAfterUpsertHooks, propertyHook)
	}
}

// One returns a single property record from the query.
func (q propertyQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Property, error) {
	o := &Property{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for properties")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Property records from the query.
func (q propertyQuery) All(ctx context.Context, exec boil.ContextExecutor) (PropertySlice, error) {
	var o []*Property

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Property slice")
	}

	if len(propertyAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Property records in the query.
func (q propertyQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count properties rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q propertyQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if properties exists")
	}

	return count > 0, nil
}

// CrawlProperties retrieves all the crawl_property's CrawlProperties with an executor.
func (o *Property) CrawlProperties(mods ...qm.QueryMod) crawlPropertyQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"crawl_properties\".\"property_id\"=?", o.ID),
	)

	query := CrawlProperties(queryMods...)
	queries.SetFrom(query.Query, "\"crawl_properties\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"crawl_properties\".*"})
	}

	return query
}

// Peers retrieves all the peer's Peers with an executor.
func (o *Property) Peers(mods ...qm.QueryMod) peerQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.InnerJoin("\"peers_properties\" on \"peers\".\"id\" = \"peers_properties\".\"peer_id\""),
		qm.Where("\"peers_properties\".\"property_id\"=?", o.ID),
	)

	query := Peers(queryMods...)
	queries.SetFrom(query.Query, "\"peers\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"peers\".*"})
	}

	return query
}

// LoadCrawlProperties allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (propertyL) LoadCrawlProperties(ctx context.Context, e boil.ContextExecutor, singular bool, maybeProperty interface{}, mods queries.Applicator) error {
	var slice []*Property
	var object *Property

	if singular {
		object = maybeProperty.(*Property)
	} else {
		slice = *maybeProperty.(*[]*Property)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &propertyR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &propertyR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`crawl_properties`),
		qm.WhereIn(`crawl_properties.property_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load crawl_properties")
	}

	var resultSlice []*CrawlProperty
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice crawl_properties")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on crawl_properties")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for crawl_properties")
	}

	if len(crawlPropertyAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.CrawlProperties = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &crawlPropertyR{}
			}
			foreign.R.Property = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.PropertyID {
				local.R.CrawlProperties = append(local.R.CrawlProperties, foreign)
				if foreign.R == nil {
					foreign.R = &crawlPropertyR{}
				}
				foreign.R.Property = local
				break
			}
		}
	}

	return nil
}

// LoadPeers allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (propertyL) LoadPeers(ctx context.Context, e boil.ContextExecutor, singular bool, maybeProperty interface{}, mods queries.Applicator) error {
	var slice []*Property
	var object *Property

	if singular {
		object = maybeProperty.(*Property)
	} else {
		slice = *maybeProperty.(*[]*Property)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &propertyR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &propertyR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.Select("\"peers\".multi_hash, \"peers\".updated_at, \"peers\".created_at, \"peers\".id, \"a\".\"property_id\""),
		qm.From("\"peers\""),
		qm.InnerJoin("\"peers_properties\" as \"a\" on \"peers\".\"id\" = \"a\".\"peer_id\""),
		qm.WhereIn("\"a\".\"property_id\" in ?", args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load peers")
	}

	var resultSlice []*Peer

	var localJoinCols []int
	for results.Next() {
		one := new(Peer)
		var localJoinCol int

		err = results.Scan(&one.MultiHash, &one.UpdatedAt, &one.CreatedAt, &one.ID, &localJoinCol)
		if err != nil {
			return errors.Wrap(err, "failed to scan eager loaded results for peers")
		}
		if err = results.Err(); err != nil {
			return errors.Wrap(err, "failed to plebian-bind eager loaded slice peers")
		}

		resultSlice = append(resultSlice, one)
		localJoinCols = append(localJoinCols, localJoinCol)
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on peers")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for peers")
	}

	if len(peerAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.Peers = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &peerR{}
			}
			foreign.R.Properties = append(foreign.R.Properties, object)
		}
		return nil
	}

	for i, foreign := range resultSlice {
		localJoinCol := localJoinCols[i]
		for _, local := range slice {
			if local.ID == localJoinCol {
				local.R.Peers = append(local.R.Peers, foreign)
				if foreign.R == nil {
					foreign.R = &peerR{}
				}
				foreign.R.Properties = append(foreign.R.Properties, local)
				break
			}
		}
	}

	return nil
}

// AddCrawlProperties adds the given related objects to the existing relationships
// of the property, optionally inserting them as new records.
// Appends related to o.R.CrawlProperties.
// Sets related.R.Property appropriately.
func (o *Property) AddCrawlProperties(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*CrawlProperty) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.PropertyID = o.ID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"crawl_properties\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"property_id"}),
				strmangle.WhereClause("\"", "\"", 2, crawlPropertyPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.PropertyID = o.ID
		}
	}

	if o.R == nil {
		o.R = &propertyR{
			CrawlProperties: related,
		}
	} else {
		o.R.CrawlProperties = append(o.R.CrawlProperties, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &crawlPropertyR{
				Property: o,
			}
		} else {
			rel.R.Property = o
		}
	}
	return nil
}

// AddPeers adds the given related objects to the existing relationships
// of the property, optionally inserting them as new records.
// Appends related to o.R.Peers.
// Sets related.R.Properties appropriately.
func (o *Property) AddPeers(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Peer) error {
	var err error
	for _, rel := range related {
		if insert {
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		}
	}

	for _, rel := range related {
		query := "insert into \"peers_properties\" (\"property_id\", \"peer_id\") values ($1, $2)"
		values := []interface{}{o.ID, rel.ID}

		if boil.IsDebug(ctx) {
			writer := boil.DebugWriterFrom(ctx)
			fmt.Fprintln(writer, query)
			fmt.Fprintln(writer, values)
		}
		_, err = exec.ExecContext(ctx, query, values...)
		if err != nil {
			return errors.Wrap(err, "failed to insert into join table")
		}
	}
	if o.R == nil {
		o.R = &propertyR{
			Peers: related,
		}
	} else {
		o.R.Peers = append(o.R.Peers, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &peerR{
				Properties: PropertySlice{o},
			}
		} else {
			rel.R.Properties = append(rel.R.Properties, o)
		}
	}
	return nil
}

// SetPeers removes all previously related items of the
// property replacing them completely with the passed
// in related items, optionally inserting them as new records.
// Sets o.R.Properties's Peers accordingly.
// Replaces o.R.Peers with related.
// Sets related.R.Properties's Peers accordingly.
func (o *Property) SetPeers(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Peer) error {
	query := "delete from \"peers_properties\" where \"property_id\" = $1"
	values := []interface{}{o.ID}
	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, query)
		fmt.Fprintln(writer, values)
	}
	_, err := exec.ExecContext(ctx, query, values...)
	if err != nil {
		return errors.Wrap(err, "failed to remove relationships before set")
	}

	removePeersFromPropertiesSlice(o, related)
	if o.R != nil {
		o.R.Peers = nil
	}
	return o.AddPeers(ctx, exec, insert, related...)
}

// RemovePeers relationships from objects passed in.
// Removes related items from R.Peers (uses pointer comparison, removal does not keep order)
// Sets related.R.Properties.
func (o *Property) RemovePeers(ctx context.Context, exec boil.ContextExecutor, related ...*Peer) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	query := fmt.Sprintf(
		"delete from \"peers_properties\" where \"property_id\" = $1 and \"peer_id\" in (%s)",
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(related), 2, 1),
	)
	values := []interface{}{o.ID}
	for _, rel := range related {
		values = append(values, rel.ID)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, query)
		fmt.Fprintln(writer, values)
	}
	_, err = exec.ExecContext(ctx, query, values...)
	if err != nil {
		return errors.Wrap(err, "failed to remove relationships before set")
	}
	removePeersFromPropertiesSlice(o, related)
	if o.R == nil {
		return nil
	}

	for _, rel := range related {
		for i, ri := range o.R.Peers {
			if rel != ri {
				continue
			}

			ln := len(o.R.Peers)
			if ln > 1 && i < ln-1 {
				o.R.Peers[i] = o.R.Peers[ln-1]
			}
			o.R.Peers = o.R.Peers[:ln-1]
			break
		}
	}

	return nil
}

func removePeersFromPropertiesSlice(o *Property, related []*Peer) {
	for _, rel := range related {
		if rel.R == nil {
			continue
		}
		for i, ri := range rel.R.Properties {
			if o.ID != ri.ID {
				continue
			}

			ln := len(rel.R.Properties)
			if ln > 1 && i < ln-1 {
				rel.R.Properties[i] = rel.R.Properties[ln-1]
			}
			rel.R.Properties = rel.R.Properties[:ln-1]
			break
		}
	}
}

// Properties retrieves all the records using an executor.
func Properties(mods ...qm.QueryMod) propertyQuery {
	mods = append(mods, qm.From("\"properties\""))
	return propertyQuery{NewQuery(mods...)}
}

// FindProperty retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindProperty(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Property, error) {
	propertyObj := &Property{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"properties\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, propertyObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from properties")
	}

	if err = propertyObj.doAfterSelectHooks(ctx, exec); err != nil {
		return propertyObj, err
	}

	return propertyObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Property) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no properties provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(propertyColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	propertyInsertCacheMut.RLock()
	cache, cached := propertyInsertCache[key]
	propertyInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			propertyAllColumns,
			propertyColumnsWithDefault,
			propertyColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(propertyType, propertyMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(propertyType, propertyMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"properties\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"properties\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into properties")
	}

	if !cached {
		propertyInsertCacheMut.Lock()
		propertyInsertCache[key] = cache
		propertyInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Property.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Property) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	propertyUpdateCacheMut.RLock()
	cache, cached := propertyUpdateCache[key]
	propertyUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			propertyAllColumns,
			propertyPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update properties, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"properties\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, propertyPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(propertyType, propertyMapping, append(wl, propertyPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update properties row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for properties")
	}

	if !cached {
		propertyUpdateCacheMut.Lock()
		propertyUpdateCache[key] = cache
		propertyUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q propertyQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for properties")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for properties")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PropertySlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), propertyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"properties\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, propertyPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in property slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all property")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Property) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no properties provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(propertyColumnsWithDefault, o)

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

	propertyUpsertCacheMut.RLock()
	cache, cached := propertyUpsertCache[key]
	propertyUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			propertyAllColumns,
			propertyColumnsWithDefault,
			propertyColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			propertyAllColumns,
			propertyPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert properties, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(propertyPrimaryKeyColumns))
			copy(conflict, propertyPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"properties\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(propertyType, propertyMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(propertyType, propertyMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert properties")
	}

	if !cached {
		propertyUpsertCacheMut.Lock()
		propertyUpsertCache[key] = cache
		propertyUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Property record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Property) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Property provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), propertyPrimaryKeyMapping)
	sql := "DELETE FROM \"properties\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from properties")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for properties")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q propertyQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no propertyQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from properties")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for properties")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PropertySlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(propertyBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), propertyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"properties\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, propertyPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from property slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for properties")
	}

	if len(propertyAfterDeleteHooks) != 0 {
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
func (o *Property) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindProperty(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PropertySlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PropertySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), propertyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"properties\".* FROM \"properties\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, propertyPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in PropertySlice")
	}

	*o = slice

	return nil
}

// PropertyExists checks if the Property row exists.
func PropertyExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"properties\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if properties exists")
	}

	return exists, nil
}
