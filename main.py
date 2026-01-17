import math
import re
from typing import List, Optional, Literal, Dict, Set, Tuple
from collections import Counter, defaultdict
from dataclasses import dataclass
import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

app = FastAPI(title="Recommender Service")

ItemType = Literal["movie", "book", "series"]
_TOKEN_RE = re.compile(r"[a-ząćęłńóśżź0-9]+", re.IGNORECASE)


def item_key(item_type: str, item_id: int) -> str:
    return f"{item_type}:{int(item_id)}"


class ItemPayload(BaseModel):
    id: int
    type: ItemType
    title: str
    genres: List[str]


class ItemsSyncRequest(BaseModel):
    items: List[ItemPayload]
    full_replace: bool = True


class InteractionPayload(BaseModel):
    user_id: int
    item_id: int
    item_type: ItemType
    event: Literal["like", "view", "purchase"]
    ts: Optional[int] = None


class InteractionsSyncRequest(BaseModel):
    interactions: List[InteractionPayload]


class ItemObj(BaseModel):
    id: int
    type: ItemType


class RecommendRequest(BaseModel):
    user_id: Optional[int] = None
    liked_items: List[ItemObj]
    target_type: Optional[ItemType] = None
    limit: int = Field(default=10, le=50, ge=1)
    diversity: float = Field(default=0.2, ge=0.0, le=1.0)
    max_candidates: int = Field(default=50_000, ge=1, le=200_000)
    mmr_pool: int = Field(default=500, ge=50, le=5000)
    min_score: float = Field(default=0.0001, ge=0.0, le=1.0)


class RecommendedItem(BaseModel):
    id: int
    type: ItemType
    title: str
    genres: List[str]
    score: float
    reason: str


