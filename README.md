# System rekomendacji filmów, seriali i koncertów.
---
## Wymagania
Tutaj wymagania.
---
## Katalogi
### `api`
Konfiguracje: projektu, klas, zaimportowanych bibliotek i api, specyfikacje, itp.

### `internal`
Kod (biblioteki), którego nie powinno się używać poza projektem.

### `cmd`
Główny kod aplikacji, głównie pliki `main.*`

### `test`
Testy na funkcjonowanie **całego** projektu (testy jednostkowe funkcji / klas w konkretnych modułach)

### `pkg`
Kod (biblioteki), który można upublicznić.

### `docs`
Dokumentacja.

### `web`
Statyczne strony, inne pliki po stronie serwerowej.
---
## CLI
Wszystkie komendy znajdują się pod:
```bash
go run cmd/app/main.go --help
```
Standardowy setup zazwyczaj składa się z:
```bash
go run cmd/app/main.go --service "credentials" --migrate --up --version=0 # zaaplikuje najwyzsza migracje
go run cmd/app/main.go --service "credentials" # uruchomienie serwisu
```