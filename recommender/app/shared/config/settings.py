import os
from dotenv import load_dotenv

load_dotenv()

PRODUCT_API = os.getenv("PRODUCT_API")
KAFKA_SERVER = os.getenv('KAFKA_BOOTSTRAP_SERVERS', 'kafka:9092')

DATABASE_URL = os.getenv("DATABASE_URL")
