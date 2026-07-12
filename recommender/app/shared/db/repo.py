import pandas as pd

from recommendations.weights import PURCHASE_WEIGHT, VIEW_WEIGHT
from shared.db.models import Interaction, Product
from shared.db.session import ReplicaSession, replica_engine


def fetch_interactions() -> pd.DataFrame:
    with ReplicaSession() as session:
        query = session.query(Interaction)
        df = pd.read_sql(query.statement, con=replica_engine)

    df["confidence"] = df["interaction_type"].apply(
        lambda t: PURCHASE_WEIGHT if t == "purchase" else VIEW_WEIGHT
    )
    return df


def get_all_product_ids(session) -> set[str]:
    return {p.id for p in session.query(Product.id).all()}


def get_product_by_id(session, product_id: str):
    return session.query(Product).filter(Product.id == product_id).first()


def get_products_by_ids(session, product_ids: list[str]):
    if not product_ids:
        return []
    products = session.query(Product).filter(Product.id.in_(product_ids)).all()
    by_id = {product.id: product for product in products}
    return [by_id[product_id] for product_id in product_ids if product_id in by_id]


def get_interacted_ids_for_user(session, user_id: str) -> set[str]:
    return {
        i.product_id
        for i in session.query(Interaction.product_id)
        .filter(Interaction.user_id == user_id)
        .all()
    }


def record_interaction(session, event: dict) -> None:
    data = event["data"]
    user_id = data.get("user_id", data.get("accountID", data.get("account_id")))
    if user_id is None:
        raise ValueError(f"Interaction event missing user identifier: {event}")

    session.add(
        Interaction(
            user_id=str(user_id),
            product_id=data["product_id"],
            interaction_type=event["type"],
        )
    )
    session.commit()


def create_product(session, product_data: dict) -> None:
    session.add(
        Product(
            id=product_data["product_id"],
            name=product_data["name"],
            description=product_data["description"],
            price=product_data["price"],
            account_id=product_data["accountID"],
        )
    )


def update_product(session, product_data: dict) -> None:
    product = get_product_by_id(session, product_data["product_id"])
    product.name = product_data["name"]
    product.description = product_data["description"]
    product.price = product_data["price"]
    product.account_id = product_data["accountID"]


def create_or_update_product(session, product_data: dict) -> None:
    if get_product_by_id(session, product_data["product_id"]):
        update_product(session, product_data)
    else:
        create_product(session, product_data)
    session.commit()


def delete_product_by_id(session, product_id: str) -> None:
    product = get_product_by_id(session, product_id)
    if product:
        session.delete(product)
        session.commit()
