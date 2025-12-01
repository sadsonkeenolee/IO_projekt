from typing import List, Optional, Literal, Dict
from fastapi import FastAPI
from pydantic import BaseModel
import re
from collections import Counter
import math

app = FastAPI()

ALL_ITEMS_DB = [
    {"id": 1, "type": "movie", "title": "Szybcy i Wściekli", "genres": ["Action", "Crime"]},
    {"id": 2, "type": "book", "title": "Władca Pierścieni", "genres": ["Fantasy", "Adventure"]},
    {"id": 3, "type": "movie", "title": "Matrix", "genres": ["Action", "Sci-Fi"]},
    {"id": 4, "type": "movie", "title": "Avengers", "genres": ["Action", "Sci-Fi", "Adventure"]},
    {"id": 5, "type": "movie", "title": "Notting Hill", "genres": ["Romance", "Comedy"]},
    {"id": 6, "type": "book", "title": "Harry Potter", "genres": ["Fantasy", "Adventure"]},
    {"id": 7, "type": "book", "title": "Sherlock Holmes", "genres": ["Crime", "Mystery"]},
    {"id": 8, "type": "concert", "title": "Metallica Live", "genres": ["Music", "Metal"]},
    {"id": 9, "type": "concert", "title": "Chopin Piano", "genres": ["Music", "Classical"]},
]

ALL_GENRES = sorted(list(set(g for item in ALL_ITEMS_DB for g in item['genres'])))
GENRE_INDEX = {genre: i for i, genre in enumerate(ALL_GENRES)}
Vector = List[float]

N_ITEMS = len(ALL_ITEMS_DB)
genre_counts = [0] * len(ALL_GENRES)

for item in ALL_ITEMS_DB:
    for genre in item['genres']:
        idx = GENRE_INDEX[genre]
        genre_counts[idx] += 1

GENRE_WEIGHTS: List[float] = []

def tokenize(text: str) -> List[str]:
    text = text.lower()
    tokens = re.findall(r'[a-ząćęłńóśżź0-9]+', text)
    return tokens
vocab_index: Dict[str, int] = {}
df_counts: Counter = Counter()

for item in ALL_ITEMS_DB:
    tokens_in_title = set(tokenize(item['title']))
    for token in tokens_in_title:
        if token not in vocab_index:
            vocab_index[token] = len(vocab_index)
        df_counts[token] += 1

VOCAB_SIZE = len(vocab_index)

TITLE_IDF: List[float] = [0.0] * VOCAB_SIZE
for token, idx in vocab_index.items():
    df = df_counts[token]
    TITLE_IDF[idx] = math.log(N_ITEMS/(1.0 + df)) if df > 0 else 0.0

def encode_title_tfidf(title: str) -> List[float]:
    vector = [0.0] * VOCAB_SIZE
    tokens = tokenize(title)
    if not tokens:
        return vector

    counts = Counter(tokens)
    total = len(tokens)
    for token, count in counts.items():
        idx = vocab_index.get(token)
        if idx is None:
            continue
        tf = count/total
        vector[idx] = tf * TITLE_IDF[idx]
    return vector
EPS = 1e-8
for c in genre_counts:
    p = c / N_ITEMS if N_ITEMS > 0 else 0.0
    w = math.log(1.0/(p+EPS)) if p > 0 else 0.0
    GENRE_WEIGHTS.append(w)

LAMBDA_TEXT = 1.0
def encode_item(item: Dict) -> Vector:
  genres_vector = [0.0] * len(ALL_GENRES)

  for genre in item['genres']:
    idx = GENRE_INDEX.get(genre)
    if idx is not None:
      genres_vector[idx] = GENRE_WEIGHTS[idx]

  title_vector = encode_title_tfidf(item['title'])
  title_vector_scaled = [LAMBDA_TEXT * x for x in title_vector]
  return genres_vector + title_vector_scaled

def cosine_similarity(a: Vector, b: Vector) -> float:
  dot = sum(x*y for x,y in zip(a,b))
  na = sum(x*x for x in a) ** 0.5
  nb = sum(y*y for y in b) ** 0.5

  if na == 0 or nb == 0:
    return 0.0

  return dot/(na*nb)

ItemType = Literal['movie', 'book', 'concert']

class Item(BaseModel):
  id: int
  type: ItemType

class RecommendRequest(BaseModel):
  user_id: Optional[int] = None
  liked_items: List[Item]
  limit: int = 10

class RecommendedItem(Item):
  score: float
  title: Optional[str] = None

class RecommendResponse(BaseModel):
    items: List[RecommendedItem]

class FeedbackRequest(BaseModel):
  user_id: int
  item_id: int
  item_type: ItemType
  event: str
  score_shown: Optional[float] = None

@app.get("/ml/health")
def health():
    return {"status": "ok", "service": "ml", "version": "0.1.0"}

@app.post("/ml/recommend", response_model=RecommendResponse)
def recommend(req: RecommendRequest):
    liked_full_items = []
    liked_ids = set()
    
    for incoming_item in req.liked_items:
        found = next((x for x in ALL_ITEMS_DB if x["id"] == incoming_item.id), None)
        if found:
            liked_full_items.append(found)
            liked_ids.add(found["id"])
    if not liked_full_items:
        candidates = []
        for item in ALL_ITEMS_DB[:req.limit]:
            candidates.append(
                RecommendedItem(
                    id=item["id"], 
                    type=item["type"], # type: ignore
                    score=0.1, 
                    title=item["title"]
                )
            )
        return RecommendResponse(items=candidates)
    user_vector = [0.0] * len(encode_item(ALL_ITEMS_DB[0]))
    
    for item in liked_full_items:
        item_vec = encode_item(item)
        user_vector = [u + i for u, i in zip(user_vector, item_vec)]
    
    count = len(liked_full_items)
    if count > 0:
        user_vector = [x / count for x in user_vector]
    scored_candidates = []
    
    for db_item in ALL_ITEMS_DB:
        if db_item["id"] in liked_ids:
            continue
            
        item_vector = encode_item(db_item)
        score = cosine_similarity(user_vector, item_vector)
        
        if score > 0:
            scored_candidates.append(
                RecommendedItem(
                    id=db_item["id"],
                    type=db_item["type"], # type: ignore
                    score=round(score, 4),
                    title=db_item["title"]
                )
            )
    scored_candidates.sort(key=lambda x: x.score, reverse=True)
    return RecommendResponse(items=scored_candidates[:req.limit])

@app.post("/ml/feedback")
def feedback(req: FeedbackRequest):
    print(f"LOG: User {req.user_id} {req.event} item {req.item_id}")
    return {"status": "logged"}
