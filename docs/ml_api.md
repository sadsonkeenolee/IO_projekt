# ML API

## 1. Healthcheck

### GET /ml/health
Służy do sprawdzenia, czy serwis ML działa.

#### Response 200
```json
{
  "status": "ok",
  "service": "ml",
  "version": "0.1.0"
}
```
## 2. Rekomendacje

### POST /ml/recommend
Serwis ML przyjmuje od backendu listę ulubionych elementów użytkownika i zwraca listę rekomendacji z wartością score.

```json
{
  "user_id": 123,
  "liked_items": [
    { "id": 10, "type": "movie" },
    { "id": 51, "type": "book" }
  ],
  "limit": 20
}

```
Parametry:
- user_id - opcjonalne ID użytkownika
- liked_items - lista polubionych obiektów
- limit - ile rekomendacji ma zwrócić serwis ML

#### Response 200
```json
{
  "items": [
    { "id": 101, "type": "movie",   "score": 0.95 },
    { "id": 202, "type": "book",    "score": 0.89 },
    { "id": 303, "type": "concert", "score": 0.88 }
  ]
}
```

Opis pól:
- id - ID elementu w bazie
- type - "movie", "book", "concert" (typ rozrywki: film, książka lub koncert)
- score - wartość z przedziału [0,1] oznaczająca stopień dopasowania

## 3. Feedback

### POST /ml/feedback

Backend wysyła reakcje użytkownika na rekomendacje (like/dislike)

```json
{
  "user_id": 123,
  "item_id": 101,
  "item_type": "movie",
  "event": "like",
  "score_shown": 0.95
}
```

Opis pól:
- event - "like", "dislike"
- score_shown - wynik rekomendacji
