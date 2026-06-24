from sqlalchemy import String, Float, Integer, DateTime, func, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import Mapped, mapped_column, relationship
from generated.pb import recommender_pb2

Base = declarative_base()

class Product(Base):
    __tablename__ = "products"
    id: Mapped[str] = mapped_column(String, primary_key=True)
    name: Mapped[str] = mapped_column(String)
    description: Mapped[str] = mapped_column(String)
    price: Mapped[float] = mapped_column(Float)
    account_id: Mapped[int] = mapped_column(Integer)

    interactions: Mapped[list["Interaction"]] = relationship(
        "Interaction", back_populates="product", cascade="all, delete-orphan")

    def to_grpc_model(self) -> recommender_pb2.ProductReplica:
        return recommender_pb2.ProductReplica(
            id=self.id,
            name=self.name,
            description=self.description,
            price=self.price
        )

class Interaction(Base):
    __tablename__ = "interactions"
    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[str] = mapped_column(String)
    product_id: Mapped[str] = mapped_column(String, ForeignKey("products.id"))
    interaction_type: Mapped[str] = mapped_column(String)
    timestamp: Mapped[DateTime] = mapped_column(DateTime, default=func.now())

    product: Mapped["Product"] = relationship(
        "Product", back_populates="interactions")
