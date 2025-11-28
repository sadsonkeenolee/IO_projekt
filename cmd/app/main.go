package main

import (
	"flag"
	"log"
	"path/filepath"

	"os"

	"github.com/sadsonkeenolee/IO_projekt/internal/services/credentials"
	_ "github.com/sadsonkeenolee/IO_projekt/internal/services/credentials"
	"github.com/sadsonkeenolee/IO_projekt/internal/services/etl"
	"github.com/sadsonkeenolee/IO_projekt/internal/services/fetch"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/sadsonkeenolee/IO_projekt/pkg/utils"
)

// Define global flags
var migrateFlag = flag.Bool("migrate", false, "run a migration in the current app")
var forceFlag = flag.Bool("force", false, "will force any operation you do")
var upFlag = flag.Bool("up", false, "run an up migration, otherwise down")
var dropFlag = flag.Bool("drop", false, "wipe out a database, use only if the database is bugged")
var versionFlag = flag.Int("version", 0, "specify the number of the version")
var serviceFlag = flag.String("service", "credentials", "specify which service to run")

const (
	Credentials = iota
	Etl
	Fetch
)

var serviceMap = map[string]uint{
	"credentials": Credentials,
	"etl":         Etl,
	"fetch":       Fetch,
}

func main() {
	flag.Parse()
	l := log.New(os.Stdout, "main: ", log.LstdFlags)

	// Define all the necessary environment variables.
	// Global variables here vvv
	configsPath, err := filepath.Abs("api/configs")
	if err != nil {
		l.Fatalf("The function failed: %v", err)
	}

	downloadedDataPath, err := filepath.Abs("temp/")
	if err != nil {
		l.Fatalf("The function failed: %v", err)
	}
	_ = os.Setenv("DOWNLOAD_DIR", downloadedDataPath)

	// Global variables here ^^^

	var migrationsPath string
	switch serviceMap[*serviceFlag] {
	case Credentials:
		migrationsPath, err = filepath.Abs("api/migrations/credentials")
	case Etl:
		migrationsPath, err = filepath.Abs("api/migrations/etl")
	case Fetch:
		migrationsPath, err = filepath.Abs("api/migrations/fetch")
	}

	if err != nil {
		l.Fatalf("The function failed: %v", err)
	}
	migrationsPath = "file://" + migrationsPath

	switch serviceMap[*serviceFlag] {
	case Credentials:
		_ = os.Setenv("CREDENTIALS_CONFIG_DIR_PATH", configsPath)
		_ = os.Setenv("CREDENTIALS_MIGRATIONS_DIR_PATH", migrationsPath)
	case Etl:
		_ = os.Setenv("ETL_CONFIG_DIR_PATH", configsPath)
		_ = os.Setenv("ETL_MIGRATIONS_DIR_PATH", migrationsPath)
	case Fetch:
		_ = os.Setenv("FETCH_CONFIG_DIR_PATH", configsPath)
		_ = os.Setenv("FETCH_MIGRATIONS_DIR_PATH", migrationsPath)
	}

	var s services.IService
	switch serviceMap[*serviceFlag] {
	case Credentials:
		s, _ = credentials.NewCredentials()
	case Etl:
		s, _ = etl.NewEtl(1000)
	case Fetch:
		s, _ = fetch.NewFetch()
	}

	if *migrateFlag {
		ci := s.ExposeConnInfo()

		if *dropFlag {
			if err := utils.MigrateWipe(&migrationsPath, ci); err != nil {
				l.Fatalf("The function failed: %v\n", err)
			}
			l.Println("Database wiped.")
			return
		}

		if err := utils.MigrateToVersion(*versionFlag, ci,
			&migrationsPath, upFlag, forceFlag); err != nil {
			l.Fatalf("The function failed: %v\n", err)
		}
		l.Println("Migration completed.")
		return
	}

	l.Printf("%v is starting...\n", s)
	err = s.Start()
	l.Printf("Server returned value: %v.\n", err)
}

// TODO:
// -normalizacja i oczyszczenie stringow (baza sie buguje)
// -przebudowanie tabeli (po insert ignore mozna miec niewlasciwe rekordy)
