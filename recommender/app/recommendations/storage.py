import io
import json
from dataclasses import dataclass, field
from datetime import datetime, timezone

import numpy as np

from recommendations.weights import (
    ALS_FILE,
    METADATA_FILE,
    POPULAR_PRODUCTS_FILE,
    SIMILAR_PRODUCTS_FILE,
)
from shared.file_manager.base import FileManagerBase


@dataclass
class ALSData:
    user_ids: list[str]
    product_ids: list[str]
    user_factors: np.ndarray
    item_factors: np.ndarray


@dataclass
class TrainedModel:
    als: ALSData | None
    similar_products: dict[str, list[tuple[str, float]]]
    popular_products: list[str]
    metadata: dict = field(default_factory=dict)


def save_model(files: FileManagerBase, model: TrainedModel) -> None:
    if model.als is not None:
        buffer = io.BytesIO()
        np.savez(
            buffer,
            user_ids=np.array(model.als.user_ids, dtype=object),
            product_ids=np.array(model.als.product_ids, dtype=object),
            user_factors=model.als.user_factors,
            item_factors=model.als.item_factors,
        )
        files.upload_file(ALS_FILE, buffer.getvalue())

    buffer = io.BytesIO()
    product_ids = np.array(list(model.similar_products.keys()), dtype=object)
    neighbor_lists = np.array(
        [json.dumps(neighbors) for neighbors in model.similar_products.values()],
        dtype=object,
    )
    np.savez(buffer, product_ids=product_ids, neighbor_lists=neighbor_lists)
    files.upload_file(SIMILAR_PRODUCTS_FILE, buffer.getvalue())

    files.upload_file(
        POPULAR_PRODUCTS_FILE,
        json.dumps(model.popular_products).encode("utf-8"),
    )

    metadata = {**model.metadata, "saved_at": datetime.now(timezone.utc).isoformat()}
    files.upload_file(METADATA_FILE, json.dumps(metadata, indent=2).encode("utf-8"))


def has_model(files: FileManagerBase) -> bool:
    return files.file_exists(POPULAR_PRODUCTS_FILE)


def load_model(files: FileManagerBase) -> TrainedModel:
    als = _load_als(files) if files.file_exists(ALS_FILE) else None
    return TrainedModel(
        als=als,
        similar_products=_load_similar_products(files),
        popular_products=_load_popular_products(files),
        metadata=_load_metadata(files),
    )


def _load_als(files: FileManagerBase) -> ALSData:
    payload = np.load(io.BytesIO(files.download_file(ALS_FILE)), allow_pickle=True)
    return ALSData(
        user_ids=payload["user_ids"].tolist(),
        product_ids=payload["product_ids"].tolist(),
        user_factors=payload["user_factors"],
        item_factors=payload["item_factors"],
    )


def _load_similar_products(files: FileManagerBase) -> dict[str, list[tuple[str, float]]]:
    if not files.file_exists(SIMILAR_PRODUCTS_FILE):
        return {}

    payload = np.load(
        io.BytesIO(files.download_file(SIMILAR_PRODUCTS_FILE)),
        allow_pickle=True,
    )
    product_ids = payload["product_ids"].tolist()
    neighbor_lists = payload["neighbor_lists"].tolist()
    return {
        product_id: [tuple(neighbor) for neighbor in json.loads(neighbors)]
        for product_id, neighbors in zip(product_ids, neighbor_lists)
    }


def _load_popular_products(files: FileManagerBase) -> list[str]:
    if not files.file_exists(POPULAR_PRODUCTS_FILE):
        return []
    return json.loads(files.download_file(POPULAR_PRODUCTS_FILE).decode("utf-8"))


def _load_metadata(files: FileManagerBase) -> dict:
    if not files.file_exists(METADATA_FILE):
        return {}
    return json.loads(files.download_file(METADATA_FILE).decode("utf-8"))
