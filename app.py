from typing import List, Optional, Literal
from fastapi import FastAPI
from pydantic import BaseModel
 
app = FastAPI()
 
ItemType = Literal["movie", "book", "concert"]
 
class Item(BaseModel):
    id: int
    type: ItemType
 
class RecommendRequest(BaseModel):
    user_id: Optional[int] = None
    liked_items: List[Item]
    limit: int = 10
 
class RecommendedItem(Item):
    score: float
 
class RecommendResponse(BaseModel):
    items: List[RecommendedItem]

class FeedbackRequest(BaseModel):
  user_id: int
  item_id: int
  item_type: ItemType
  event: str # 'like', 'dislike'
  score_shown: Optional[float] = None

@app.get("/ml/health")
def health():
  return {"status":"ok", "service":"ml", "version":"0.1.0"}

@app.post("/ml/recommend", response_model=RecommendResponse)
def recommend(req: RecommendRequest):
    dummy = [
        RecommendedItem(id=101, type="movie", score=0.95),
        RecommendedItem(id=202, type="book", score=0.89),
        RecommendedItem(id=303, type="concert", score=0.88)
    ]
    return RecommendResponse(items=dummy[:req.limit])

@app.post("/ml/feedback")
def feedback(req: FeedbackRequest):
  print("Feedback:", req.dict())
  return {"status": "logged"}
 
