import grpc
from concurrent import futures

from generated.pb import recommender_pb2, recommender_pb2_grpc
from app.services.recommender import recommender
from recommender.app.shared.db.repo import get_products_by_ids
from recommender.app.shared.db.session import get_session
from recommender.app.shared.product.utils import fetch_product_by_id


def _handle_exception(context, error_message):
    """Helper method to set gRPC status code and details."""
    context.set_code(grpc.StatusCode.INTERNAL)
    context.set_details(error_message)


class RecommenderServiceServicer(recommender_pb2_grpc.RecommenderServiceServicer):

    def GetRecommendations(self, request, context):
        user_id = request.user_id
        skip = request.skip or 0  # default to 0 if not set
        take = request.take or 5  # default to 5 if not set

        try:
            # Get recommended product IDs
            recommended_product_ids = recommender.recommend_on_user_id(
                user_id=user_id,
                skip=skip,
                take=take
            )

            # Fetch product details as gRPC objects
            recommended_products = []
            with get_session() as session:
                products = get_products_by_ids(session, recommended_product_ids)
                recommended_products = [product.to_grpc_model() for product in products]

            return recommender_pb2.RecommendationResponse(
                recommended_products=recommended_products
            )

        except Exception as e:
            _handle_exception(context, f"Failed to get recommendations: {str(e)}")
            return recommender_pb2.RecommendationResponse()

    def GetRecommendationsBasedOnViewed(self, request, context):
        viewed_product_ids = request.viewed_product_ids
        skip = request.skip or 0
        take = request.take or 5

        try:
            recommended_product_ids = recommender.recommend_on_viewed_ids(
                viewed_ids=viewed_product_ids,
                skip=skip,
                take=take
            )

            # Fetch product details as gRPC objects
            recommended_products = []
            with get_session() as session:
                products = get_products_by_ids(session, recommended_product_ids)
                recommended_products = [product.to_grpc_model() for product in products]
            return recommender_pb2.RecommendationResponse(
                recommended_products=recommended_products
            )

        except Exception as e:
            _handle_exception(context, f"Failed to get recommendations: {str(e)}")
            return recommender_pb2.RecommendationResponse()

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    recommender_pb2_grpc.add_RecommenderServiceServicer_to_server(
        RecommenderServiceServicer(), server
    )
    server.add_insecure_port('[::]:50051')
    print("gRPC server started on port 50051")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()