from surprise import SVD, Dataset, Reader
from app.db.session import ReplicaSession
from shared.repo import (
    fetch_interactions,
    get_all_product_ids,
    get_interacted_ids_for_user,
    get_interacted_ids_for_viewed
)

class Recommender:
    def __init__(self):
        self.model = SVD(n_factors=50, random_state=42)
        self.trainset = None
        self.product_ids = set()

    def train(self):
        df = fetch_interactions()
        self.product_ids = set(df["product_id"].unique())
        reader = Reader(rating_scale=(1, 3))
        data = Dataset.load_from_df(df[["user_id", "product_id", "rating"]], reader)
        self.trainset = data.build_full_trainset()
        self.model.fit(self.trainset)

    def _predict_and_sort(self, user_id: str, candidates: list[str]) -> list:
        """
        Predict ratings for each candidate and return a list
        of (product_id, est_prediction), sorted descending by est_prediction.
        """
        predictions = [self.model.predict(user_id, pid) for pid in candidates]
        sorted_predictions = sorted(predictions, key=lambda x: x.est, reverse=True)
        return sorted_predictions

    def recommend_on_user_id(self, user_id: str, skip: int = 0, take: int = 5) -> list[str]:
        """Recommend based on user interactions."""
        with ReplicaSession() as session:
            all_product_ids = get_all_product_ids(session)
            interacted_ids = get_interacted_ids_for_user(session, user_id)

        candidates = [pid for pid in all_product_ids if pid not in interacted_ids]

        sorted_predictions = self._predict_and_sort(user_id, candidates)

        sliced = sorted_predictions[skip : skip + take]

        return [pred.iid for pred in sliced]

    def recommend_on_viewed_ids(self, viewed_ids: list[str], skip: int = 0, take: int = 5) -> list[str]:
        """Recommend based on a set of viewed product IDs."""
        with ReplicaSession() as session:
            all_product_ids = get_all_product_ids(session)
            interacted_ids = get_interacted_ids_for_viewed(session, viewed_ids)

        candidates = [pid for pid in all_product_ids if pid not in interacted_ids]

        sorted_predictions = self._predict_and_sort("anonymous_user", candidates)

        sliced = sorted_predictions[skip : skip + take]

        return [pred.iid for pred in sliced]


recommender = Recommender()  # Singleton for simplicity
