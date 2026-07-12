from enum import Enum


class EventType(str, Enum):
    PRODUCT_CREATED = "product_created"
    PRODUCT_UPDATED = "product_updated"
    PRODUCT_DELETED = "product_deleted"
    INTERACTION_CREATED = "interaction_created"
    INTERACTION_UPDATED = "interaction_updated"
