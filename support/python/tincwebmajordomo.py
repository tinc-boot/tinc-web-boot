from aiohttp import client

from dataclasses import dataclass

from typing import Any, List, Optional



@dataclass
class Node:
    name: 'str'
    subnet: 'str'
    port: 'int'
    address: 'Optional[List[Address]]'
    public_key: 'str'
    version: 'int'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "subnet": self.subnet,
            "port": self.port,
            "address": [x.to_json() for x in self.address],
            "publicKey": self.public_key,
            "version": self.version,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Node':
        return Node(
                name=payload['name'],
                subnet=payload['subnet'],
                port=payload['port'],
                address=[Address.from_json(x) for x in (payload['address'] or [])],
                public_key=payload['publicKey'],
                version=payload['version'],
        )


@dataclass
class Address:
    host: 'str'
    port: 'Optional[int]'

    def to_json(self) -> dict:
        return {
            "host": self.host,
            "port": self.port,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Address':
        return Address(
                host=payload['host'],
                port=payload['port'],
        )


@dataclass
class Sharing:
    name: 'str'
    subnet: 'str'
    nodes: 'Optional[List[Node]]'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "subnet": self.subnet,
            "node": [x.to_json() for x in self.nodes],
        }

    @staticmethod
    def from_json(payload: dict) -> 'Sharing':
        return Sharing(
                name=payload['name'],
                subnet=payload['subnet'],
                nodes=[Node.from_json(x) for x in (payload['node'] or [])],
        )


class TincWebMajordomoError(RuntimeError):
    def __init__(self, method: str, code: int, message: str, data: Any):
        super().__init__('{}: {}: {} - {}'.format(method, code, message, data))
        self.code = code
        self.message = message
        self.data = data

    @staticmethod
    def from_json(method: str, payload: dict) -> 'TincWebMajordomoError':
        return TincWebMajordomoError(
            method=method,
            code=payload['code'],
            message=payload['message'],
            data=payload.get('data')
        )


class TincWebMajordomoClient:
    """
    Operations for joining public network
    """

    def __init__(self, base_url: str = 'http://127.0.0.1:8686/api/', session: Optional[client.ClientSession] = None):
        self.__url = base_url
        self.__id = 1
        self.__request = session.request if session is not None else client.request

    def __next_id(self):
        self.__id += 1
        return self.__id

    async def join(self, network: str, self: Node) -> Sharing:
        """
        Join public network if code matched. Will generate error if node subnet not matched
        """
        response = await self.__request('POST', self.__url, json={
            "jsonrpc": "2.0",
            "method": "TincWebMajordomo.Join",
            "id": self.__next_id(),
            "params": [network, self.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebMajordomoError.from_json('join', payload['error'])
        return Sharing.from_json(payload['result'])
