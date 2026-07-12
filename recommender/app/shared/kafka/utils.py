from shared.kafka.models import EventType


def product_is_created_or_updated(event: dict) -> bool:
    return event["type"] in (EventType.PRODUCT_CREATED, EventType.PRODUCT_UPDATED)


def product_is_deleted(event: dict) -> bool:
    return event["type"] == EventType.PRODUCT_DELETED
