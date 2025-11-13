package utils

import (
	"database/sql"
	"fmt"
	"log"

	gsdmysql "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
)

// MigrateToVersion allows database migrating.
// If isUp is true, then the state of a database will be changed to the given version.
// If isUp is false, then the database state will be undone by the number of
// versions.
func MigrateToVersion(version int, ci *services.ConnInfo,
	migrationsPath *string, isUp *bool) error {
	cfg := gsdmysql.NewConfig()
	cfg.User = ci.Username
	cfg.Passwd = ci.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", ci.Ip, ci.Port)
	cfg.DBName = ci.Name

	db, err := sql.Open(ci.Type, cfg.FormatDSN())
	if err != nil {
		log.Fatalf("Couldn't open database connection: %v", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("Couldn't create a driver: %v.\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		*migrationsPath,
		ci.Type,
		driver,
	)

	if err != nil {
		log.Fatalf("Couldn't migrate: %v.\n", err)
	}

	if !*isUp {
		version = -version
	}

	m.Steps(version)
	return nil
}
