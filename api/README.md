# Serwis rekomendacji filmów, książek, seriali


# OPIS
Niezależny mikroserwis ML wystawiony przez HTTP (FastAPI).

# KONFIGURACJA I URUCHOMIENIE

Wymagania: Python >= 3.9, pip

Instalacja:
$ pip install -r requirements.txt

Start serwisu:
$ uvicorn main:app --host 0.0.0.0 --port 8001

Base URL: http://localhost:8001
Dokumentacja Swagger: http://localhost:8001/docs

# ENDPOINTY API

1. GET /health
   - Opis: Sprawdzenie stanu serwisu.
   - Odpowiedź:
     ```
     {"ok": true, "items_loaded": 3, "index_ready": true}

2. POST /sync/items
   - Opis: Zasilenie modelu listą itemów (wymagane przed /recommend).
   - Body:
     ```json
     {
       "full_replace": boolean,
       "items": [
         {"id": 1, "type": "movie", "title": "Matrix", "genres": ["Sci-Fi"]}
       ]
     }

3. POST /recommend
   - Opis: Generowanie rekomendacji dla użytkownika.
   - Body:
     ```
     {
       "user_id": number,
       "liked_items": [{"id": 1, "type": "movie"}],
       "target_type": "movie|book|series",
       "limit": 5,
       "diversity": 0.2
     }

4. POST /feedback (opcjonalny)
   - Opis: Przekazanie info o interakcji (like).
   - Body:
     ```json{
     {"user_id": 42, "item_id": 2, "item_type": "movie", "event": "like"}


