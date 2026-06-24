from collections import defaultdict
from datetime import datetime, timezone

import implicit
import pandas as pd
from loguru import logger
from scipy.sparse import coo_matrix

from recommendations.storage import ALSData, TrainedModel, save_model
from shared.db.repo import fetch_interactions, get_all_product_ids
from shared.db.session import get_session
from shared.file_manager.factory import get_file_manager


def train_and_save() -> None:
    interactions = fetch_interactions()
    model = _train(interactions)

    if not model.popular_products:
        with get_session() as session:
            model.popular_products = sorted(get_all_product_ids(session))

    save_model(get_file_manager(), model)
    logger.info("Model saved: {}", model.metadata.get("metrics", {}))


def _train(interactions: pd.DataFrame) -> TrainedModel:
    if interactions.empty:
        logger.warning("No interactions found; saving empty model")
        return TrainedModel(
            als=None,
            similar_products={},
            popular_products=[],
            metadata=_metadata(interactions, {}),
        )

    als = _train_als(interactions)
    similar_products = _train_similar_products(interactions)
    popular_products = _train_popular_products(interactions)
    metrics = {
        "interactions": len(interactions),
        "users": int(interactions["user_id"].nunique()),
        "products": int(interactions["product_id"].nunique()),
        "purchases": int((interactions["interaction_type"] == "purchase").sum()),
    }

    return TrainedModel(
        als=als,
        similar_products=similar_products,
        popular_products=popular_products,
        metadata=_metadata(interactions, metrics),
    )


def _train_als(interactions: pd.DataFrame) -> ALSData | None:
    aggregated = interactions.groupby(["user_id", "product_id"], as_index=False)["confidence"].sum()
    if aggregated.empty:
        return None

    user_ids = aggregated["user_id"].unique().tolist()
    product_ids = aggregated["product_id"].unique().tolist()
    user_index = {user_id: i for i, user_id in enumerate(user_ids)}
    product_index = {product_id: i for i, product_id in enumerate(product_ids)}

    matrix = coo_matrix(
        (
            aggregated["confidence"].astype(float).tolist(),
            (
                [user_index[u] for u in aggregated["user_id"]],
                [product_index[p] for p in aggregated["product_id"]],
            ),
        ),
        shape=(len(user_ids), len(product_ids)),
    ).tocsr()

    model = implicit.als.AlternatingLeastSquares(factors=50, random_state=42, iterations=15)
    model.fit(matrix)

    return ALSData(user_ids, product_ids, model.user_factors, model.item_factors)


def _train_similar_products(
    interactions: pd.DataFrame,
    top_k: int = 50,
) -> dict[str, list[tuple[str, float]]]:
    pairs: dict[str, dict[str, float]] = defaultdict(lambda: defaultdict(float))

    for _, group in interactions.groupby("user_id"):
        items = group.drop_duplicates("product_id")[["product_id", "confidence"]].to_dict("records")
        for i, left in enumerate(items):
            for right in items[i + 1 :]:
                weight = min(left["confidence"], right["confidence"])
                pairs[left["product_id"]][right["product_id"]] += weight
                pairs[right["product_id"]][left["product_id"]] += weight

    return {
        product_id: sorted(scores.items(), key=lambda x: x[1], reverse=True)[:top_k]
        for product_id, scores in pairs.items()
    }


def _train_popular_products(interactions: pd.DataFrame) -> list[str]:
    purchases = interactions[interactions["interaction_type"] == "purchase"]
    source = purchases if not purchases.empty else interactions
    return (
        source.groupby("product_id")["confidence"]
        .sum()
        .sort_values(ascending=False)
        .index.tolist()
    )


def _metadata(interactions: pd.DataFrame, metrics: dict) -> dict:
    return {"trained_at": datetime.now(timezone.utc).isoformat(), "metrics": metrics}
