import requests

from shared.config.settings import PRODUCT_API


def fetch_product_by_id(product_id: str) -> dict:
    """Fetch a product data by its ID from the product API."""
    response = requests.get(f"{PRODUCT_API}/{product_id}")
    response.raise_for_status()
    product_data = response.json()
    return product_data
