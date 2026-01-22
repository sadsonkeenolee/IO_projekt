import os
from typing import List, Dict, Optional

from sqlalchemy import create_engine, text

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "mysql+pymysql://root:test@127.0.0.1:3308/data"
)

engine = create_engine(DATABASE_URL, pool_pre_ping=True, future=True)


def fetch_items_from_db(limit_movies: int = 50000, limit_books: int = 50000) -> List[Dict]:
    items: List[Dict] = []

    q_movies = text("""
        SELECT
          m.tmdb_id AS id,
          'movie'   AS type,
          m.title   AS title,
          GROUP_CONCAT(g.genre ORDER BY g.genre SEPARATOR ',') AS genres
        FROM movies m
        LEFT JOIN movie2genres mg ON mg.movie_id = m.tmdb_id
        LEFT JOIN genres g        ON g.ID = mg.genre_id
        GROUP BY m.tmdb_id, m.title
        ORDER BY m.tmdb_id
        LIMIT :limit_movies
    """)

    q_books = text("""
                   SELECT b.ID                                                 AS id,
                          'book'                                               AS type,
                          b.title                                              AS title,
                          GROUP_CONCAT(g.genre ORDER BY g.genre SEPARATOR ',') AS genres
                   FROM books b
                            LEFT JOIN book2genres bg ON bg.book_id = b.ID
                            LEFT JOIN genres g ON g.ID = bg.genre_id
                   GROUP BY b.ID, b.title
                   ORDER BY b.ID LIMIT :limit_books
                   """)

    try:
        with engine.connect() as conn:
            for row in conn.execute(q_movies, {"limit_movies": int(limit_movies)}):
                genres_list = row.genres.split(",") if row.genres else []
                items.append({
                    "id": int(row.id),
                    "type": "movie",
                    "title": row.title or "",
                    "genres": [g.strip() for g in genres_list if g and g.strip()],
                })

            for row in conn.execute(q_books, {"limit_books": int(limit_books)}):
                items.append({
                    "id": int(row.id),
                    "type": "book",
                    "title": row.title or "",
                    "genres": [],
                })

    except Exception as e:
        print(f"[DB] fetch_items_from_db skipped: {e}")

    return items


def fetch_interactions_from_db() -> List[Dict]:
    return []