class SparseTfidfIndex:
    def __init__(self):
        self.vocab: Dict[str, int] = {}
        self.idf: Optional[np.ndarray] = None
        self.indptr: Optional[np.ndarray] = None
        self.indices: Optional[np.ndarray] = None
        self.data: Optional[np.ndarray] = None
        self.doc_norm: Optional[np.ndarray] = None
        self.post_docs: List[np.ndarray] = []
        self.post_w: List[np.ndarray] = []
        self.genre_docs: Dict[str, np.ndarray] = {}
        self._genre_tmp: Dict[str, List[int]] = defaultdict(list)
        self.genre_cooc: Dict[str, Counter] = defaultdict(Counter)
        self.doc_key: List[str] = []
        self.key_to_doc: Dict[str, int] = {}
        self.doc_type: List[str] = []
        self.doc_title: List[str] = []
        self.doc_genres: List[List[str]] = []

    def reset(self):
        self.__init__()

    @staticmethod
    def tokenize(text: str) -> List[str]:
        return _TOKEN_RE.findall(text.lower())

    @staticmethod
    def bigrams(tokens: List[str]) -> List[str]:
        if len(tokens) < 2:
            return []
        return [f"{tokens[i]}_{tokens[i+1]}" for i in range(len(tokens) - 1)]

    @staticmethod
    def genre_token(genre: str) -> str:
        return f"genre:{genre.lower()}"

    def _get_tid(self, term: str, df_list: List[int]) -> int:
        tid = self.vocab.get(term)
        if tid is None:
            tid = len(self.vocab)
            self.vocab[term] = tid
            df_list.append(0)
        return tid

    def add_items(self, items: List[Dict]):
        df_list: List[int] = []
        doc_terms: List[List[Tuple[int, int]]] = []
        n = 0

        for it in items:
            k = item_key(it["type"], it["id"])
            doc_id = n
            n += 1

            self.doc_key.append(k)
            self.key_to_doc[k] = doc_id
            self.doc_type.append(it["type"])
            self.doc_title.append(it.get("title", ""))
            self.doc_genres.append(list(it.get("genres", [])))

            title_tokens = self.tokenize(it.get("title", ""))
            title_tokens += self.bigrams(title_tokens)

            genres = list(dict.fromkeys(it.get("genres", [])))
            for g in genres:
                self._genre_tmp[g].append(doc_id)

            for i in range(len(genres)):
                for j in range(len(genres)):
                    if i != j:
                        self.genre_cooc[genres[i]][genres[j]] += 1

            tf = Counter(title_tokens)
            for g in genres:
                tf[self.genre_token(g)] += 1

            terms_for_doc: List[Tuple[int, int]] = []
            for term, freq in tf.items():
                tid = self._get_tid(term, df_list)
                terms_for_doc.append((tid, int(freq)))
            terms_for_doc.sort(key=lambda x: x[0])

            for tid, _freq in terms_for_doc:
                df_list[tid] += 1

            doc_terms.append(terms_for_doc)

        self.genre_docs = {g: np.array(docs, dtype=np.int32) for g, docs in self._genre_tmp.items()}
        self._genre_tmp.clear()

        N = float(n)
        df = np.array(df_list, dtype=np.float32)
        self.idf = np.log((N + 1.0) / (df + 1.0)) + 1.0

        indptr = np.zeros(n + 1, dtype=np.int32)
        total_nnz = sum(len(x) for x in doc_terms)
        indices = np.zeros(total_nnz, dtype=np.int32)
        data = np.zeros(total_nnz, dtype=np.float32)
        doc_norm = np.zeros(n, dtype=np.float32)

        V = len(self.vocab)
        post_docs_l: List[List[int]] = [list() for _ in range(V)]
        post_w_l: List[List[float]] = [list() for _ in range(V)]

        cursor = 0
        for doc_id, terms in enumerate(doc_terms):
            indptr[doc_id] = cursor
            norm2 = 0.0

            for tid, freq in terms:
                w = (1.0 + math.log(1.0 + float(freq))) * float(self.idf[tid])
                indices[cursor] = tid
                data[cursor] = w
                norm2 += w * w
                post_docs_l[tid].append(doc_id)
                post_w_l[tid].append(w)
                cursor += 1

            doc_norm[doc_id] = math.sqrt(norm2) if norm2 > 0.0 else 0.0

        indptr[n] = cursor

        self.indptr = indptr
        self.indices = indices
        self.data = data
        self.doc_norm = doc_norm

        self.post_docs = [
            np.array(x, dtype=np.int32) if x else np.zeros(0, dtype=np.int32) for x in post_docs_l
        ]
        self.post_w = [
            np.array(x, dtype=np.float32) if x else np.zeros(0, dtype=np.float32) for x in post_w_l
        ]

    def expand_genres(self, base_genres: Set[str], top_k_per_genre: int = 2) -> Set[str]:
        out = set(base_genres)
        for g in base_genres:
            rel = self.genre_cooc.get(g)
            if not rel:
                continue
            for ng, _cnt in rel.most_common(top_k_per_genre):
                out.add(ng)
        return out

    def build_user_profile(self, liked_doc_ids: List[int]) -> Tuple[Dict[int, float], float, Set[str], Set[str]]:
        profile = defaultdict(float)
        title_tokens: Set[str] = set()
        user_genres: Set[str] = set()

        assert self.indptr is not None and self.indices is not None and self.data is not None

        for doc_id in liked_doc_ids:
            tt = self.tokenize(self.doc_title[doc_id])
            title_tokens.update(tt)
            title_tokens.update(self.bigrams(tt))
            for g in self.doc_genres[doc_id]:
                user_genres.add(g)

            a = int(self.indptr[doc_id])
            b = int(self.indptr[doc_id + 1])
            for i in range(a, b):
                tid = int(self.indices[i])
                profile[tid] += float(self.data[i])

        norm2 = 0.0
        for v in profile.values():
            norm2 += v * v
        profile_norm = math.sqrt(norm2) if norm2 > 0.0 else 0.0
        return dict(profile), profile_norm, title_tokens, user_genres

    def collect_candidates(self, query_tokens: Set[str], genres: Set[str], max_candidates: int = 50_000) -> np.ndarray:
        n_docs = len(self.doc_key)
        visited = bytearray(n_docs)
        out: List[int] = []

        for g in genres:
            docs = self.genre_docs.get(g)
            if docs is None or docs.size == 0:
                continue
            for di in docs:
                d = int(di)
                if not visited[d]:
                    visited[d] = 1
                    out.append(d)
                    if len(out) >= max_candidates:
                        return np.array(out, dtype=np.int32)

        for tok in query_tokens:
            tid = self.vocab.get(tok)
            if tid is None:
                continue
            docs = self.post_docs[tid]
            if docs.size == 0:
                continue
            for di in docs:
                d = int(di)
                if not visited[d]:
                    visited[d] = 1
                    out.append(d)
                    if len(out) >= max_candidates:
                        return np.array(out, dtype=np.int32)

        return np.array(out, dtype=np.int32)

    def score_candidates_content(self, profile: Dict[int, float], profile_norm: float, candidate_doc_ids: np.ndarray) -> np.ndarray:
        if profile_norm <= 0.0 or candidate_doc_ids.size == 0:
            return np.zeros(candidate_doc_ids.size, dtype=np.float32)

        assert self.doc_norm is not None

        sorter = np.argsort(candidate_doc_ids)
        cand_sorted = candidate_doc_ids[sorter]
        acc_sorted = np.zeros(cand_sorted.size, dtype=np.float32)

        for tid, q_w in profile.items():
            docs = self.post_docs[tid]
            if docs.size == 0:
                continue
            w = self.post_w[tid]
            idx = np.searchsorted(cand_sorted, docs)
            m = (idx < cand_sorted.size) & (cand_sorted[idx] == docs)
            if np.any(m):
                acc_sorted[idx[m]] += (float(q_w) * w[m])

        acc = np.zeros(acc_sorted.size, dtype=np.float32)
        acc[sorter] = acc_sorted

        cand_norm = self.doc_norm[candidate_doc_ids]
        denom = float(profile_norm) * cand_norm

        out = np.zeros(candidate_doc_ids.size, dtype=np.float32)
        mask = denom > 0.0
        out[mask] = acc[mask] / denom[mask]
        return out

    def cosine_docs(self, doc_a: int, doc_b: int) -> float:
        assert self.indptr is not None and self.indices is not None and self.data is not None and self.doc_norm is not None

        na = float(self.doc_norm[doc_a])
        nb = float(self.doc_norm[doc_b])
        if na <= 0.0 or nb <= 0.0:
            return 0.0

        a0, a1 = int(self.indptr[doc_a]), int(self.indptr[doc_a + 1])
        b0, b1 = int(self.indptr[doc_b]), int(self.indptr[doc_b + 1])

        ia, ib = a0, b0
        dot = 0.0
        while ia < a1 and ib < b1:
            ta = int(self.indices[ia])
            tb = int(self.indices[ib])
            if ta == tb:
                dot += float(self.data[ia]) * float(self.data[ib])
                ia += 1
                ib += 1
            elif ta < tb:
                ia += 1
            else:
                ib += 1

        return dot / (na * nb)


