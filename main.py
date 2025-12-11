from typing import List, Optional, Literal, Dict, Set
from fastapi import FastAPI
from pydantic import BaseModel, Field
import re
from collections import Counter
import math
import numpy as np

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
N_GENRES = len(ALL_GENRES)

N_ITEMS = len(ALL_ITEMS_DB)
genre_counts = np.zeros(N_GENRES)

for item in ALL_ITEMS_DB:
    for genre in item['genres']:
        idx = GENRE_INDEX[genre]
        genre_counts[idx] += 1

EPS = 1e-8
genre_probs = genre_counts / (N_ITEMS + EPS)
GENRE_WEIGHTS = np.log(1.0 / (genre_probs + EPS))


def tokenize(text: str) -> List[str]:
    text = text.lower()
    words = re.findall(r'[a-ząćęłńóśżź0-9]+', text)
    tokens = words.copy()
    if len(words) > 1:
        for i in range(len(words) - 1):
            tokens.append(f"{words[i]}_{words[i + 1]}")
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
N_FEATURES = N_GENRES + VOCAB_SIZE

TITLE_IDF = np.zeros(VOCAB_SIZE)
for token, idx in vocab_index.items():
    df = df_counts[token]
    TITLE_IDF[idx] = math.log(N_ITEMS / (1.0 + df)) if df > 0 else 0.0


def encode_item(item: Dict) -> np.ndarray:
    vec = np.zeros(N_FEATURES)

    for genre in item['genres']:
        idx = GENRE_INDEX.get(genre)
        if idx is not None:
            vec[idx] = GENRE_WEIGHTS[idx]

    tokens = tokenize(item['title'])
    if tokens:
        counts = Counter(tokens)
        total = len(tokens)
        for token, count in counts.items():
            idx = vocab_index.get(token)
            if idx is not None:
                tf = count / total
                vec[N_GENRES + idx] = tf * TITLE_IDF[idx]

    return vec


ITEM_MATRIX = np.zeros((N_ITEMS, N_FEATURES))
ID_TO_INDEX = {}
for i, item in enumerate(ALL_ITEMS_DB):
    ITEM_MATRIX[i] = encode_item(item)
    ID_TO_INDEX[item['id']] = i


def apply_svd(matrix: np.ndarray, k: int = 5):
    mean_vec = np.mean(matrix, axis=0)
    centered_matrix = matrix - mean_vec

    U, s, Vt = np.linalg.svd(centered_matrix, full_matrices=False)

    k = min(k, Vt.shape[0])
    components = Vt[:k]

    reduced_matrix = np.dot(centered_matrix, components.T)
    return reduced_matrix, components, mean_vec


LATENT_DIM = 5
ITEM_MATRIX_REDUCED, SVD_COMPONENTS, SVD_MEAN = apply_svd(ITEM_MATRIX, k=LATENT_DIM)


def cosine_similarity(a: np.ndarray, b: np.ndarray) -> float:
    dot = np.dot(a, b)
    na = np.linalg.norm(a)
    nb = np.linalg.norm(b)

    if na == 0 or nb == 0:
        return 0.0

    return float(dot / (na * nb))


def softmax(scores: List[float], temperature: float = 1.0) -> List[float]:
    if not scores:
        return []

    arr_scores = np.array(scores)
    max_score = np.max(arr_scores)
    exps = np.exp((arr_scores - max_score) / temperature)
    sum_exps = np.sum(exps)

    if sum_exps == 0:
        return list(exps)

    return list(exps / sum_exps)


def mmr_selection(
        user_vector: np.ndarray,
        candidate_indices: List[int],
        item_matrix: np.ndarray,
        limit: int,
        lambda_param: float = 0.5
) -> List[int]:
    selected_indices = []
    candidates = candidate_indices[:]

    if not candidates:
        return []

    sim_to_user = {}
    for idx in candidates:
        sim_to_user[idx] = cosine_similarity(user_vector, item_matrix[idx])

    while len(selected_indices) < limit and candidates:
        best_score = -float('inf')
        best_candidate = -1

        for candidate_idx in candidates:
            relevance = sim_to_user[candidate_idx]

            max_sim_to_selected = 0.0
            if selected_indices:
                similarities = []
                cand_vec = item_matrix[candidate_idx]
                for sel_idx in selected_indices:
                    sel_vec = item_matrix[sel_idx]
                    similarities.append(cosine_similarity(cand_vec, sel_vec))
                max_sim_to_selected = max(similarities)

            mmr_score = (lambda_param * relevance) - ((1 - lambda_param) * max_sim_to_selected)

            if mmr_score > best_score:
                best_score = mmr_score
                best_candidate = candidate_idx

        if best_candidate != -1:
            selected_indices.append(best_candidate)
            candidates.remove(best_candidate)
        else:
            break

    return selected_indices


