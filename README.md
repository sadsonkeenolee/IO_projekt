# System rekomendacji filmów, seriali i koncertów.
___
## Wymagania
Tutaj wymagania.
___
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
___
## CLI
### `--migrate`
Migracja bazy danych dla danego serwisu.
- `--up`: Jeżeli ta flaga jest obecna, to aplikujemy zmiany, w przeciwnym razie wycofujemy.
- `--version`: Jeżeli wcześniej `--up` jest obecne, to baza danych zostanie uaktualniona do tej wersji. W przeciwnym wypadku wycofamy daną liczbę wersji.
