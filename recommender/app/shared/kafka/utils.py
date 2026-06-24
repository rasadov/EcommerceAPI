from recommender.app.shared.kafka.models import EventType


def product_is_created_or_updated(event: dict) -> bool:
    return event["type"] == EventType.PRODUCT_CREATED or event["type"] == EventType.PRODUCT_UPDATED

def product_is_deleted(event: dict) -> bool:
    return event["type"] == EventType.PRODUCT_DELETED

def is_interaction_event(event: dict) -> bool:
    return event["type"] == EventType.INTERACTION_CREATED or event["type"] == EventType.INTERACTION_UPDATED