from shared.config.settings import (
    ARTIFACTS_BACKEND,
    ARTIFACTS_LOCAL_PATH,
    S3_ACCESS_KEY_ID,
    S3_BUCKET_NAME,
    S3_ENDPOINT_URL,
    S3_REGION,
    S3_SECRET_ACCESS_KEY,
)
from shared.file_manager.base import FileManagerBase
from shared.file_manager.bucket import BucketFileManager
from shared.file_manager.local import LocalFileManager


def get_file_manager() -> FileManagerBase:
    if ARTIFACTS_BACKEND == "bucket":
        if not S3_BUCKET_NAME:
            raise ValueError("S3_BUCKET_NAME is required when ARTIFACTS_BACKEND=bucket")
        return BucketFileManager(
            bucket_name=S3_BUCKET_NAME,
            endpoint_url=S3_ENDPOINT_URL,
            access_key_id=S3_ACCESS_KEY_ID,
            secret_access_key=S3_SECRET_ACCESS_KEY,
            region_name=S3_REGION,
        )

    return LocalFileManager(base_path=ARTIFACTS_LOCAL_PATH)
