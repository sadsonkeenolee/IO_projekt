package utils

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
)

// createDriver creates driver that allows for connection to the database and
// make the migrations.
func createDriver(migrationsPath string, ci *services.Connection) (*migrate.Migrate, error) {
	cfg := services.ParseDriverConfig(ci)
	url := fmt.Sprintf("%v?multiStatements=true", cfg.FormatDSN())
	db, err := sql.Open(ci.Type, url)
	if err != nil {
		return nil, fmt.Errorf("couldn't open database connection, reason: %v\n", err)
	}
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("couldn't create a driver, reason: %v\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		ci.Type,
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't create a db instance, reason: %v\n", err)
	}
	return m, nil
}

func MigrateWipe(migrationsPath string, ci *services.Connection) error {
	m, err := createDriver(migrationsPath, ci)
	if err != nil {
		return err
	}
	if err = m.Drop(); err != nil {
		return err
	}
	return nil
}

// MigrateDatabase allows database migrating.
// If isUp is true, then the state of the database is up by version variable.
// If isUp is false, then the state of the database is down by version variable.
// If shouldForce is true, then the migration will be forced.
func MigrateDatabase(version int, ci *services.Connection,
	migrationsPath string, isUp *bool, shouldForce *bool) error {
	m, err := createDriver(migrationsPath, ci)
	if err != nil {
		return err
	}

	// sets to the given version
	if *shouldForce {
		if version <= 0 {
			return fmt.Errorf("%v is not a valid migration value for the force option (must be >=1)", version)
		}
		return m.Force(version)
	}

	// migrate up to the highest or down to the lowest migration
	if version == 0 {
		if !*isUp {
			return m.Down()
		}
		return m.Up()
	}

	if !*isUp {
		version = -version
	}

	// increase or decrease by version value
	return m.Steps(version)
}
