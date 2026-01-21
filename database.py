import os
from sqlalchemy import create_engine, text
from typing import List, Dict

DATABASE_URL = os.getenv("DATABASE_URL", "mysql+pymysql://root:passwd@localhost:3307/users")

engine = create_engine(DATABASE_URL)


def fetch_items_from_db() -> List[Dict]:
    items = []
    query = text("SELECT id, type, title, genres FROM media_items")

    with engine.connect() as conn:
        result = conn.execute(query)
        for row in result:
            genres_list = row.genres.split(',') if isinstance(row.genres, str) else []

            items.append({
                "id": row.id,
                "type": row.type,  # movie, series, book
                "title": row.title,
                "genres": [g.strip() for g in genres_list]
            })
    return items


def fetch_interactions_from_db() -> List[Dict]:
    interactions = []
    query = text("SELECT user_id, item_id, item_type, event_type FROM user_interactions")

    with engine.connect() as conn:
        result = conn.execute(query)
        for row in result:
            interactions.append({
                "user_id": row.user_id,
                "item_id": row.item_id,
                "item_type": row.item_type,
                "event": row.event_type
            })
    return interactions