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

func testCrawls(t *testing.T) {
	t.Parallel()

	query := Crawls()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testCrawlsDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
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

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testCrawlsQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := Crawls().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testCrawlsSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := CrawlSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testCrawlsExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := CrawlExists(ctx, tx, o.ID)
	if err != nil {
		t.Errorf("Unable to check if Crawl exists: %s", err)
	}
	if !e {
		t.Errorf("Expected CrawlExists to return true, but got false.")
	}
}

func testCrawlsFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	crawlFound, err := FindCrawl(ctx, tx, o.ID)
	if err != nil {
		t.Error(err)
	}

	if crawlFound == nil {
		t.Error("want a record, got nil")
	}
}

func testCrawlsBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = Crawls().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testCrawlsOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := Crawls().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testCrawlsAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	crawlOne := &Crawl{}
	crawlTwo := &Crawl{}
	if err = randomize.Struct(seed, crawlOne, crawlDBTypes, false, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}
	if err = randomize.Struct(seed, crawlTwo, crawlDBTypes, false, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = crawlOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = crawlTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Crawls().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testCrawlsCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	crawlOne := &Crawl{}
	crawlTwo := &Crawl{}
	if err = randomize.Struct(seed, crawlOne, crawlDBTypes, false, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}
	if err = randomize.Struct(seed, crawlTwo, crawlDBTypes, false, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = crawlOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = crawlTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func crawlBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func crawlAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func crawlAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func crawlBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func crawlAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func crawlBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func crawlAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func crawlBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func crawlAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Crawl) error {
	*o = Crawl{}
	return nil
}

func testCrawlsHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &Crawl{}
	o := &Crawl{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, crawlDBTypes, false); err != nil {
		t.Errorf("Unable to randomize Crawl object: %s", err)
	}

	AddCrawlHook(boil.BeforeInsertHook, crawlBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	crawlBeforeInsertHooks = []CrawlHook{}

	AddCrawlHook(boil.AfterInsertHook, crawlAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	crawlAfterInsertHooks = []CrawlHook{}

	AddCrawlHook(boil.AfterSelectHook, crawlAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	crawlAfterSelectHooks = []CrawlHook{}

	AddCrawlHook(boil.BeforeUpdateHook, crawlBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	crawlBeforeUpdateHooks = []CrawlHook{}

	AddCrawlHook(boil.AfterUpdateHook, crawlAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	crawlAfterUpdateHooks = []CrawlHook{}

	AddCrawlHook(boil.BeforeDeleteHook, crawlBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	crawlBeforeDeleteHooks = []CrawlHook{}

	AddCrawlHook(boil.AfterDeleteHook, crawlAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	crawlAfterDeleteHooks = []CrawlHook{}

	AddCrawlHook(boil.BeforeUpsertHook, crawlBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	crawlBeforeUpsertHooks = []CrawlHook{}

	AddCrawlHook(boil.AfterUpsertHook, crawlAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	crawlAfterUpsertHooks = []CrawlHook{}
}

func testCrawlsInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testCrawlsInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(crawlColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testCrawlToManyPeerProperties(t *testing.T) {
	var err error
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a Crawl
	var b, c PeerProperty

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	if err = randomize.Struct(seed, &b, peerPropertyDBTypes, false, peerPropertyColumnsWithDefault...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, peerPropertyDBTypes, false, peerPropertyColumnsWithDefault...); err != nil {
		t.Fatal(err)
	}

	b.CrawlID = a.ID
	c.CrawlID = a.ID

	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = c.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	check, err := a.PeerProperties().All(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	bFound, cFound := false, false
	for _, v := range check {
		if v.CrawlID == b.CrawlID {
			bFound = true
		}
		if v.CrawlID == c.CrawlID {
			cFound = true
		}
	}

	if !bFound {
		t.Error("expected to find b")
	}
	if !cFound {
		t.Error("expected to find c")
	}

	slice := CrawlSlice{&a}
	if err = a.L.LoadPeerProperties(ctx, tx, false, (*[]*Crawl)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if got := len(a.R.PeerProperties); got != 2 {
		t.Error("number of eager loaded records wrong, got:", got)
	}

	a.R.PeerProperties = nil
	if err = a.L.LoadPeerProperties(ctx, tx, true, &a, nil); err != nil {
		t.Fatal(err)
	}
	if got := len(a.R.PeerProperties); got != 2 {
		t.Error("number of eager loaded records wrong, got:", got)
	}

	if t.Failed() {
		t.Logf("%#v", check)
	}
}

func testCrawlToManyAddOpPeerProperties(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a Crawl
	var b, c, d, e PeerProperty

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, crawlDBTypes, false, strmangle.SetComplement(crawlPrimaryKeyColumns, crawlColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	foreigners := []*PeerProperty{&b, &c, &d, &e}
	for _, x := range foreigners {
		if err = randomize.Struct(seed, x, peerPropertyDBTypes, false, strmangle.SetComplement(peerPropertyPrimaryKeyColumns, peerPropertyColumnsWithoutDefault)...); err != nil {
			t.Fatal(err)
		}
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = c.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	foreignersSplitByInsertion := [][]*PeerProperty{
		{&b, &c},
		{&d, &e},
	}

	for i, x := range foreignersSplitByInsertion {
		err = a.AddPeerProperties(ctx, tx, i != 0, x...)
		if err != nil {
			t.Fatal(err)
		}

		first := x[0]
		second := x[1]

		if a.ID != first.CrawlID {
			t.Error("foreign key was wrong value", a.ID, first.CrawlID)
		}
		if a.ID != second.CrawlID {
			t.Error("foreign key was wrong value", a.ID, second.CrawlID)
		}

		if first.R.Crawl != &a {
			t.Error("relationship was not added properly to the foreign slice")
		}
		if second.R.Crawl != &a {
			t.Error("relationship was not added properly to the foreign slice")
		}

		if a.R.PeerProperties[i*2] != first {
			t.Error("relationship struct slice not set to correct value")
		}
		if a.R.PeerProperties[i*2+1] != second {
			t.Error("relationship struct slice not set to correct value")
		}

		count, err := a.PeerProperties().Count(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		if want := int64((i + 1) * 2); count != want {
			t.Error("want", want, "got", count)
		}
	}
}

func testCrawlsReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
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

func testCrawlsReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := CrawlSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testCrawlsSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Crawls().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	crawlDBTypes = map[string]string{`ID`: `integer`, `State`: `enum.crawl_state('started','cancelled','failed','succeeded')`, `StartedAt`: `timestamp with time zone`, `FinishedAt`: `timestamp with time zone`, `UpdatedAt`: `timestamp with time zone`, `CreatedAt`: `timestamp with time zone`, `CrawledPeers`: `integer`, `DialablePeers`: `integer`, `UndialablePeers`: `integer`}
	_            = bytes.MinRead
)

func testCrawlsUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(crawlPrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(crawlAllColumns) == len(crawlPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testCrawlsSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(crawlAllColumns) == len(crawlPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Crawl{}
	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, crawlDBTypes, true, crawlPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(crawlAllColumns, crawlPrimaryKeyColumns) {
		fields = crawlAllColumns
	} else {
		fields = strmangle.SetComplement(
			crawlAllColumns,
			crawlPrimaryKeyColumns,
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

	slice := CrawlSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testCrawlsUpsert(t *testing.T) {
	t.Parallel()

	if len(crawlAllColumns) == len(crawlPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := Crawl{}
	if err = randomize.Struct(seed, &o, crawlDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Crawl: %s", err)
	}

	count, err := Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, crawlDBTypes, false, crawlPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Crawl struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Crawl: %s", err)
	}

	count, err = Crawls().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
