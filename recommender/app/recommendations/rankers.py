from collections import defaultdict

import numpy as np


class UserRanker:
    """Personalized recommendations for a known user (ALS)."""

    def __init__(
        self,
        user_ids: list[str],
        product_ids: list[str],
        user_factors: np.ndarray,
        item_factors: np.ndarray,
    ):
        self._user_index = {user_id: i for i, user_id in enumerate(user_ids)}
        self._product_ids = product_ids
        self._user_factors = user_factors
        self._item_factors = item_factors

    def knows_user(self, user_id: str) -> bool:
        return user_id in self._user_index

    def recommend(self, user_id: str, *, exclude: set[str], limit: int) -> list[str]:
        if not self.knows_user(user_id):
            return []

        scores = self._user_factors[self._user_index[user_id]] @ self._item_factors.T
        ranked = np.argsort(scores)[::-1]

        results: list[str] = []
        for index in ranked:
            product_id = self._product_ids[index]
            if product_id in exclude:
                continue
            results.append(product_id)
            if len(results) >= limit:
                break
        return results


class SimilarProductsRanker:
    """Recommend products similar to a set of viewed products."""

    def __init__(self, similar_products: dict[str, list[tuple[str, float]]]):
        self._similar_products = similar_products

    def has_data_for(self, product_ids: list[str]) -> bool:
        return any(pid in self._similar_products for pid in product_ids)

    def recommend(self, viewed_ids: list[str], *, exclude: set[str], limit: int) -> list[str]:
        scores: dict[str, float] = defaultdict(float)

        for product_id in viewed_ids:
            for neighbor_id, score in self._similar_products.get(product_id, []):
                if neighbor_id in exclude or neighbor_id in viewed_ids:
                    continue
                scores[neighbor_id] += score

        ranked = sorted(scores.items(), key=lambda item: item[1], reverse=True)
        return [product_id for product_id, _ in ranked[:limit]]


class PopularProductsRanker:
    """Fallback: return best-selling products."""

    def __init__(self, product_ids: list[str]):
        self._product_ids = product_ids

    def recommend(self, *, exclude: set[str], limit: int) -> list[str]:
        results: list[str] = []
        for product_id in self._product_ids:
            if product_id in exclude:
                continue
            results.append(product_id)
            if len(results) >= limit:
                break
        return results
