// Code generated by SQLBoiler 4.6.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/volatiletech/randomize"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/strmangle"
)

var (
	// Relationships sometimes use the reflection helper queries.Equal/queries.Assign
	// so force a package dependency in case they don't.
	_ = queries.Equal
)

func testConnections(t *testing.T) {
	t.Parallel()

	query := Connections()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testConnectionsDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := o.Delete(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testConnectionsQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := Connections().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testConnectionsSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := ConnectionSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testConnectionsExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := ConnectionExists(ctx, tx, o.ID)
	if err != nil {
		t.Errorf("Unable to check if Connection exists: %s", err)
	}
	if !e {
		t.Errorf("Expected ConnectionExists to return true, but got false.")
	}
}

func testConnectionsFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	connectionFound, err := FindConnection(ctx, tx, o.ID)
	if err != nil {
		t.Error(err)
	}

	if connectionFound == nil {
		t.Error("want a record, got nil")
	}
}

func testConnectionsBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = Connections().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testConnectionsOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := Connections().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testConnectionsAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	connectionOne := &Connection{}
	connectionTwo := &Connection{}
	if err = randomize.Struct(seed, connectionOne, connectionDBTypes, false, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}
	if err = randomize.Struct(seed, connectionTwo, connectionDBTypes, false, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = connectionOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = connectionTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Connections().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testConnectionsCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	connectionOne := &Connection{}
	connectionTwo := &Connection{}
	if err = randomize.Struct(seed, connectionOne, connectionDBTypes, false, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}
	if err = randomize.Struct(seed, connectionTwo, connectionDBTypes, false, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = connectionOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = connectionTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func connectionBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func connectionAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func connectionAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func connectionBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func connectionAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func connectionBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func connectionAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func connectionBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func connectionAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Connection) error {
	*o = Connection{}
	return nil
}

func testConnectionsHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &Connection{}
	o := &Connection{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, connectionDBTypes, false); err != nil {
		t.Errorf("Unable to randomize Connection object: %s", err)
	}

	AddConnectionHook(boil.BeforeInsertHook, connectionBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	connectionBeforeInsertHooks = []ConnectionHook{}

	AddConnectionHook(boil.AfterInsertHook, connectionAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	connectionAfterInsertHooks = []ConnectionHook{}

	AddConnectionHook(boil.AfterSelectHook, connectionAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	connectionAfterSelectHooks = []ConnectionHook{}

	AddConnectionHook(boil.BeforeUpdateHook, connectionBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	connectionBeforeUpdateHooks = []ConnectionHook{}

	AddConnectionHook(boil.AfterUpdateHook, connectionAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	connectionAfterUpdateHooks = []ConnectionHook{}

	AddConnectionHook(boil.BeforeDeleteHook, connectionBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	connectionBeforeDeleteHooks = []ConnectionHook{}

	AddConnectionHook(boil.AfterDeleteHook, connectionAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	connectionAfterDeleteHooks = []ConnectionHook{}

	AddConnectionHook(boil.BeforeUpsertHook, connectionBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	connectionBeforeUpsertHooks = []ConnectionHook{}

	AddConnectionHook(boil.AfterUpsertHook, connectionAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	connectionAfterUpsertHooks = []ConnectionHook{}
}

func testConnectionsInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testConnectionsInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(connectionColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testConnectionsReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = o.Reload(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testConnectionsReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := ConnectionSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testConnectionsSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Connections().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	connectionDBTypes = map[string]string{`ID`: `integer`, `PeerID`: `character varying`, `MultiAddress`: `ARRAYcharacter varying`, `AgentVersion`: `character varying`, `DialAttempt`: `timestamp with time zone`, `Latency`: `interval`, `IsSucceed`: `boolean`}
	_                 = bytes.MinRead
)

func testConnectionsUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(connectionPrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(connectionAllColumns) == len(connectionPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testConnectionsSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(connectionAllColumns) == len(connectionPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Connection{}
	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, connectionDBTypes, true, connectionPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(connectionAllColumns, connectionPrimaryKeyColumns) {
		fields = connectionAllColumns
	} else {
		fields = strmangle.SetComplement(
			connectionAllColumns,
			connectionPrimaryKeyColumns,
		)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	typ := reflect.TypeOf(o).Elem()
	n := typ.NumField()

	updateMap := M{}
	for _, col := range fields {
		for i := 0; i < n; i++ {
			f := typ.Field(i)
			if f.Tag.Get("boil") == col {
				updateMap[col] = value.Field(i).Interface()
			}
		}
	}

	slice := ConnectionSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testConnectionsUpsert(t *testing.T) {
	t.Parallel()

	if len(connectionAllColumns) == len(connectionPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := Connection{}
	if err = randomize.Struct(seed, &o, connectionDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Connection: %s", err)
	}

	count, err := Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, connectionDBTypes, false, connectionPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Connection struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Connection: %s", err)
	}

	count, err = Connections().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
