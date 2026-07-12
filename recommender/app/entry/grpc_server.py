import grpc
from concurrent import futures
from loguru import logger

from generated.pb import recommender_pb2, recommender_pb2_grpc
from recommendations.load import load_service
from shared.config.settings import GRPC_PORT
from shared.db.repo import get_products_by_ids
from shared.db.session import get_session


class RecommenderServiceServicer(recommender_pb2_grpc.RecommenderServiceServicer):
    def __init__(self, service):
        self._service = service

    def GetRecommendations(self, request, context):
        try:
            product_ids = self._service.for_user(
                user_id=request.user_id,
                skip=request.skip or 0,
                take=request.take or 5,
            )
            return _to_response(product_ids)
        except Exception as exc:
            logger.exception("Failed to get recommendations for user {}", request.user_id)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(exc))
            return recommender_pb2.RecommendationResponse()

    def GetRecommendationsBasedOnViewed(self, request, context):
        try:
            product_ids = self._service.for_viewed_products(
                product_ids=list(request.ids),
                skip=request.skip or 0,
                take=request.take or 5,
            )
            return _to_response(product_ids)
        except Exception as exc:
            logger.exception("Failed to get viewed-based recommendations")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(exc))
            return recommender_pb2.RecommendationResponse()


def _to_response(product_ids: list[str]) -> recommender_pb2.RecommendationResponse:
    with get_session() as session:
        products = get_products_by_ids(session, product_ids)
    return recommender_pb2.RecommendationResponse(
        recommended_products=[p.to_grpc_model() for p in products]
    )


def serve() -> None:
    service = load_service()
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    recommender_pb2_grpc.add_RecommenderServiceServicer_to_server(
        RecommenderServiceServicer(service),
        server,
    )
    server.add_insecure_port(f"[::]:{GRPC_PORT}")
    logger.info("gRPC server started on port {}", GRPC_PORT)
    server.start()
    server.wait_for_termination()
