package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/volatiletech/sqlboiler/v4/boil"

	_ "github.com/lib/pq"
	//. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func TestAnything(t *testing.T) {
	// Open handle to database like normal
	db, err := sql.Open("postgres", "dbname=nebula user=nebula password=password sslmode=disable")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	boil.SetDB(db)

	crawl := &models.Crawl{
		StartedAt:       time.Now(),
		CrawledPeers:    1,
		DialablePeers:   2,
		UndialablePeers: 3,
	}

	err = crawl.Insert(ctx, db, boil.Infer())
	if err != nil {
		panic(err)
	}

	pp := &models.PeerProperty{
		Property: "AgentVersion",
		Count:    122,
	}
	err = pp.SetCrawl(ctx, db, false, crawl)
	if err != nil {
		panic(err)
	}
	err = pp.Insert(ctx, db, boil.Infer())
	if err != nil {
		panic(err)
	}

	all, err := models.Crawls().All(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, c := range all {
		fmt.Println(c.ID)
	}

	//// If you don't want to pass in db to all generated methods
	//// you can use boil.SetDB to set it globally, and then use
	//// the G variant methods like so (--add-global-variants to enable)
	//boil.SetDB(db)
	//users, err := models.Users().AllG(ctx)
	//
	//// Query all users
	//users, err := models.Users().All(ctx, db)
	//
	//// Panic-able if you like to code that way (--add-panic-variants to enable)
	//users := models.Users().AllP(db)
	//
	//// More complex query
	//users, err := models.Users(Where("age > ?", 30), Limit(5), Offset(6)).All(ctx, db)
	//
	//// Ultra complex query
	//users, err := models.Users(
	//	Select("id", "name"),
	//	InnerJoin("credit_cards c on c.user_id = users.id"),
	//	Where("age > ?", 30),
	//	AndIn("c.kind in ?", "visa", "mastercard"),
	//	Or("email like ?", `%aol.com%`),
	//	GroupBy("id", "name"),
	//	Having("count(c.id) > ?", 2),
	//	Limit(5),
	//	Offset(6),
	//).All(ctx, db)
	//
	//// Use any "boil.Executor" implementation (*sql.DB, *sql.Tx, data-dog mock db)
	//// for any query.
	//tx, err := db.BeginTx(ctx, nil)
	//if err != nil {
	//	return err
	//}
	//users, err := models.Users().All(ctx, tx)
	//
	//// Relationships
	//user, err := models.Users().One(ctx, db)
	//if err != nil {
	//	return err
	//}
	//movies, err := user.FavoriteMovies().All(ctx, db)
	//
	//// Eager loading
	//users, err := models.Users(Load("FavoriteMovies")).All(ctx, db)
	//if err != nil {
	//	return err
	//}
	//fmt.Println(len(users.R.FavoriteMovies))
}
