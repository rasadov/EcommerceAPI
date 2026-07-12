from pathlib import Path

from shared.file_manager.base import FileManagerBase


class LocalFileManager(FileManagerBase):
    def __init__(self, base_path: str | Path):
        self._base_path = Path(base_path).resolve()
        self._base_path.mkdir(parents=True, exist_ok=True)

    def _resolve_key(self, key: str) -> Path:
        normalized = Path(key)
        if normalized.is_absolute():
            raise ValueError(f"Absolute paths are not allowed: {key}")

        destination = (self._base_path / normalized).resolve()
        if not str(destination).startswith(str(self._base_path)):
            raise ValueError(f"Path traversal detected: {key}")

        return destination

    def upload_file(self, key: str, file: bytes) -> str:
        destination = self._resolve_key(key)
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_bytes(file)
        return key

    def download_file(self, file_path: str) -> bytes:
        destination = self._resolve_key(file_path)
        if not destination.is_file():
            raise FileNotFoundError(f"File not found: {file_path}")
        return destination.read_bytes()

    def delete_file(self, file_path: str) -> None:
        destination = self._resolve_key(file_path)
        if destination.is_file():
            destination.unlink()

    def file_exists(self, file_path: str) -> bool:
        destination = self._resolve_key(file_path)
        return destination.is_file()
