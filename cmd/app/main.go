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

var MainLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix|log.Lshortfile)

// Define global flags
var migrateFlag = flag.Bool("migrate", false, "run migration")
var forceFlag = flag.Bool("force", false, "will force migration to the given version")
var upFlag = flag.Bool("up", false, "up migration, otherwise down")
var wipeFlag = flag.Bool("wipe", false, "wipe out a database")
var versionFlag = flag.Int("version", 0, "value of migration")
var serviceNameFlag = flag.String("service", "", "service name to run")

const (
	Invalid = iota
	Credentials
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
	// Define all the necessary environment variables.
	configsPath, err := filepath.Abs("api/configs")
	if err != nil {
		MainLogger.Fatalf("Error while setting up the config file: %v\n", err)
	}

	downloadedDataPath, err := filepath.Abs("temp/")
	if err != nil {
		MainLogger.Fatalf("Error while setting up the temp directory: %v\n", err)
	}
	_ = os.Setenv("DOWNLOAD_DIR", downloadedDataPath)

	// Here goes the rest of the program
	var migrationsPath string
	switch serviceMap[*serviceNameFlag] {
	case Credentials:
		migrationsPath, err = filepath.Abs("api/migrations/credentials")
	case Etl:
		migrationsPath, err = filepath.Abs("api/migrations/etl")
	case Fetch:
		migrationsPath, err = filepath.Abs("api/migrations/fetch")
	default:
		MainLogger.Fatalf("`%v` is an invalid name for a service\n", *serviceNameFlag)
	}

	if err != nil {
		MainLogger.Fatalf("Error while setting up the migration paths: %v\n", err)
	}

	migrationsPath = filepath.Join("file:/", migrationsPath)

	switch serviceMap[*serviceNameFlag] {
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
	switch serviceMap[*serviceNameFlag] {
	case Credentials:
		s = credentials.CredentialsBuilder(
			credentials.WithLogger(
				os.Stdout,
				"Credentials Service:  ",
				log.LstdFlags|log.Lmsgprefix,
			),
			credentials.WithRouter(),
			credentials.WithConfig(
				"CredentialsConfig",
				"toml",
				os.Getenv("CREDENTIALS_CONFIG_DIR_PATH"),
			),
			credentials.WithConnectionInfo("ConnInfo"),
			credentials.WithDatabase(),
		)
	case Etl:
		s = etl.EtlBuilder(
			etl.WithLogger(
				os.Stdout,
				"Etl Service:  ",
				log.LstdFlags|log.Lmsgprefix,
			),
			etl.WithRouter(),
			etl.WithConfig(
				"EtlConfig",
				"toml",
				os.Getenv("ETL_CONFIG_DIR_PATH"),
			),
			etl.WithConnectionInfo("ConnInfo"),
			etl.WithDatabase(),
			etl.WithBatchSize(64),
		)
	case Fetch:
		s = fetch.FetchBuilder(
			fetch.WithLogger(
				os.Stdout,
				"Fetch Service:  ",
				log.LstdFlags|log.Lmsgprefix,
			),
			fetch.WithRouter(),
			fetch.WithConfig(
				"FetchConfig",
				"toml",
				os.Getenv("FETCH_CONFIG_DIR_PATH"),
			),
			fetch.WithConnectionInfo("ConnInfo"),
			fetch.WithDatabase(),
		)
	}

	if *migrateFlag {
		ci := s.ExposeConnInfo()

		if *wipeFlag {
			if err := utils.MigrateWipe(&migrationsPath, ci); err != nil {
				MainLogger.Fatalf("The function failed: %v\n", err)
			}
			MainLogger.Println("Database wiped.")
			return
		}

		if err := utils.MigrateDatabase(*versionFlag, ci,
			&migrationsPath, upFlag, forceFlag); err != nil {
			MainLogger.Fatalf("The function failed: %v\n", err)
		}
		MainLogger.Println("Migration completed.")
		return
	}

	MainLogger.Printf("%v is starting...\n", s)
	err = s.Start()
	MainLogger.Printf("Server returned value: %v.\n", err)
}

// TODO:
// -normalizacja i oczyszczenie stringow (baza sie buguje)
// - utworzenie widokow w bazie danych