class InteractionGraph:
    def __init__(self):
        self.user_likes: Dict[int, Set[str]] = defaultdict(set)
        self.item_liked_by: Dict[str, Set[int]] = defaultdict(set)
        self.item_popularity: Counter = Counter()

    def reset(self):
        self.__init__()

    def add_interaction(self, user_id: int, it_key: str, event: str):
        if event not in ("like", "purchase"):
            return
        self.user_likes[user_id].add(it_key)
        self.item_liked_by[it_key].add(user_id)
        self.item_popularity[it_key] += 1

    def get_collaborative_candidates(
        self,
        liked_item_keys: List[str],
        limit: int = 500,
        max_users_per_item: int = 200,
        max_likes_per_user: int = 400,
    ) -> Dict[str, float]:
        if not liked_item_keys:
            return {}

        liked_set = set(liked_item_keys)
        scores = Counter()

        for k in liked_item_keys:
            users = list(self.item_liked_by.get(k, ()))
            if not users:
                continue
            if len(users) > max_users_per_item:
                users = users[:max_users_per_item]

            for u in users:
                their = list(self.user_likes.get(u, ()))
                if not their:
                    continue
                if len(their) > max_likes_per_user:
                    their = their[:max_likes_per_user]

                denom_u = math.log(1.0 + len(their))
                w_user = (1.0 / denom_u) if denom_u > 0.0 else 1.0

                for cand_k in their:
                    if cand_k in liked_set:
                        continue
                    pop = self.item_popularity.get(cand_k, 0)
                    denom_pop = math.log(1.0 + pop)
                    w_item = (1.0 / denom_pop) if denom_pop > 0.0 else 1.0
                    scores[cand_k] += w_user * w_item

        if not scores:
            return {}
        return dict(scores.most_common(limit))


@dataclass
class Store:
    items_db: Dict[str, Dict]
    content: SparseTfidfIndex
    collab: InteractionGraph
    index_ready: bool = False


store = Store(
    items_db={},
    content=SparseTfidfIndex(),
    collab=InteractionGraph(),
    index_ready=False,
)


def rebuild_index():
    store.content.reset()
    store.content.add_items(list(store.items_db.values()))
    store.index_ready = True


def _human_reason(liked_genres: Set[str], liked_title_tokens: Set[str], cand_doc_id: int, has_cf: bool, score: float) -> str:
    if score <= 0.0:
        return "Propozycja eksploracyjna (mało sygnału)"

    cgenres = set(store.content.doc_genres[cand_doc_id])
    overlap_g = list(liked_genres & cgenres)[:3]

    ctoks = set(store.content.tokenize(store.content.doc_title[cand_doc_id]))
    overlap_t = list(liked_title_tokens & ctoks)[:2]

    parts = []
    if has_cf:
        parts.append("Użytkownicy o podobnych gustach też to lubią")
    if overlap_g:
        parts.append("Wspólne gatunki: " + ", ".join(overlap_g))
    if overlap_t:
        parts.append("Podobne słowa w tytule: " + ", ".join(overlap_t))
    if not parts:
        parts.append("Pasuje do Twojego profilu (gatunki/tematy)")
    return " • ".join(parts)


