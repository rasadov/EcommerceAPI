from loguru import logger

from recommendations.service import RecommendationService
from recommendations.storage import TrainedModel, has_model, load_model
from shared.db.repo import get_all_product_ids
from shared.db.session import get_session
from shared.file_manager.factory import get_file_manager


def load_service() -> RecommendationService:
    files = get_file_manager()

    if not has_model(files):
        logger.warning("No trained model found; using product catalog as fallback")
        with get_session() as session:
            product_ids = sorted(get_all_product_ids(session))
        model = TrainedModel(
            als=None,
            similar_products={},
            popular_products=product_ids,
            metadata={"source": "fallback"},
        )
        return RecommendationService(model)

    model = load_model(files)
    logger.info("Loaded model trained at {}", model.metadata.get("trained_at", "unknown"))
    return RecommendationService(model)
