from loguru import logger
from sqlalchemy.exc import IntegrityError, OperationalError

from shared.db.models import Base
from shared.db.session import replica_engine


def ensure_schema() -> None:
    """Create tables if missing. Tolerates concurrent startup across containers."""
    try:
        Base.metadata.create_all(replica_engine, checkfirst=True)
    except (IntegrityError, OperationalError) as exc:
        logger.info("Database schema already initialized by another process")
        logger.debug("Schema init detail: {}", exc)
        replica_engine.dispose()
