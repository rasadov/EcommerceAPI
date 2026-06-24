import pandas as pd

from shared.db.session import ReplicaSession
from shared.db.models import Interaction, Product

def fetch_interactions() -> pd.DataFrame:
    with ReplicaSession() as session:
        interactions = session.query(Interaction).all()
        data = [
            {
                "user_id": i.user_id,
                "product_id": i.product_id,
                "rating": 3.0 if i.interaction_type == "purchase" else 1.0
            }
            for i in interactions
        ]
        return pd.DataFrame(data)

def get_all_product_ids(session):
    """Fetch all product IDs from the database."""
    return {p.id for p in session.query(Product.id).all()}

def get_product_by_id(session, product_id: str):
    return session.query(Product).filter(Product.id == product_id).first()

def get_products_by_ids(session, product_ids: list[str]):
    return session.query(Product).filter(Product.id.in_(product_ids)).all()

def get_interacted_ids_for_user(session, user_id: str):
    """Get set of product IDs that the user has interacted with."""
    return {
        i.product_id
        for i in session.query(Interaction.product_id)
                      .filter(Interaction.user_id == user_id)
                      .all()
    }

def get_interacted_ids_for_viewed(session, viewed_ids: list[str]):
    """Get set of product IDs among 'viewed_ids' that have existing interactions."""
    return {
        i.product_id
        for i in session.query(Interaction.product_id)
                      .filter(Interaction.product_id.in_(viewed_ids))
                      .all()
    }

def record_interaction(session, event: dict):
    interaction = Interaction(
        user_id=event["data"]["user_id"],
        product_id=event["data"]["product_id"],
        interaction_type=event["type"]
    )
    session.add(interaction)

def create_product(session, product_data: dict):
    product = Product(
        id=product_data["product_id"],
        name=product_data["name"],
        description=product_data["description"],
        price=product_data["price"],
        account_id=product_data["accountID"]
    )
    session.add(product)

def update_product(session, product_data: dict):
    product = get_product_by_id(session, product_data["product_id"])
    product.name = product_data["name"]
    product.description = product_data["description"]
    product.price = product_data["price"]
    product.account_id = product_data["account_id"]

def create_or_update_product(session, product_data: dict):
    product = get_product_by_id(session, product_data["product_id"])
    if product:
        update_product(session, product_data)
    else:
        create_product(session, product_data)
    session.commit()


def delete_product_by_id(session, product_id: str):
    product = get_product_by_id(session, product_id)
    if product:
        session.delete(product)
        session.commit()