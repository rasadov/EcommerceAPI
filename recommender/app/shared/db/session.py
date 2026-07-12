from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from shared.config.settings import DATABASE_URL

replica_engine = create_engine(DATABASE_URL, pool_pre_ping=True)
ReplicaSession = sessionmaker(bind=replica_engine)


def get_session():
    return ReplicaSession()