# Pliki Konfiguracyjne
## Nazwy Plików
Aby serwis poprawnie działał należy stworzyć poniższe pliki:
- `AuthConfig.toml`
- `IngestConfig.toml`
- `SearchConfig.toml`
## Struktura
```toml
[ConnInfo]
type = "mysql"    # typ bazy danych (postgres, sql, itp.) (na razie tylko sql)
name = "data"     # nazwa bazy danych 
ip = "127.0.0.1"  # IP bazy danych (na razie tylko `localhost`)
port = 3307       # port bazy danych (port na który `docker` przekierowuje połączenia)
username = "root" # login dla bazy danych
password = "test" # hasło dla bazy danych
```
---
## Migracje
Każda migracja zawiera:
- `.up` i `.down`
- liczbę definiującą migrację: `1_(...).up.sql`, `1_(...).down.sql`
