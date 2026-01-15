import platform
import pathlib
import os
import subprocess
from typing import Tuple
import requests 
import zipfile
from functools import wraps
"""
    Przed uruchomieniem `setup.py` odpowiednie bazy danych powinny zostać
    utworzone. Zalecane jest utworzenie instancji dockera:

    docker run -p PORT_DOCKER:PORT_BAZY_DANYCH
    --name NAZWA_KONTENERA_DOCKERA \\ 
    -e MYSQL_ROOT_PASSWORD=HASŁO \\
    -e MYSQL_DATABASE=NAZWA_BAZY_DANYCH \\
    -d mysql:latest
"""

class ServiceConfig:
    TEMPLATE="""[ConnInfo]
type = "{}"
name = "{}"
ip = "{}"
port = {} 
username = "{}"
password = '{}'"""
    def __init__(self, full_path: str):
        self.full_path=full_path
        # Tylko ten typ jest obsługiwany
        self.type: str="mysql"
        self.name: str=""
        # Tylko localhost
        self.ip: str="127.0.0.1"
        self.port: str=""
        self.username: str=""
        self.password: str=""

    def into_config_file(self) -> bool:
        with open(self.full_path, 'w') as obj:
            obj.write(
                self.TEMPLATE.format(
                    self.type,
                    self.name,
                    self.ip,
                    self.port,
                    self.username,
                    self.password
                )
            )
        return True



CURRENT_OS: str = platform.system()
CFG_PATH: pathlib.Path = pathlib.Path(os.getcwd(), "api", "configs")
MAIN_FILE: pathlib.Path = pathlib.Path(os.getcwd(), "cmd", "app", "main.go")
TEMP_PATH: pathlib.Path = pathlib.Path(os.getcwd(), "temp")
N_SERVICES: int = 3
SERVICES: Tuple = ("AuthConfig.toml", "IngestConfig.toml", "SearchConfig.toml")

assert N_SERVICES == len(SERVICES), f"""Liczba serwisów ({N_SERVICES}) nie zgadza się z 
zdefiniowanymi plikami konfiguracyjnymi ({SERVICES})"""

def ask_skip_prompt(prompt: str):
    def call_skip_prompt_wrapper(fn):
        @wraps(fn)
        def call_skip_prompt():
            if input(f"[S] Pomiń {prompt}: ").upper() == "S":
                return
            fn()
        return call_skip_prompt
    return call_skip_prompt_wrapper 


@ask_skip_prompt("konfiguracje plików")
def make_config_files():
    for cfg_filename in SERVICES:
        print(f"Config dla {cfg_filename}")
        cfg = ServiceConfig(CFG_PATH / cfg_filename)
        cfg.name = input("Podaj nazwę bazy danych: ")
        cfg.port = input("Podaj port dockera: ")
        cfg.username = input("Podaj nazwę użytkownika: ")
        cfg.password = input("Podaj hasło: ")
        assert cfg.into_config_file(), f"""nie udało się utworzyć pliku konfiguracyjnego 
        ({cfg_filename})\n\n{cfg}"""

@ask_skip_prompt("migracje baz danych")
def make_migrations():
    # 2. Migracje
    service_map = {
        "1" : "auth",
        "2" : "ingest",
        "3" : "search"
    }
    print("Jakie migracje przeprowadzić?")
    print("[1] Auth Service", "[2] Ingest Service", 
          "[3] Search Serivce", sep="\n")
    print("Użycie: >>> 1 3")
    migrations = input(">>> ").split(" ")
    assert len(migrations) <= N_SERVICES, "niepoprawne opcje"
    for migration_option in migrations:
        assert migration_option in service_map, f"{migration_option} nie jest prawidłową opcją"
        subprocess.run(["go", "run", MAIN_FILE,
                        f"--service={service_map[migration_option]}", 
                        "--migrate", "--up"])

@ask_skip_prompt("pobieranie danych")
def download_data():
    def download_file(name, url) -> Tuple:
        zip_path = TEMP_PATH / f"{name}.zip"
        unzipped_path = TEMP_PATH / name

        if zip_path.exists():
            return (None, None)

        print(f"Pobieranie: {name}...")
        resp = requests.get(url, allow_redirects=True)
        with open(zip_path, "wb") as obj:
            obj.write(resp.content)
        return (zip_path, unzipped_path)

    DATA = {
        "goodreads-books-data": "https://www.kaggle.com/api/v1/datasets/download/jealousleopard/goodreadsbooks",
        "books-data": "https://www.kaggle.com/api/v1/datasets/download/elvinrustam/books-dataset",
        "tmdb-movies-data": "https://www.kaggle.com/api/v1/datasets/download/tmdb/tmdb-movie-metadata",
        "movies-data": "https://www.kaggle.com/api/v1/datasets/download/rounakbanik/the-movies-dataset"
    }
    TEMP_PATH.mkdir(parents=True, exist_ok=True)
    downloads = [
        download_file(name, url) for name, url in DATA.items()
    ]
    for zip_path, unzipped_path in downloads:
        if zip_path is not None and zip_path.exists():
            with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                zip_ref.extractall(unzipped_path)
            os.remove(zip_path)



def main():
    make_config_files()
    make_migrations()
    download_data()

if __name__ == "__main__":
    main()