def dcg_at_k(recommended_ids: List[int], relevant_ids: Set[int], k: int) -> float:
    k = min(k, len(recommended_ids))
    if k == 0: return 0.0

    rel = np.array([1.0 if pid in relevant_ids else 0.0 for pid in recommended_ids[:k]])
    if np.sum(rel) == 0: return 0.0

    discounts = np.log2(np.arange(len(rel)) + 2)
    return float(np.sum(rel / discounts))


def ndcg_at_k(recommended_ids: List[int], relevant_ids: Set[int], k: int) -> float:
    dcg = dcg_at_k(recommended_ids, relevant_ids, k)

    k_prime = min(len(relevant_ids), k)
    if k_prime == 0: return 0.0

    ideal_rel = np.ones(k_prime)
    ideal_discounts = np.log2(np.arange(k_prime) + 2)
    idcg = np.sum(ideal_rel / ideal_discounts)

    if idcg == 0: return 0.0
    return float(dcg / idcg)


ItemType = Literal['movie', 'book', 'concert']


class Item(BaseModel):
    id: int
    type: ItemType


class RecommendRequest(BaseModel):
    user_id: Optional[int] = None
    liked_items: List[Item]
    limit: int = Field(default=10, ge=1, le=50)
    temperature: float = Field(default=1.0, gt=0.0)
    diversity: float = Field(default=0.3, ge=0.0, le=1.0)


class RecommendedItem(Item):
    score: float
    prob: Optional[float] = None
    title: Optional[str] = None


class RecommendResponse(BaseModel):
    items: List[RecommendedItem]


class FeedbackRequest(BaseModel):
    user_id: int
    item_id: int
    item_type: ItemType
    event: str
    score_shown: Optional[float] = None


class EvaluateRequest(RecommendRequest):
    relevant_item_ids: List[int]
    k: int = 10


class EvaluateResponse(BaseModel):
    ndcg: float
    items: List[RecommendedItem]


@app.get("/ml/health")
def health():
    return {"status": "ok", "service": "ml", "version": "0.3.0"}


@app.post("/ml/recommend", response_model=RecommendResponse)
def recommend(req: RecommendRequest):
    liked_indices = []
    liked_ids = set()

    for incoming_item in req.liked_items:
        if incoming_item.id in ID_TO_INDEX:
            liked_indices.append(ID_TO_INDEX[incoming_item.id])
            liked_ids.add(incoming_item.id)

    if not liked_indices:
        candidates = []
        for item in ALL_ITEMS_DB[:req.limit]:
            candidates.append(
                RecommendedItem(
                    id=item["id"],
                    type=item["type"],
                    score=0.1,
                    prob=None,
                    title=item["title"],
                )
            )
        return RecommendResponse(items=candidates)

    liked_matrix_svd = ITEM_MATRIX_REDUCED[liked_indices]
    user_vector_svd = np.mean(liked_matrix_svd, axis=0)

    pre_candidates_scores = []

    for i in range(N_ITEMS):
        db_item = ALL_ITEMS_DB[i]
        if db_item["id"] in liked_ids:
            continue

        sim = cosine_similarity(user_vector_svd, ITEM_MATRIX_REDUCED[i])
        if sim > -1.0:
            pre_candidates_scores.append((i, sim))

    pre_candidates_scores.sort(key=lambda x: x[1], reverse=True)

    top_n_pool_size = min(len(pre_candidates_scores), req.limit * 3)
    top_candidates_indices = [x[0] for x in pre_candidates_scores[:top_n_pool_size]]

    mmr_lambda = 1.0 - req.diversity

    final_indices = mmr_selection(
        user_vector=user_vector_svd,
        candidate_indices=top_candidates_indices,
        item_matrix=ITEM_MATRIX_REDUCED,
        limit=req.limit,
        lambda_param=mmr_lambda
    )

    final_items = []
    raw_scores = []

    for idx in final_indices:
        db_item = ALL_ITEMS_DB[idx]
        score = cosine_similarity(user_vector_svd, ITEM_MATRIX_REDUCED[idx])

        final_items.append(RecommendedItem(
            id=db_item["id"],
            type=db_item["type"],
            score=round(float(score), 4),
            prob=None,
            title=db_item["title"],
        ))
        raw_scores.append(score)

    probs = softmax(raw_scores, temperature=req.temperature)
    for c, p in zip(final_items, probs):
        c.prob = round(float(p), 6)

    return RecommendResponse(items=final_items)


@app.post("/ml/feedback")
def feedback(req: FeedbackRequest):
    print(f"LOG: User {req.user_id} {req.event} item {req.item_id} (score_shown={req.score_shown})")
    return {"status": "logged"}


@app.post("/ml/evaluate", response_model=EvaluateResponse)
def evaluate(req: EvaluateRequest):
    rec_resp = recommend(req)

    recommended_ids = [item.id for item in rec_resp.items]
    relevant_ids = set(req.relevant_item_ids)

    ndcg_value = ndcg_at_k(recommended_ids, relevant_ids, req.k)

    return EvaluateResponse(
        ndcg=ndcg_value,
        items=rec_resp.items,
    )