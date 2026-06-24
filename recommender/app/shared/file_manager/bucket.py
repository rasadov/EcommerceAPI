import boto3
from botocore.exceptions import ClientError

from shared.file_manager.base import FileManagerBase


class BucketFileManager(FileManagerBase):
    def __init__(
        self,
        bucket_name: str,
        *,
        endpoint_url: str | None = None,
        access_key_id: str | None = None,
        secret_access_key: str | None = None,
        region_name: str | None = None,
    ):
        if not bucket_name:
            raise ValueError("bucket_name is required")

        self._bucket_name = bucket_name
        session_kwargs: dict = {}
        if access_key_id and secret_access_key:
            session_kwargs["aws_access_key_id"] = access_key_id
            session_kwargs["aws_secret_access_key"] = secret_access_key
        if region_name:
            session_kwargs["region_name"] = region_name

        session = boto3.session.Session(**session_kwargs)
        client_kwargs: dict = {}
        if endpoint_url:
            client_kwargs["endpoint_url"] = endpoint_url

        self._client = session.client("s3", **client_kwargs)

    def upload_file(self, key: str, file: bytes) -> str:
        self._client.put_object(Bucket=self._bucket_name, Key=key, Body=file)
        return key

    def download_file(self, file_path: str) -> bytes:
        try:
            response = self._client.get_object(Bucket=self._bucket_name, Key=file_path)
        except ClientError as exc:
            error_code = exc.response.get("Error", {}).get("Code")
            if error_code in {"NoSuchKey", "404", "NotFound"}:
                raise FileNotFoundError(f"File not found: {file_path}") from exc
            raise
        return response["Body"].read()

    def delete_file(self, file_path: str) -> None:
        self._client.delete_object(Bucket=self._bucket_name, Key=file_path)

    def file_exists(self, file_path: str) -> bool:
        try:
            self._client.head_object(Bucket=self._bucket_name, Key=file_path)
        except ClientError as exc:
            error_code = exc.response.get("Error", {}).get("Code")
            if error_code in {"404", "NoSuchKey", "NotFound"}:
                return False
            raise
        return True