def _mmr_select(candidate_doc_ids: np.ndarray, relevance: np.ndarray, limit: int, diversity: float) -> np.ndarray:
    if candidate_doc_ids.size == 0 or limit <= 0:
        return np.zeros(0, dtype=np.int32)

    lam = 1.0 - float(diversity)
    selected: List[int] = []
    remaining = candidate_doc_ids.tolist()
    rel_map = {int(d): float(r) for d, r in zip(candidate_doc_ids.tolist(), relevance.tolist())}

    while len(selected) < limit and remaining:
        best_doc = None
        best_score = -1e18

        for d in remaining:
            rel = rel_map.get(int(d), 0.0)
            if not selected:
                mmr = rel
            else:
                max_sim = 0.0
                for s in selected:
                    sim = store.content.cosine_docs(int(d), int(s))
                    if sim > max_sim:
                        max_sim = sim
                mmr = (lam * rel) - ((1.0 - lam) * max_sim)

            if mmr > best_score:
                best_score = mmr
                best_doc = int(d)

        if best_doc is None:
            break
        selected.append(best_doc)
        remaining.remove(best_doc)

    return np.array(selected, dtype=np.int32)


def _popular_fallback(limit: int, target_type: Optional[str]) -> List[RecommendedItem]:
    out: List[RecommendedItem] = []

    for k, _cnt in store.collab.item_popularity.most_common(limit * 10):
        d = store.items_db.get(k)
        if not d:
            continue
        if target_type and d.get("type") != target_type:
            continue
        out.append(
            RecommendedItem(
                id=int(d["id"]),
                type=d["type"],
                title=d["title"],
                genres=d["genres"],
                score=0.1,
                reason="Popular choice",
            )
        )
        if len(out) >= limit:
            return out

    for d in store.items_db.values():
        if target_type and d.get("type") != target_type:
            continue
        out.append(
            RecommendedItem(
                id=int(d["id"]),
                type=d["type"],
                title=d["title"],
                genres=d["genres"],
                score=0.05,
                reason="Discover something new",
            )
        )
        if len(out) >= limit:
            break
    return out


@app.get("/health")
def health():
    return {
        "ok": True,
        "items_loaded": len(store.items_db),
        "index_ready": store.index_ready,
    }


@app.post("/sync/items")
def sync_items(req: ItemsSyncRequest):
    if req.full_replace:
        store.items_db.clear()

    for it in req.items:
        k = item_key(it.type, it.id)
        store.items_db[k] = {"id": it.id, "type": it.type, "title": it.title, "genres": it.genres}

    rebuild_index()

    return {"status": "ok", "items_total": len(store.items_db), "index_ready": store.index_ready}


@app.post("/sync/interactions")
def sync_interactions(req: InteractionsSyncRequest):
    if not store.index_ready:
        raise HTTPException(status_code=409, detail="Index not built. Sync items first.")

    ingested = 0
    skipped = 0

    for inter in req.interactions:
        k = item_key(inter.item_type, inter.item_id)
        if k not in store.items_db:
            skipped += 1
            continue
        store.collab.add_interaction(inter.user_id, k, inter.event)
        ingested += 1

    return {"status": "ok", "ingested": ingested, "skipped": skipped}


@app.post("/rebuild")
def rebuild():
    rebuild_index()
    return {"status": "ok", "items_total": len(store.items_db), "index_ready": store.index_ready}


