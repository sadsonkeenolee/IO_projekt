import os
from sqlalchemy import create_engine, text
from typing import List, Dict

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "mysql+pymysql://root:passwd@localhost:3307/recommender"
)

engine = create_engine(DATABASE_URL, pool_pre_ping=True, future=True)


def fetch_items_from_db() -> List[Dict]:
    items: List[Dict] = []

    query = text("""
        SELECT id, type, title, genres
        FROM media_items
    """)

    try:
        with engine.connect() as conn:
            result = conn.execute(query)
            for row in result:
                genres_list = (
                    row.genres.split(",")
                    if isinstance(row.genres, str)
                    else []
                )

                items.append({
                    "id": row.id,
                    "type": row.type,
                    "title": row.title,
                    "genres": [g.strip() for g in genres_list],
                })
    except Exception as e:
        # BARDZO WAŻNE: ML NIE MA SIĘ WYWRACAĆ
        print(f"[DB] fetch_items_from_db skipped: {e}")

    return items


def fetch_interactions_from_db() -> List[Dict]:
    interactions: List[Dict] = []

    query = text("""
        SELECT user_id, item_id, item_type, event_type
        FROM user_interactions
    """)

    try:
        with engine.connect() as conn:
            result = conn.execute(query)
            for row in result:
                interactions.append({
                    "user_id": row.user_id,
                    "item_id": row.item_id,
                    "item_type": row.item_type,
                    "event": row.event_type,
                })
    except Exception as e:
        print(f"[DB] fetch_interactions_from_db skipped: {e}")

    return interactions