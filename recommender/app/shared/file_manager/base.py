from abc import ABC, abstractmethod


class FileManagerBase(ABC):
    @abstractmethod
    def upload_file(self, key: str, file: bytes) -> str:
        """
        Upload a file under the given key.
        Returns the key/path where the file was stored.
        """
        pass

    @abstractmethod
    def download_file(self, file_path: str) -> bytes:
        """
        Download a file by key/path.
        """
        pass

    @abstractmethod
    def delete_file(self, file_path: str) -> None:
        """
        Delete a file by key/path.
        """
        pass

    @abstractmethod
    def file_exists(self, file_path: str) -> bool:
        """
        Return True if a file exists at the given key/path.
        """
        pass
