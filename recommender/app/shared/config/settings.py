import os
from dotenv import load_dotenv

load_dotenv()

PRODUCT_API = os.getenv("PRODUCT_API")
KAFKA_SERVER = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")

PRODUCT_EVENTS_TOPIC = os.getenv("PRODUCT_EVENTS_TOPIC", "product_events")
INTERACTION_EVENTS_TOPIC = os.getenv("INTERACTION_EVENTS_TOPIC", "interaction_events")

DATABASE_URL = os.getenv("DATABASE_URL")

ARTIFACTS_BACKEND = os.getenv("ARTIFACTS_BACKEND", "local")
ARTIFACTS_LOCAL_PATH = os.getenv("ARTIFACTS_LOCAL_PATH", "artifacts")

S3_BUCKET_NAME = os.getenv("S3_BUCKET_NAME")
S3_ENDPOINT_URL = os.getenv("S3_ENDPOINT_URL")
S3_ACCESS_KEY_ID = os.getenv("S3_ACCESS_KEY_ID")
S3_SECRET_ACCESS_KEY = os.getenv("S3_SECRET_ACCESS_KEY")
S3_REGION = os.getenv("S3_REGION", "auto")

GRPC_PORT = int(os.getenv("GRPC_PORT", "8080"))
