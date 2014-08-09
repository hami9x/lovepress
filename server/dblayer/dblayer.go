package dblayer

import "github.com/phaikawl/lovepress/server/services/db"

var (
	gDb *db.DbService
)

func Init() {
	gDb = db.Service()
}
