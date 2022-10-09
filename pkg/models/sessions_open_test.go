// Code generated by SQLBoiler 4.13.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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

func testSessionsOpens(t *testing.T) {
	t.Parallel()

	query := SessionsOpens()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testSessionsOpensDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
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

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testSessionsOpensQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := SessionsOpens().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testSessionsOpensSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := SessionsOpenSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testSessionsOpensExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := SessionsOpenExists(ctx, tx, o.ID, o.State, o.CreatedAt)
	if err != nil {
		t.Errorf("Unable to check if SessionsOpen exists: %s", err)
	}
	if !e {
		t.Errorf("Expected SessionsOpenExists to return true, but got false.")
	}
}

func testSessionsOpensFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	sessionsOpenFound, err := FindSessionsOpen(ctx, tx, o.ID, o.State, o.CreatedAt)
	if err != nil {
		t.Error(err)
	}

	if sessionsOpenFound == nil {
		t.Error("want a record, got nil")
	}
}

func testSessionsOpensBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = SessionsOpens().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testSessionsOpensOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := SessionsOpens().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testSessionsOpensAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	sessionsOpenOne := &SessionsOpen{}
	sessionsOpenTwo := &SessionsOpen{}
	if err = randomize.Struct(seed, sessionsOpenOne, sessionsOpenDBTypes, false, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}
	if err = randomize.Struct(seed, sessionsOpenTwo, sessionsOpenDBTypes, false, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = sessionsOpenOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = sessionsOpenTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := SessionsOpens().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testSessionsOpensCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	sessionsOpenOne := &SessionsOpen{}
	sessionsOpenTwo := &SessionsOpen{}
	if err = randomize.Struct(seed, sessionsOpenOne, sessionsOpenDBTypes, false, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}
	if err = randomize.Struct(seed, sessionsOpenTwo, sessionsOpenDBTypes, false, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = sessionsOpenOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = sessionsOpenTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func sessionsOpenBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func sessionsOpenAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func sessionsOpenAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func sessionsOpenBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func sessionsOpenAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func sessionsOpenBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func sessionsOpenAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func sessionsOpenBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func sessionsOpenAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *SessionsOpen) error {
	*o = SessionsOpen{}
	return nil
}

func testSessionsOpensHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &SessionsOpen{}
	o := &SessionsOpen{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, false); err != nil {
		t.Errorf("Unable to randomize SessionsOpen object: %s", err)
	}

	AddSessionsOpenHook(boil.BeforeInsertHook, sessionsOpenBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	sessionsOpenBeforeInsertHooks = []SessionsOpenHook{}

	AddSessionsOpenHook(boil.AfterInsertHook, sessionsOpenAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	sessionsOpenAfterInsertHooks = []SessionsOpenHook{}

	AddSessionsOpenHook(boil.AfterSelectHook, sessionsOpenAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	sessionsOpenAfterSelectHooks = []SessionsOpenHook{}

	AddSessionsOpenHook(boil.BeforeUpdateHook, sessionsOpenBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	sessionsOpenBeforeUpdateHooks = []SessionsOpenHook{}

	AddSessionsOpenHook(boil.AfterUpdateHook, sessionsOpenAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	sessionsOpenAfterUpdateHooks = []SessionsOpenHook{}

	AddSessionsOpenHook(boil.BeforeDeleteHook, sessionsOpenBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	sessionsOpenBeforeDeleteHooks = []SessionsOpenHook{}

	AddSessionsOpenHook(boil.AfterDeleteHook, sessionsOpenAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	sessionsOpenAfterDeleteHooks = []SessionsOpenHook{}

	AddSessionsOpenHook(boil.BeforeUpsertHook, sessionsOpenBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	sessionsOpenBeforeUpsertHooks = []SessionsOpenHook{}

	AddSessionsOpenHook(boil.AfterUpsertHook, sessionsOpenAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	sessionsOpenAfterUpsertHooks = []SessionsOpenHook{}
}

func testSessionsOpensInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testSessionsOpensInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(sessionsOpenColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testSessionsOpenToOnePeerUsingPeer(t *testing.T) {
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var local SessionsOpen
	var foreign Peer

	seed := randomize.NewSeed()
	if err := randomize.Struct(seed, &local, sessionsOpenDBTypes, false, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}
	if err := randomize.Struct(seed, &foreign, peerDBTypes, false, peerColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Peer struct: %s", err)
	}

	if err := foreign.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	local.PeerID = foreign.ID
	if err := local.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	check, err := local.Peer().One(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	if check.ID != foreign.ID {
		t.Errorf("want: %v, got %v", foreign.ID, check.ID)
	}

	slice := SessionsOpenSlice{&local}
	if err = local.L.LoadPeer(ctx, tx, false, (*[]*SessionsOpen)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if local.R.Peer == nil {
		t.Error("struct should have been eager loaded")
	}

	local.R.Peer = nil
	if err = local.L.LoadPeer(ctx, tx, true, &local, nil); err != nil {
		t.Fatal(err)
	}
	if local.R.Peer == nil {
		t.Error("struct should have been eager loaded")
	}
}

func testSessionsOpenToOneSetOpPeerUsingPeer(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a SessionsOpen
	var b, c Peer

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, sessionsOpenDBTypes, false, strmangle.SetComplement(sessionsOpenPrimaryKeyColumns, sessionsOpenColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &b, peerDBTypes, false, strmangle.SetComplement(peerPrimaryKeyColumns, peerColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, peerDBTypes, false, strmangle.SetComplement(peerPrimaryKeyColumns, peerColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	for i, x := range []*Peer{&b, &c} {
		err = a.SetPeer(ctx, tx, i != 0, x)
		if err != nil {
			t.Fatal(err)
		}

		if a.R.Peer != x {
			t.Error("relationship struct not set to correct value")
		}

		if x.R.SessionsOpen != &a {
			t.Error("failed to append to foreign relationship struct")
		}
		if a.PeerID != x.ID {
			t.Error("foreign key was wrong value", a.PeerID)
		}

		zero := reflect.Zero(reflect.TypeOf(a.PeerID))
		reflect.Indirect(reflect.ValueOf(&a.PeerID)).Set(zero)

		if err = a.Reload(ctx, tx); err != nil {
			t.Fatal("failed to reload", err)
		}

		if a.PeerID != x.ID {
			t.Error("foreign key was wrong value", a.PeerID, x.ID)
		}
	}
}

func testSessionsOpensReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
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

func testSessionsOpensReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := SessionsOpenSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testSessionsOpensSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := SessionsOpens().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	sessionsOpenDBTypes = map[string]string{`ID`: `integer`, `PeerID`: `integer`, `FirstSuccessfulVisit`: `timestamp with time zone`, `LastSuccessfulVisit`: `timestamp with time zone`, `NextVisitDueAt`: `timestamp with time zone`, `FirstFailedVisit`: `timestamp with time zone`, `LastFailedVisit`: `timestamp with time zone`, `UpdatedAt`: `timestamp with time zone`, `CreatedAt`: `timestamp with time zone`, `MinDuration`: `interval`, `MaxDuration`: `interval`, `SuccessfulVisitsCount`: `integer`, `State`: `enum.session_state('open','pending','closed')`, `FailedVisitsCount`: `smallint`, `RecoveredCount`: `integer`, `FinishReason`: `enum.dial_error('unknown','io_timeout','connection_refused','protocol_not_supported','peer_id_mismatch','no_route_to_host','network_unreachable','no_good_addresses','context_deadline_exceeded','no_public_ip','max_dial_attempts_exceeded','maddr_reset','stream_reset','host_is_down','negotiate_security_protocol_no_trailing_new_line')`}
	_                   = bytes.MinRead
)

func testSessionsOpensUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(sessionsOpenPrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(sessionsOpenAllColumns) == len(sessionsOpenPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testSessionsOpensSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(sessionsOpenAllColumns) == len(sessionsOpenPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &SessionsOpen{}
	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, sessionsOpenDBTypes, true, sessionsOpenPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(sessionsOpenAllColumns, sessionsOpenPrimaryKeyColumns) {
		fields = sessionsOpenAllColumns
	} else {
		fields = strmangle.SetComplement(
			sessionsOpenAllColumns,
			sessionsOpenPrimaryKeyColumns,
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

	slice := SessionsOpenSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testSessionsOpensUpsert(t *testing.T) {
	t.Parallel()

	if len(sessionsOpenAllColumns) == len(sessionsOpenPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := SessionsOpen{}
	if err = randomize.Struct(seed, &o, sessionsOpenDBTypes, true); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert SessionsOpen: %s", err)
	}

	count, err := SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, sessionsOpenDBTypes, false, sessionsOpenPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize SessionsOpen struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert SessionsOpen: %s", err)
	}

	count, err = SessionsOpens().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
