import requests


class Bridge:
    def __init__(self, api_url: str):
        self.api_url = api_url.rstrip("/")
        self.session = requests.Session()

    def get_services(self) -> list:
        return []

    def get_active_incidents(self) -> list:
        return []

    def get_service_incidents(self, service_id: str, limit: int = 20) -> list:
        _ = (service_id, limit)
        return []

    def get_service_results(self, service_id: str, limit: int = 50) -> list:
        _ = (service_id, limit)
        return []
