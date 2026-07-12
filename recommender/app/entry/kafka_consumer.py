import json

from kafka import KafkaConsumer
import requests
from loguru import logger

from shared.config.settings import (
    INTERACTION_EVENTS_TOPIC,
    KAFKA_SERVER,
    PRODUCT_EVENTS_TOPIC,
)
from shared.db.repo import (
    create_or_update_product,
    create_product,
    delete_product_by_id,
    get_product_by_id,
    record_interaction,
)
from shared.db.session import get_session
from shared.kafka.utils import product_is_created_or_updated, product_is_deleted
from shared.product.utils import fetch_product_by_id


def start_kafka_consumer() -> None:
    consumer = KafkaConsumer(
        PRODUCT_EVENTS_TOPIC,
        INTERACTION_EVENTS_TOPIC,
        bootstrap_servers=KAFKA_SERVER,
        group_id="recommender-sync",
    )

    for message in consumer:
        event = json.loads(message.value)

        if message.topic == PRODUCT_EVENTS_TOPIC:
            _handle_product_event(event)
        elif message.topic == INTERACTION_EVENTS_TOPIC:
            _handle_interaction_event(event)


def _handle_product_event(event: dict) -> None:
    with get_session() as session:
        if product_is_created_or_updated(event):
            product_data = event["data"]
            logger.info(
                "Processing product event {} for product ID: {}",
                event["type"],
                product_data["product_id"],
            )
            create_or_update_product(session, product_data)
        elif product_is_deleted(event):
            delete_product_by_id(session, event["data"]["product_id"])


def _handle_interaction_event(event: dict) -> None:
    with get_session() as session:
        record_interaction(session, event)
        product = get_product_by_id(session, event["data"]["product_id"])
        if not product:
            try:
                product_data = fetch_product_by_id(event["data"]["product_id"])
                create_product(session, product_data)
                session.commit()
            except requests.RequestException as exc:
                logger.error(
                    "Failed to fetch product {} for interaction event {}: {}",
                    event["data"]["product_id"],
                    event["type"],
                    exc,
                )
                session.rollback()

if __name__ == "__main__":
    start_kafka_consumer()
