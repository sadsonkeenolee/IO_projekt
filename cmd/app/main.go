package main

import (
	"flag"
	"log"
	"path/filepath"

	"os"

	"github.com/sadsonkeenolee/IO_projekt/internal/services/auth"
	"github.com/sadsonkeenolee/IO_projekt/internal/services/ingest"
	"github.com/sadsonkeenolee/IO_projekt/internal/services/search"
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
	Auth
	Ingest
	Search
)

var ServiceMap = map[string]uint{
	"auth":   Auth,
	"ingest": Ingest,
	"search": Search,
}

var EnvVariables map[string]string = map[string]string{
	"DOWNLOAD_DIR": InitPath("temp/"),
	"MIGRATIONS":   InitJoinedPath("file:/", InitPath("api/migrations")),
	"CONFIG":       InitPath("api/configs"),
}

func InitPath(subpath string) string {
	fullPath, err := filepath.Abs(subpath)
	if err != nil {
		MainLogger.Printf("Error while parsing `%v`: %v.\n", subpath, err)
		return ""
	}
	return fullPath
}

func InitJoinedPath(subpaths ...string) string {
	if len(subpaths) <= 1 {
		MainLogger.Printf("Got only `%v` paths.\n", len(subpaths))
		return ""
	}
	fullPath := filepath.Join(subpaths...)
	if fullPath == "" {
		MainLogger.Printf("Couldn't join your paths: `%v`.\n", subpaths)
		return ""
	}
	return fullPath
}

func main() {
	flag.Parse()
	if _, ok := ServiceMap[*serviceNameFlag]; !ok {
		MainLogger.Fatalf("`%v` is an incorrect service, flag.\n", *serviceNameFlag)
	}

	// Set a correct migration file.
	switch ServiceMap[*serviceNameFlag] {
	case Auth:
		EnvVariables["MIGRATIONS"] = InitJoinedPath(EnvVariables["MIGRATIONS"], "auth")
	case Ingest:
		EnvVariables["MIGRATIONS"] = InitJoinedPath(EnvVariables["MIGRATIONS"], "ingest")
	case Search:
		EnvVariables["MIGRATIONS"] = InitJoinedPath(EnvVariables["MIGRATIONS"], "search")
	}

	// Set all variables
	for k, v := range EnvVariables {
		_ = os.Setenv(k, v)
	}

	var s services.IService
	switch ServiceMap[*serviceNameFlag] {
	case Auth:
		l := services.NewLogger(
			os.Stdout,
			"Auth Service:  ",
			log.LstdFlags|log.Lmsgprefix,
		)
		ge := services.NewRouter()
		v := services.NewViper(
			"AuthConfig",
			"toml",
			os.Getenv("CONFIG"),
		)
		c := services.NewConnection("ConnInfo", v)
		db := services.NewDatabase(c)
		s = auth.AuthBuilder(
			auth.WithLogger(l),
			auth.WithRouter(ge),
			auth.WithViper(v),
			auth.WithConnectionInfo(c),
			auth.WithDatabase(db))
	case Ingest:
		l := services.NewLogger(
			os.Stdout,
			"Ingest Service:  ",
			log.LstdFlags|log.Lmsgprefix,
		)
		ge := services.NewRouter()
		v := services.NewViper(
			"IngestConfig",
			"toml",
			os.Getenv("CONFIG"),
		)
		c := services.NewConnection("ConnInfo", v)
		db := services.NewDatabase(c)
		s = ingest.IngestBuilder(
			ingest.WithLogger(l),
			ingest.WithRouter(ge),
			ingest.WithViper(v),
			ingest.WithConnectionInfo(c),
			ingest.WithDatabase(db),
			ingest.WithBatch(256),
		)
	case Search:
		l := services.NewLogger(
			os.Stdout,
			"Search Service:  ",
			log.LstdFlags|log.Lmsgprefix,
		)
		ge := services.NewRouter()
		v := services.NewViper(
			"SearchConfig",
			"toml",
			os.Getenv("CONFIG"),
		)
		c := services.NewConnection("ConnInfo", v)
		db := services.NewDatabase(c)
		s = search.SearchBuilder(
			search.WithLogger(l),
			search.WithRouter(ge),
			search.WithViper(v),
			search.WithConnectionInfo(c),
			search.WithDatabase(db),
		)
	}

	if *migrateFlag {
		ci := s.ExposeConnection()
		if *wipeFlag {
			if err := utils.MigrateWipe(os.Getenv("MIGRATIONS"), ci); err != nil {
				MainLogger.Fatalf("The function failed: %v\n", err)
			}
			MainLogger.Println("Database wiped.")
			return
		}

		if err := utils.MigrateDatabase(*versionFlag, ci,
			os.Getenv("MIGRATIONS"), upFlag, forceFlag); err != nil {
			MainLogger.Fatalf("The function failed: %v\n", err)
		}
		MainLogger.Println("Migration completed.")
		return
	}

	MainLogger.Printf("%v is starting...\n", s)
	err := s.Start()
	MainLogger.Printf("Server returned value: %v.\n", err)
}

// TODO:
// -normalizacja i oczyszczenie stringow (baza sie buguje)
// - utworzenie widokow w bazie danych
// - dodac opisy co funkcje robia
