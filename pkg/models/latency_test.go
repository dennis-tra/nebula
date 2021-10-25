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

func testLatencies(t *testing.T) {
	t.Parallel()

	query := Latencies()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testLatenciesDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
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

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testLatenciesQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := Latencies().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testLatenciesSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := LatencySlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testLatenciesExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := LatencyExists(ctx, tx, o.PeerID)
	if err != nil {
		t.Errorf("Unable to check if Latency exists: %s", err)
	}
	if !e {
		t.Errorf("Expected LatencyExists to return true, but got false.")
	}
}

func testLatenciesFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	latencyFound, err := FindLatency(ctx, tx, o.PeerID)
	if err != nil {
		t.Error(err)
	}

	if latencyFound == nil {
		t.Error("want a record, got nil")
	}
}

func testLatenciesBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = Latencies().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testLatenciesOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := Latencies().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testLatenciesAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	latencyOne := &Latency{}
	latencyTwo := &Latency{}
	if err = randomize.Struct(seed, latencyOne, latencyDBTypes, false, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}
	if err = randomize.Struct(seed, latencyTwo, latencyDBTypes, false, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = latencyOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = latencyTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Latencies().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testLatenciesCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	latencyOne := &Latency{}
	latencyTwo := &Latency{}
	if err = randomize.Struct(seed, latencyOne, latencyDBTypes, false, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}
	if err = randomize.Struct(seed, latencyTwo, latencyDBTypes, false, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = latencyOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = latencyTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func latencyBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func latencyAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func latencyAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func latencyBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func latencyAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func latencyBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func latencyAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func latencyBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func latencyAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Latency) error {
	*o = Latency{}
	return nil
}

func testLatenciesHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &Latency{}
	o := &Latency{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, latencyDBTypes, false); err != nil {
		t.Errorf("Unable to randomize Latency object: %s", err)
	}

	AddLatencyHook(boil.BeforeInsertHook, latencyBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	latencyBeforeInsertHooks = []LatencyHook{}

	AddLatencyHook(boil.AfterInsertHook, latencyAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	latencyAfterInsertHooks = []LatencyHook{}

	AddLatencyHook(boil.AfterSelectHook, latencyAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	latencyAfterSelectHooks = []LatencyHook{}

	AddLatencyHook(boil.BeforeUpdateHook, latencyBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	latencyBeforeUpdateHooks = []LatencyHook{}

	AddLatencyHook(boil.AfterUpdateHook, latencyAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	latencyAfterUpdateHooks = []LatencyHook{}

	AddLatencyHook(boil.BeforeDeleteHook, latencyBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	latencyBeforeDeleteHooks = []LatencyHook{}

	AddLatencyHook(boil.AfterDeleteHook, latencyAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	latencyAfterDeleteHooks = []LatencyHook{}

	AddLatencyHook(boil.BeforeUpsertHook, latencyBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	latencyBeforeUpsertHooks = []LatencyHook{}

	AddLatencyHook(boil.AfterUpsertHook, latencyAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	latencyAfterUpsertHooks = []LatencyHook{}
}

func testLatenciesInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testLatenciesInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(latencyColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testLatenciesReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
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

func testLatenciesReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := LatencySlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testLatenciesSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Latencies().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	latencyDBTypes = map[string]string{`PeerID`: `character varying`, `DialAttempts`: `integer`, `AvgLatency`: `interval`}
	_              = bytes.MinRead
)

func testLatenciesUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(latencyPrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(latencyAllColumns) == len(latencyPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testLatenciesSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(latencyAllColumns) == len(latencyPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Latency{}
	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, latencyDBTypes, true, latencyPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(latencyAllColumns, latencyPrimaryKeyColumns) {
		fields = latencyAllColumns
	} else {
		fields = strmangle.SetComplement(
			latencyAllColumns,
			latencyPrimaryKeyColumns,
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

	slice := LatencySlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testLatenciesUpsert(t *testing.T) {
	t.Parallel()

	if len(latencyAllColumns) == len(latencyPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := Latency{}
	if err = randomize.Struct(seed, &o, latencyDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Latency: %s", err)
	}

	count, err := Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, latencyDBTypes, false, latencyPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Latency struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Latency: %s", err)
	}

	count, err = Latencies().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
