import os
from typing import List, Dict, Optional

from sqlalchemy import create_engine, text

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "mysql+pymysql://root:passwd@127.0.0.1:3307/recommender"
)

engine = create_engine(DATABASE_URL, pool_pre_ping=True, future=True)


def _table_exists(conn, table: str) -> bool:
    q = text("""
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = DATABASE() AND table_name = :t
        LIMIT 1
    """)
    return conn.execute(q, {"t": table}).first() is not None


def _column_exists(conn, table: str, col: str) -> bool:
    q = text("""
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = DATABASE()
          AND table_name = :t
          AND column_name = :c
        LIMIT 1
    """)
    return conn.execute(q, {"t": table, "c": col}).first() is not None


def fetch_items_from_db(limit_movies: int = 20000, limit_books: int = 20000) -> List[Dict]:
    """
    Zwraca itemy w formacie ML:
    {"id": int, "type": "movie|book|series", "title": str, "genres": [str,...]}
    """
    items: List[Dict] = []

    try:
        with engine.connect() as conn:
            if _table_exists(conn, "movies") and _table_exists(conn, "movie2genres") and _table_exists(conn, "genres"):
                q_movies = text("""
                    SELECT
                        m.tmdb_id AS id,
                        m.title   AS title,
                        GROUP_CONCAT(DISTINCT g.genre ORDER BY g.genre SEPARATOR ',') AS genres
                    FROM movies m
                    LEFT JOIN movie2genres m2g ON m2g.movie_id = m.tmdb_id
                    LEFT JOIN genres g         ON g.ID = m2g.genre_id
                    GROUP BY m.tmdb_id, m.title
                    ORDER BY m.tmdb_id
                    LIMIT :lim
                """)
                for row in conn.execute(q_movies, {"lim": int(limit_movies)}):
                    glist = row.genres.split(",") if row.genres else []
                    items.append({
                        "id": int(row.id),
                        "type": "movie",
                        "title": (row.title or "").strip(),
                        "genres": [g.strip() for g in glist if g and g.strip()],
                    })

            if _table_exists(conn, "books"):
                has_authors = _table_exists(conn, "authors") and _column_exists(conn, "authors", "book_id")

                if has_authors:
                    q_books = text("""
                        SELECT
                            b.ID AS id,
                            b.title AS title,
                            GROUP_CONCAT(DISTINCT a.author ORDER BY a.author SEPARATOR ',') AS authors
                        FROM books b
                        LEFT JOIN authors a ON a.book_id = b.ID
                        GROUP BY b.ID, b.title
                        ORDER BY b.ID
                        LIMIT :lim
                    """)
                    for row in conn.execute(q_books, {"lim": int(limit_books)}):
                        alist = row.authors.split(",") if row.authors else []
                        items.append({
                            "id": int(row.id),
                            "type": "book",
                            "title": (row.title or "").strip(),
                            # traktujemy autorów jako "tagi/genres" dla content-based
                            "genres": [a.strip() for a in alist if a and a.strip()][:8],
                        })
                else:
                    q_books = text("""
                        SELECT b.ID AS id, b.title AS title
                        FROM books b
                        ORDER BY b.ID
                        LIMIT :lim
                    """)
                    for row in conn.execute(q_books, {"lim": int(limit_books)}):
                        items.append({
                            "id": int(row.id),
                            "type": "book",
                            "title": (row.title or "").strip(),
                            "genres": [],
                        })

            if _table_exists(conn, "top_100_shows"):
                title_col = "title" if _column_exists(conn, "top_100_shows", "title") else ("name" if _column_exists(conn, "top_100_shows", "name") else None)
                id_col = "ID" if _column_exists(conn, "top_100_shows", "ID") else ("id" if _column_exists(conn, "top_100_shows", "id") else None)

                if title_col and id_col:
                    q_series = text(f"""
                        SELECT {id_col} AS id, {title_col} AS title
                        FROM top_100_shows
                        ORDER BY {id_col}
                        LIMIT 20000
                    """)
                    for row in conn.execute(q_series):
                        items.append({
                            "id": int(row.id),
                            "type": "series",
                            "title": (row.title or "").strip(),
                            "genres": [],
                        })

    except Exception as e:
        print(f"[DB] fetch_items_from_db skipped: {e}")

    return items


def fetch_interactions_from_db() -> List[Dict]:
    interactions: List[Dict] = []

    try:
        with engine.connect() as conn:
            if not _table_exists(conn, "user_interactions"):
                return []

            q = text("""
                SELECT user_id, item_id, item_type, event_type
                FROM user_interactions
            """)
            for row in conn.execute(q):
                interactions.append({
                    "user_id": int(row.user_id),
                    "item_id": int(row.item_id),
                    "item_type": str(row.item_type),
                    "event": str(row.event_type),
                })

    except Exception as e:
        print(f"[DB] fetch_interactions_from_db skipped: {e}")

    return interactions