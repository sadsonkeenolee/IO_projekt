# Pliki Konfiguracyjne
Tutaj znajdują się wszystkie pliki konfiguracyjne.
---
## Nazwy Plików
Obecnie serwisy obsługują:
`CredentialsConfig.toml`, `EtlConfig.toml`, `FetchConfig.toml`.
Plik musi wyglądać w następujący sposób:

`*Config.toml`
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
Każda migracja powinna mieć `.up` i `.down`. Ponadto, na początku nazwy należy podać wersję (numer), np.: `1_user_table.up.sql` i `1_user_table.down.sql`. Następna migracja będzie zaczynać się od `2_(...).up.sql` oraz `2_(...).down.sql.`