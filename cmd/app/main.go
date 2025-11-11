package main

import (
	"fmt"
	"path/filepath"

	"github.com/sadsonkeenolee/IO_projekt/internal/services/credentials"
	"os"
)

func main() {
	// WARNING: Jesli serwisy beda rozdzielone, to sciezki musza byc tu
	// WARNING: Ta apka to TEORETYCZNIE serwis logowania / rejestracji
	credentialsConfigDirPath, err := filepath.Abs("api/configs")
	if err != nil {
		panic(fmt.Sprintf("The function failed: %v", err))
	}

	credentialsMigrationsDirPath, err := filepath.Abs("api/configs/migrations/credentials")
	if err != nil {
		panic(fmt.Sprintf("The function failed: %v", err))
	}

	_ = os.Setenv("CREDENTIALS_CONFIG_DIR_PATH", credentialsConfigDirPath)
	_ = os.Setenv("CREDENTIALS_MIGRATIONS_DIR_PATH", credentialsMigrationsDirPath)
	c, _ := credentials.NewCredentials()
	c.Start()
}
