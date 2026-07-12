import requests

from shared.config.settings import PRODUCT_API


def normalize_product_data(product_data: dict) -> dict:
    product_id = product_data.get("product_id") or product_data.get("id")
    account_id = product_data.get("accountID") or product_data.get("account_id")
    return {
        "product_id": product_id,
        "name": product_data["name"],
        "description": product_data["description"],
        "price": product_data["price"],
        "accountID": account_id,
    }


def fetch_product_by_id(product_id: str) -> dict:
    response = requests.get(f"{PRODUCT_API}/{product_id}")
    response.raise_for_status()
    return normalize_product_data(response.json())
