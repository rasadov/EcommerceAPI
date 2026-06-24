import json

from kafka import KafkaConsumer
import requests
from loguru import logger

from recommender.app.shared.product.utils import fetch_product_by_id
from shared.db.session import get_session
from shared.db.repo import (
    get_product_by_id,
    create_or_update_product,
    create_product,
    record_interaction,
    delete_product_by_id,
    )
from shared.kafka.utils import (
    product_is_created_or_updated,
    product_is_deleted,
    is_interaction_event)
from shared.config.settings import KAFKA_SERVER

def start_kafka_consumer():
    consumer = KafkaConsumer(
        "product_events",
        "interaction_events",
        bootstrap_servers=KAFKA_SERVER)

    for message in consumer:
        event = json.loads(message.value)
        with get_session() as session:
            if product_is_created_or_updated(event):
                product_data = event["data"]
                logger.info("Processing product event {} for product ID: {}", event['type'], product_data['product_id'])
                create_or_update_product(session, product_data)

            elif product_is_deleted(event):
                delete_product_by_id(session, event["data"]["product_id"])

            elif is_interaction_event(event):
                record_interaction(session, event)
                product = get_product_by_id(session, event["data"]["product_id"])
                if not product:
                    try:
                        product = fetch_product_by_id(event["data"]["product_id"])
                        create_product(session, product)
                        session.commit()
                    except requests.RequestException as e:
                        logger.error("Failed to fetch product {} for interaction event {}: {}", event["data"]["product_id"], event["type"], e)
                        session.rollback()

if __name__ == "__main__":
    start_kafka_consumer()