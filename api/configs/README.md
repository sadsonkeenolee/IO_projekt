# Pliki Konfiguracyjne
Tutaj znajdują się wszystkie pliki konfiguracyjne.
## Nazwy Plików
Obecnie serwisy obsługują:
* `CredentialsConfig.toml`
    ```toml
        [ConnInfo]
        type = "mysql"    # typ bazy danych (postgres, sql, itp.)
        name = "baza"     # nazwa bazy danych 
        ip = "127.0.0.1"  # IP bazy danych 
        port = 3307       # port bazy danych 
        username = "user" # login dla bazy danych 
        password = "pass" # hasło dla bazy danych 
    ```
    Należy zapoznać się ze standardem dla danej bazy danych.
## Migracje
Każda migracja powinna mieć `.up` lub `.down`. Ponadto, na początku nazwy należy podać wersję, np.: `1_user_table.up.sql` i `1_user_table.down.sql`. 
Następna migracja będzie zaczynać się od `2_(...).up.sql` oraz `2_(...).down.sql`.