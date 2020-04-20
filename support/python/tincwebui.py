import requests






class TincWebUIError(RuntimeError):
    def __init__(self, method: str, code: int, message: str, data: Any):
        super().__init__('{}: {}: {} - {}'.format(method, code, message, data))
        self.code = code
        self.message = message
        self.data = data

    @staticmethod
    def from_json(method: str, payload: dict) -> 'TincWebUIError':
        return TincWebUIError(
            method=method,
            code=payload['code'],
            message=payload['message'],
            data=payload.get('data')
        )


class TincWebUIClient:
    """
    Operations with tinc-web-boot related to UI
    """

    def __init__(self, base_url: str = 'http://127.0.0.1:8686/api/', session: Optional[requests.Session] = None):
        self.__url = base_url
        self.__id = 1
        self.__session = session or requests

    def __next_id(self):
        self.__id += 1
        return self.__id

    def issue_access_token(self, valid_days: int) -> str:
        """
        Issue and sign token
        """
        response = self.__session.post(self.__url, json={
            "jsonrpc": "2.0",
            "method": "TincWebUI.IssueAccessToken",
            "id": self.__next_id(),
            "params": [valid_days, ]
        })
        assert response.ok, str(response.status_code) + " " + str(response.reason)
        payload = response.json()
        if 'error' in payload:
            raise TincWebUIError.from_json('issue_access_token', payload['error'])
        return payload['result']

    def notify(self, title: str, message: str) -> bool:
        """
        Make desktop notification if system supports it
        """
        response = self.__session.post(self.__url, json={
            "jsonrpc": "2.0",
            "method": "TincWebUI.Notify",
            "id": self.__next_id(),
            "params": [title, message, ]
        })
        assert response.ok, str(response.status_code) + " " + str(response.reason)
        payload = response.json()
        if 'error' in payload:
            raise TincWebUIError.from_json('notify', payload['error'])
        return payload['result']
