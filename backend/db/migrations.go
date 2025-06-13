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
	u, err := url.Parse(connectionString)
	if err != nil {
		return err
	}

	db := dbmate.New(u)
	db.FS = fs
	db.MigrationsDir = []string{"migrations"}
	db.AutoDumpSchema = false

	return db.CreateAndMigrate()
}

func Drop(connectionString string) error {
	u, err := url.Parse(connectionString)
	if err != nil {
		return err
	}

	db := dbmate.New(u)

	return db.Drop()
}
