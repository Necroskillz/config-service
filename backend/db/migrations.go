package db

import (
	"embed"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
)

//go:embed migrations/*.sql
var fs embed.FS

func Migrate(connectionString string) error {
	u, _ := url.Parse(connectionString)
	db := dbmate.New(u)
	db.FS = fs
	db.MigrationsDir = []string{"migrations"}
	db.AutoDumpSchema = true // TODO: Change for production

	return db.CreateAndMigrate()
}
