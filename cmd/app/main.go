package main

import (
	"flag"
	"log"
	"path/filepath"

	"os"

	"github.com/sadsonkeenolee/IO_projekt/internal/services/credentials"
	"github.com/sadsonkeenolee/IO_projekt/pkg/utils"
)

func main() {
	l := log.New(os.Stdout, "main: ", log.LstdFlags)
	// Define all the necessary environment variables.
	configsPath, err := filepath.Abs("api/configs")
	if err != nil {
		l.Fatalf("The function failed: %v", err)
	}

	migrationsPath, err := filepath.Abs("api/migrations/credentials")
	if err != nil {
		l.Fatalf("The function failed: %v", err)
	}
	migrationsPath = "file://" + migrationsPath

	_ = os.Setenv("CREDENTIALS_CONFIG_DIR_PATH", configsPath)
	_ = os.Setenv("CREDENTIALS_MIGRATIONS_DIR_PATH", migrationsPath)

	// Define flags
	migrateFlag := flag.Bool("migrate", false, "run a migration in the current app")
	UpFlag := flag.Bool("up", false, "run an up migration, otherwise down")
	versionFlag := flag.Int("version", 0, "specify the number of the version")
	flag.Parse()

	c, _ := credentials.NewCredentials()
	if *migrateFlag {
		ci := c.ExposeConnInfo()
		if err := utils.MigrateToVersion(*versionFlag, ci, &migrationsPath, UpFlag); err != nil {
			l.Fatalf("The function failed: %v", err)
		}
		l.Println("Migration completed.")
		return
	}

	err = c.Start()
	l.Printf("Server returned value: %v.\n", err)
}
