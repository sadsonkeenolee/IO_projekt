package main

import (
	"flag"
	"log"
	"path/filepath"
	"runtime"
	"strings"

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
var apiFlag = flag.String("api", "", "api key for TMDB")

const (
	Invalid = iota
	Auth
	Ingest
	Search
	Platform string = runtime.GOOS
)

var ServiceMap = map[string]uint{
	"auth":   Auth,
	"ingest": Ingest,
	"search": Search,
}

var EnvVariables map[string]string = map[string]string{
	"PLATFORM":     Platform,
	"DOWNLOAD_DIR": InitPath("temp/"),
	"MIGRATIONS":   InitPath("api/migrations"),
	"CONFIG":       InitPath("api/configs"),
}

func ConstructMigrationPath(migrationPath string) string {
	return "file://" + migrationPath
}

func InitPath(subpath string) string {
	if fullPath, err := filepath.Abs(subpath); err == nil {
		if Platform == "windows" {
			return strings.TrimPrefix(filepath.ToSlash(fullPath), "./")
		}
		return fullPath
	}
	return ""
}

func InitJoinedPath(subpaths ...string) string {
	if len(subpaths) <= 1 {
		return subpaths[0]
	}
	if fullPath := filepath.Join(subpaths...); fullPath != "" {
		if Platform == "windows" {
			return strings.TrimPrefix(filepath.ToSlash(fullPath), "./")
		}
		return fullPath
	}
	return ""
}

func CheckIfPlatformIsSupported() bool {
	if platform := EnvVariables["PLATFORM"]; platform == "linux" || platform == "windows" {
		return true
	}
	return false
}

func main() {
	flag.Parse()
	if !CheckIfPlatformIsSupported() {
		MainLogger.Println("This platform is not supported, some of the functions my not work.")
	}

	if _, ok := ServiceMap[*serviceNameFlag]; !ok {
		MainLogger.Fatalf("`%v` is an incorrect service, flag.\n", *serviceNameFlag)
	}

	// Set a correct migration file.
	switch ServiceMap[*serviceNameFlag] {
	case Auth:
		EnvVariables["MIGRATIONS"] = ConstructMigrationPath(InitJoinedPath(EnvVariables["MIGRATIONS"], "auth"))
	case Ingest:
		EnvVariables["MIGRATIONS"] = ConstructMigrationPath(InitJoinedPath(EnvVariables["MIGRATIONS"], "ingest"))
		if *apiFlag != "" {
			EnvVariables["TMDB_API_KEY"] = *apiFlag
			EnvVariables["TMDB_FETCH_URL"] = "https://api.themoviedb.org/3/search/tv?query=%v&include_adult=true&language=en-US&page=1"
		} else if !*migrateFlag {
			MainLogger.Fatalf("Missing TMDB API key for ingest service. Pass --api=\"<key>\" to run the service.")
		}
	case Search:
		EnvVariables["MIGRATIONS"] = ConstructMigrationPath(InitJoinedPath(EnvVariables["MIGRATIONS"], "search"))
		if *apiFlag != "" {
			EnvVariables["TMDB_API_KEY"] = *apiFlag
			EnvVariables["TMDB_FETCH_URL"] = "https://api.themoviedb.org/3/search/tv?query=%v&include_adult=true&language=en-US&page=1"
		} else if !*migrateFlag {
			MainLogger.Fatalf("Missing TMDB API key for search service. Pass --api=\"<key>\" to run the service.")
		}
	}
	for k, v := range EnvVariables {
		if k == "" || v == "" {
			MainLogger.Fatalf("Key-value (`%v`, `%v`) is not a valid pair of values.", k, v)
		}
		if err := os.Setenv(k, v); err != nil {
			MainLogger.Fatalf("Key-value (`%v`, `%v`) couldn't be set as environment variables, reason: %v", k, v, err)
		}
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

// FIXME:
//	1. Czasem pojawia się problem z usunięciem procedur. Na razie rozwiązanie to
//	usunięcie procedur manualnie.

// TODO:
// - Uzupelnianie pustych rekordow
