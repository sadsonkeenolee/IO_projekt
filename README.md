# System rekomendacji filmów, seriali i koncertów.
---
## Wymagania
Tutaj wymagania.
---
## Katalogi
### `api/`
Konfiguracje: projektu, klas, zaimportowanych bibliotek i api, specyfikacje, itp.

### `internal/`
Kod (biblioteki), którego nie powinno się używać poza projektem.

### `cmd/`
Główny kod aplikacji, głównie pliki `main.*`

### `pkg/`
Kod (biblioteki), który można upublicznić.

### `docs/`
Dokumentacja.

### `web/`
Statyczne strony, inne pliki po stronie serwerowej.

---
## Wstęp
### Przed uruchomieniem
Każdy z serwisów powinien mieć skonfigurowany dostęp do baz danych - pliki konfiguracyjne powinny się znajdować w `/api/configs/` - [dowiedz się więcej](./api/configs/README.md). Zalecane jest utworzenie bazy danych za pomocą `dockera`, np:
```bash
docker run -p 3307:3306 \
--name users-db \ 
-e MYSQL_ROOT_PASSWORD=passwd \
-e MYSQL_DATABASE=users \
-d mysql:latest
```
### Pierwsze uruchomienie
Dostępne serwisy to:
- `auth`
- `ingest`
- `search`

Na każdym serwisie należy przeprowadzić migrację:
```bash
go run cmd/app/main.go --service="serwis" --migrate --up
```
### Każde następne uruchomienie
Każdy z serwisu uruchamia się:
```bash
go run cmd/app/main.go --service="serwis" --api="klucz"
```
`--api` to klucz api z strony [TMDB](https://developer.themoviedb.org/reference/getting-started), brak klucza uniemożliwi aktualizowanie/pobieranie nowych danych o serialach.
# Informacje
Pomoc dostępna pod:
```bash
go run cmd/app/main.go --help
```