@app.post("/recommend", response_model=Dict[str, List[RecommendedItem]])
def recommend(req: RecommendRequest):
    if not store.index_ready or len(store.items_db) == 0:
        return {"items": []}

    liked_keys: List[str] = []
    liked_doc_ids: List[int] = []

    for it in req.liked_items:
        k = item_key(it.type, it.id)
        doc_id = store.content.key_to_doc.get(k)
        if doc_id is not None:
            liked_keys.append(k)
            liked_doc_ids.append(int(doc_id))

    if not liked_doc_ids:
        return {"items": _popular_fallback(req.limit, req.target_type)}

    profile, profile_norm, user_title_tokens, user_genres = store.content.build_user_profile(liked_doc_ids)
    expanded_genres = store.content.expand_genres(user_genres, top_k_per_genre=2)

    content_candidates = store.content.collect_candidates(
        query_tokens=user_title_tokens,
        genres=expanded_genres,
        max_candidates=req.max_candidates,
    )

    cf_scores = store.collab.get_collaborative_candidates(liked_keys, limit=2000)

    cand_set = set(content_candidates.tolist())
    for k, sc in cf_scores.items():
        doc_id = store.content.key_to_doc.get(k)
        if doc_id is None:
            continue
        if k in set(liked_keys):
            continue
        cand_set.add(int(doc_id))

    cand_set -= set(liked_doc_ids)

    if req.target_type:
        cand_set = {d for d in cand_set if store.content.doc_type[int(d)] == req.target_type}

    if not cand_set:
        return {"items": _popular_fallback(req.limit, req.target_type)}

    candidate_doc_ids = np.array(list(cand_set), dtype=np.int32)
    c_scores = store.content.score_candidates_content(profile, profile_norm, candidate_doc_ids)

    final_scores = np.zeros(candidate_doc_ids.size, dtype=np.float32)
    has_cf_flag = np.zeros(candidate_doc_ids.size, dtype=np.bool_)

    for i, d in enumerate(candidate_doc_ids.tolist()):
        k = store.content.doc_key[int(d)]
        cf_raw = float(cf_scores.get(k, 0.0))
        cf_score = cf_raw / (1.0 + cf_raw)
        if cf_score > 0.0:
            final_scores[i] = (0.45 * float(c_scores[i])) + (0.55 * cf_score)
            has_cf_flag[i] = True
        else:
            final_scores[i] = float(c_scores[i])
            has_cf_flag[i] = False

    keep = final_scores >= float(req.min_score)
    candidate_doc_ids = candidate_doc_ids[keep]
    final_scores = final_scores[keep]
    has_cf_flag = has_cf_flag[keep]

    if candidate_doc_ids.size == 0:
        return {"items": _popular_fallback(req.limit, req.target_type)}

    pool = int(min(req.mmr_pool, candidate_doc_ids.size))
    if pool < req.limit:
        pool = int(min(candidate_doc_ids.size, max(req.limit, pool)))

    top_idx = np.argpartition(-final_scores, kth=pool - 1)[:pool]
    pool_doc_ids = candidate_doc_ids[top_idx]
    pool_scores = final_scores[top_idx]
    pool_has_cf = has_cf_flag[top_idx]

    order = np.argsort(-pool_scores)
    pool_doc_ids = pool_doc_ids[order]
    pool_scores = pool_scores[order]
    pool_has_cf = pool_has_cf[order]

    picked_doc_ids = _mmr_select(pool_doc_ids, pool_scores, req.limit, req.diversity)

    score_map: Dict[int, Tuple[float, bool]] = {
        int(d): (float(s), bool(cf))
        for d, s, cf in zip(pool_doc_ids.tolist(), pool_scores.tolist(), pool_has_cf.tolist())
    }

    out: List[RecommendedItem] = []
    for d in picked_doc_ids.tolist():
        dkey = store.content.doc_key[int(d)]
        item = store.items_db.get(dkey)
        if not item:
            continue
        sc, cf = score_map.get(int(d), (0.0, False))
        out.append(
            RecommendedItem(
                id=int(item["id"]),
                type=item["type"],
                title=item["title"],
                genres=item["genres"],
                score=round(sc, 4),
                reason=_human_reason(user_genres, user_title_tokens, int(d), cf, sc),
            )
        )

    if len(out) < req.limit:
        need = req.limit - len(out)
        out.extend(_popular_fallback(need, req.target_type))

    return {"items": out}


@app.post("/feedback")
def record_feedback(req: InteractionPayload):
    if not store.index_ready:
        raise HTTPException(status_code=409, detail="Index not built. Sync items first.")

    k = item_key(req.item_type, req.item_id)
    if k not in store.items_db:
        raise HTTPException(status_code=404, detail="Item not found")

    store.collab.add_interaction(req.user_id, k, req.event)
    return {"status": "recorded", "new_popularity": int(store.collab.item_popularity[k])}


@app.get("/stats")
def stats():
    return {
        "total_items": len(store.content.doc_key),
        "vocab_size": int(len(store.content.vocab)),
        "genres_indexed": int(len(store.content.genre_docs)),
        "total_interactions_recorded": int(sum(len(v) for v in store.collab.user_likes.values())),
        "popular_items_tracked": int(len(store.collab.item_popularity)),
        "index_ready": store.index_ready,
    }