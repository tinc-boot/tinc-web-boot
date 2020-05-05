from aiohttp import client

from dataclasses import dataclass

from enum import Enum
from typing import List, Optional


class EndpointKind(Enum):
    LOCAL = "local"
    PUBLIC = "public"

    def to_json(self) -> str:
        return self.value

    @staticmethod
    def from_json(payload: str) -> 'EndpointKind':
        return EndpointKind(payload)



@dataclass
class Endpoint:
    host: 'str'
    port: 'int'
    kind: 'EndpointKind'

    def to_json(self) -> dict:
        return {
            "host": self.host,
            "port": self.port,
            "kind": self.kind.to_json(),
        }

    @staticmethod
    def from_json(payload: dict) -> 'Endpoint':
        return Endpoint(
                host=payload['host'],
                port=payload['port'],
                kind=EndpointKind.from_json(payload['kind']),
        )


@dataclass
class Config:
    binding: 'str'

    def to_json(self) -> dict:
        return {
            "binding": self.binding,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Config':
        return Config(
                binding=payload['binding'],
        )


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

    def __init__(self, base_url: str = 'http://127.0.0.1:8686/api/', session: Optional[client.ClientSession] = None):
        self.__url = base_url
        self.__id = 1
        self.__request = session.request if session is not None else client.request

    def __next_id(self):
        self.__id += 1
        return self.__id

    async def issue_access_token(self, valid_days: int) -> str:
        """
        Issue and sign token
        """
        response = await self.__request('POST', self.__url, json={
            "jsonrpc": "2.0",
            "method": "TincWebUI.IssueAccessToken",
            "id": self.__next_id(),
            "params": [valid_days, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebUIError.from_json('issue_access_token', payload['error'])
        return payload['result']

    async def notify(self, title: str, message: str) -> bool:
        """
        Make desktop notification if system supports it
        """
        response = await self.__request('POST', self.__url, json={
            "jsonrpc": "2.0",
            "method": "TincWebUI.Notify",
            "id": self.__next_id(),
            "params": [title, message, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebUIError.from_json('notify', payload['error'])
        return payload['result']

    async def endpoints(self) -> List[Endpoint]:
        """
        Endpoints list to access web UI
        """
        response = await self.__request('POST', self.__url, json={
            "jsonrpc": "2.0",
            "method": "TincWebUI.Endpoints",
            "id": self.__next_id(),
            "params": []
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebUIError.from_json('endpoints', payload['error'])
        return [Endpoint.from_json(x) for x in (payload['result'] or [])]

    async def configuration(self) -> Config:
        """
        Configuration defined for the instance
        """
        response = await self.__request('POST', self.__url, json={
            "jsonrpc": "2.0",
            "method": "TincWebUI.Configuration",
            "id": self.__next_id(),
            "params": []
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebUIError.from_json('configuration', payload['error'])
        return Config.from_json(payload['result'])
