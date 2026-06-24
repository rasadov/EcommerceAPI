from shared.kafka.models import EventType


def product_is_created_or_updated(event: dict) -> bool:
    event_type = event["type"]
    return event_type in (EventType.PRODUCT_CREATED.value, EventType.PRODUCT_UPDATED.value)


def product_is_deleted(event: dict) -> bool:
    return event["type"] == EventType.PRODUCT_DELETED.value
