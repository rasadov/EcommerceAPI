import sys

from loguru import logger

from entry.grpc_server import serve
from entry.kafka_consumer import start_kafka_consumer
from recommendations.train import train_and_save


def main() -> None:
    if len(sys.argv) < 2:
        logger.error("Usage: python main.py [consumer|server|train]")
        sys.exit(1)

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
