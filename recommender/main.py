import os
import sys

APP_DIR = os.path.join(os.path.dirname(__file__), "app")
if APP_DIR not in sys.path:
    sys.path.insert(0, APP_DIR)

from loguru import logger

from entry.grpc_server import serve
from entry.kafka_consumer import start_kafka_consumer
from recommendations.train import train_and_save
from shared.db.schema import ensure_schema


def main() -> None:
    if len(sys.argv) < 2:
        logger.error("Usage: python main.py [consumer|server|train]")
        sys.exit(1)

    ensure_schema()

    command = sys.argv[1]
    if command == "consumer":
        start_kafka_consumer()
    elif command == "server":
        serve()
    elif command == "train":
        train_and_save()
    else:
        logger.error("Unknown command: {}", command)
        sys.exit(1)


if __name__ == "__main__":
    main()
