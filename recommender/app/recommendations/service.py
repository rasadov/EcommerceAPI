from recommendations.rankers import PopularProductsRanker, SimilarProductsRanker, UserRanker
from recommendations.storage import TrainedModel
from shared.db.repo import get_interacted_ids_for_user
from shared.db.session import get_session


class RecommendationService:
    """Main API: recommend products for a user or for a browsing session."""

    def __init__(self, model: TrainedModel):
        self._user_ranker = (
            UserRanker(
                model.als.user_ids,
                model.als.product_ids,
                model.als.user_factors,
                model.als.item_factors,
            )
            if model.als
            else None
        )
        self._similar_ranker = SimilarProductsRanker(model.similar_products)
        self._popular_ranker = PopularProductsRanker(model.popular_products)

    def for_user(self, user_id: str, skip: int = 0, take: int = 5) -> list[str]:
        exclude = _products_user_already_saw(user_id)
        limit = skip + take

        if self._user_ranker and self._user_ranker.knows_user(user_id):
            ids = self._user_ranker.recommend(user_id, exclude=exclude, limit=limit)
        else:
            ids = self._popular_ranker.recommend(exclude=exclude, limit=limit)

        return ids[skip : skip + take]

    def for_viewed_products(
        self,
        product_ids: list[str],
        skip: int = 0,
        take: int = 5,
    ) -> list[str]:
        exclude = set(product_ids)
        limit = skip + take

        if product_ids and self._similar_ranker.has_data_for(product_ids):
            ids = self._similar_ranker.recommend(product_ids, exclude=exclude, limit=limit)
        else:
            ids = self._popular_ranker.recommend(exclude=exclude, limit=limit)

        return ids[skip : skip + take]


def _products_user_already_saw(user_id: str) -> set[str]:
    with get_session() as session:
        return get_interacted_ids_for_user(session, user_id)
