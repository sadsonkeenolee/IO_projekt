from typing import List, Dict

Vector = list[float]

def encode_item(metadata: Dict) -> Vector:
  return [1.0,0.0,0.0]

def cosine_similarity(a: Vector, b: Vector) -> float:
  dot = sum(x*y for x,y in zip(a,b))
  na = sum(x*x for x in a) ** (0.5)
  nb = sum(y*y for y in b) ** (0.5)
  if na == 0 or nb == 0:
    return 0.0
  return dot/(na*nb)
