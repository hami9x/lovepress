package db

import (
	"database/sql"
	"fmt"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"

	"github.com/phaikawl/lovepress/model"
)

const (
	DbType = "postgres"
)

var (
	gService DbService
)

type DbService struct {
	*gorp.DbMap
}

func Service() *DbService {
	if gService.DbMap == nil {
		panic("Database service uninitialized.")
	}
	return &gService
}

type Config struct {
	Database string
	User     string
	Password string
	Host     string
	Port     int
}

func loadDb(conf Config) *sql.DB {
	db, err := sql.Open(DbType,
		fmt.Sprintf("%v://%v:%v@%v:%v/%v",
			DbType,
			conf.User,
			conf.Password,
			conf.Host,
			conf.Port,
			conf.Database,
		))
	if err != nil {
		panic("Cannot open database: " + err.Error())
	}

	return db
}

func Init(conf Config) {
	constructDb(loadDb(conf))
}

func constructDb(db *sql.DB) {
	// construct a gorp DbMap
	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	dbMap.AddTableWithName(model.Post{}, "posts").SetKeys(true, "Id")
	dbMap.AddTableWithName(model.User{}, "users").SetKeys(true, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err := dbMap.CreateTablesIfNotExists()
	if err != nil {
		panic("Failed creating tables. " + err.Error())
	}

	gService = DbService{dbMap}
}
