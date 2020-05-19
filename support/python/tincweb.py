from aiohttp import client

from dataclasses import dataclass

from enum import Enum
from typing import Any, List, Optional


class Duration(Enum):
    MIN_DURATION = -1 << 63
    MAX_DURATION = 1<<63 - 1
    MIN_DURATION = -1 << 63
    MAX_DURATION = 1<<63 - 1
    NANOSECOND = 1

    def to_json(self) -> int:
        return self.value

    @staticmethod
    def from_json(payload: int) -> 'Duration':
        return Duration(payload)



@dataclass
class Network:
    name: 'str'
    running: 'bool'
    config: 'Optional[Config]'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "running": self.running,
            "config": self.config.to_json(),
        }

    @staticmethod
    def from_json(payload: dict) -> 'Network':
        return Network(
                name=payload['name'],
                running=payload['running'],
                config=Config.from_json(payload['config']),
        )


@dataclass
class Config:
    name: 'str'
    port: 'int'
    interface: 'str'
    mode: 'str'
    mask: 'int'
    device_type: 'Optional[str]'
    device: 'Optional[str]'
    connect_to: 'Optional[List[str]]'
    broadcast: 'str'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "port": self.port,
            "interface": self.interface,
            "mode": self.mode,
            "mask": self.mask,
            "deviceType": self.device_type,
            "device": self.device,
            "connectTo": self.connect_to,
            "broadcast": self.broadcast,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Config':
        return Config(
                name=payload['name'],
                port=payload['port'],
                interface=payload['interface'],
                mode=payload['mode'],
                mask=payload['mask'],
                device_type=payload['deviceType'],
                device=payload['device'],
                connect_to=payload['connectTo'] or [],
                broadcast=payload['broadcast'],
        )


@dataclass
class PeerInfo:
    name: 'str'
    online: 'bool'
    configuration: 'Node'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "online": self.online,
            "config": self.configuration.to_json(),
        }

    @staticmethod
    def from_json(payload: dict) -> 'PeerInfo':
        return PeerInfo(
                name=payload['name'],
                online=payload['online'],
                configuration=Node.from_json(payload['config']),
        )


@dataclass
class Node:
    name: 'str'
    subnet: 'str'
    port: 'int'
    ip: 'str'
    address: 'Optional[List[Address]]'
    public_key: 'str'
    version: 'int'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "subnet": self.subnet,
            "port": self.port,
            "ip": self.ip,
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
                ip=payload['ip'],
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


@dataclass
class Upgrade:
    port: 'Optional[int]'
    address: 'Optional[List[Address]]'
    device: 'Optional[str]'

    def to_json(self) -> dict:
        return {
            "port": self.port,
            "address": [x.to_json() for x in self.address],
            "device": self.device,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Upgrade':
        return Upgrade(
                port=payload['port'],
                address=[Address.from_json(x) for x in (payload['address'] or [])],
                device=payload['device'],
        )


class TincWebError(RuntimeError):
    def __init__(self, method: str, code: int, message: str, data: Any):
        super().__init__('{}: {}: {} - {}'.format(method, code, message, data))
        self.code = code
        self.message = message
        self.data = data

    @staticmethod
    def from_json(method: str, payload: dict) -> 'TincWebError':
        return TincWebError(
            method=method,
            code=payload['code'],
            message=payload['message'],
            data=payload.get('data')
        )


class TincWebClient:
    """
    Public Tinc-Web API (json-rpc 2.0)
    """

    def __init__(self, base_url: str = 'http://127.0.0.1:8686/api/', session: Optional[client.ClientSession] = None):
        self.__url = base_url
        self.__id = 1
        self.__request = session.request if session is not None else client.request

    def __next_id(self):
        self.__id += 1
        return self.__id

    async def networks(self) -> List[Network]:
        """
        List of available networks (briefly, without config)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Networks",
            "id": self.__next_id(),
            "params": []
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('networks', payload['error'])
        return [Network.from_json(x) for x in (payload['result'] or [])]

    async def network(self, name: str) -> Network:
        """
        Detailed network info
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Network",
            "id": self.__next_id(),
            "params": [name, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('network', payload['error'])
        return Network.from_json(payload['result'])

    async def create(self, name: str, subnet: str) -> Network:
        """
        Create new network if not exists
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Create",
            "id": self.__next_id(),
            "params": [name, subnet, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('create', payload['error'])
        return Network.from_json(payload['result'])

    async def remove(self, network: str) -> bool:
        """
        Remove network (returns true if network existed)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Remove",
            "id": self.__next_id(),
            "params": [network, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('remove', payload['error'])
        return payload['result']

    async def start(self, network: str) -> Network:
        """
        Start or re-start network
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Start",
            "id": self.__next_id(),
            "params": [network, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('start', payload['error'])
        return Network.from_json(payload['result'])

    async def stop(self, network: str) -> Network:
        """
        Stop network
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Stop",
            "id": self.__next_id(),
            "params": [network, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('stop', payload['error'])
        return Network.from_json(payload['result'])

    async def peers(self, network: str) -> List[PeerInfo]:
        """
        Peers brief list in network  (briefly, without config)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Peers",
            "id": self.__next_id(),
            "params": [network, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('peers', payload['error'])
        return [PeerInfo.from_json(x) for x in (payload['result'] or [])]

    async def peer(self, network: str, name: str) -> PeerInfo:
        """
        Peer detailed info by in the network
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Peer",
            "id": self.__next_id(),
            "params": [network, name, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('peer', payload['error'])
        return PeerInfo.from_json(payload['result'])

    async def import(self, sharing: Sharing) -> Network:
        """
        Import another tinc-web network configuration file.
It means let nodes defined in config join to the network.
Return created (or used) network with full configuration
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Import",
            "id": self.__next_id(),
            "params": [sharing.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('import', payload['error'])
        return Network.from_json(payload['result'])

    async def share(self, network: str) -> Sharing:
        """
        Share network and generate configuration file.
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Share",
            "id": self.__next_id(),
            "params": [network, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('share', payload['error'])
        return Sharing.from_json(payload['result'])

    async def node(self, network: str) -> Node:
        """
        Node definition in network (aka - self node)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Node",
            "id": self.__next_id(),
            "params": [network, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('node', payload['error'])
        return Node.from_json(payload['result'])

    async def upgrade(self, network: str, update: Upgrade) -> Node:
        """
        Upgrade node parameters.
In some cases requires restart
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Upgrade",
            "id": self.__next_id(),
            "params": [network, update.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('upgrade', payload['error'])
        return Node.from_json(payload['result'])

    async def majordomo(self, network: str, lifetime: Duration) -> str:
        """
        Generate Majordomo request for easy-sharing
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Majordomo",
            "id": self.__next_id(),
            "params": [network, lifetime.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('majordomo', payload['error'])
        return payload['result']

    async def join(self, url: str, start: bool) -> Network:
        """
        Join by Majordomo Link
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "TincWeb.Join",
            "id": self.__next_id(),
            "params": [url, start, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise TincWebError.from_json('join', payload['error'])
        return Network.from_json(payload['result'])

    async def _invoke(self, request):
        return await self.__request('POST', self.__url, json=request)


class TincWebBatch:
    """
    Public Tinc-Web API (json-rpc 2.0)
    """

    def __init__(self, client: TincWebClient, size: int = 10):
        self.__id = 1
        self.__client = client
        self.__requests = []
        self.__batch = {}
        self.__batch_size = size

    def __next_id(self):
        self.__id += 1
        return self.__id

    def networks(self):
        """
        List of available networks (briefly, without config)
        """
        params = []
        method = "TincWeb.Networks"
        self.__add_request(method, params, lambda payload: [Network.from_json(x) for x in (payload or [])])

    def network(self, name: str):
        """
        Detailed network info
        """
        params = [name, ]
        method = "TincWeb.Network"
        self.__add_request(method, params, lambda payload: Network.from_json(payload))

    def create(self, name: str, subnet: str):
        """
        Create new network if not exists
        """
        params = [name, subnet, ]
        method = "TincWeb.Create"
        self.__add_request(method, params, lambda payload: Network.from_json(payload))

    def remove(self, network: str):
        """
        Remove network (returns true if network existed)
        """
        params = [network, ]
        method = "TincWeb.Remove"
        self.__add_request(method, params, lambda payload: payload)

    def start(self, network: str):
        """
        Start or re-start network
        """
        params = [network, ]
        method = "TincWeb.Start"
        self.__add_request(method, params, lambda payload: Network.from_json(payload))

    def stop(self, network: str):
        """
        Stop network
        """
        params = [network, ]
        method = "TincWeb.Stop"
        self.__add_request(method, params, lambda payload: Network.from_json(payload))

    def peers(self, network: str):
        """
        Peers brief list in network  (briefly, without config)
        """
        params = [network, ]
        method = "TincWeb.Peers"
        self.__add_request(method, params, lambda payload: [PeerInfo.from_json(x) for x in (payload or [])])

    def peer(self, network: str, name: str):
        """
        Peer detailed info by in the network
        """
        params = [network, name, ]
        method = "TincWeb.Peer"
        self.__add_request(method, params, lambda payload: PeerInfo.from_json(payload))

    def import(self, sharing: Sharing):
        """
        Import another tinc-web network configuration file.
It means let nodes defined in config join to the network.
Return created (or used) network with full configuration
        """
        params = [sharing.to_json(), ]
        method = "TincWeb.Import"
        self.__add_request(method, params, lambda payload: Network.from_json(payload))

    def share(self, network: str):
        """
        Share network and generate configuration file.
        """
        params = [network, ]
        method = "TincWeb.Share"
        self.__add_request(method, params, lambda payload: Sharing.from_json(payload))

    def node(self, network: str):
        """
        Node definition in network (aka - self node)
        """
        params = [network, ]
        method = "TincWeb.Node"
        self.__add_request(method, params, lambda payload: Node.from_json(payload))

    def upgrade(self, network: str, update: Upgrade):
        """
        Upgrade node parameters.
In some cases requires restart
        """
        params = [network, update.to_json(), ]
        method = "TincWeb.Upgrade"
        self.__add_request(method, params, lambda payload: Node.from_json(payload))

    def majordomo(self, network: str, lifetime: Duration):
        """
        Generate Majordomo request for easy-sharing
        """
        params = [network, lifetime.to_json(), ]
        method = "TincWeb.Majordomo"
        self.__add_request(method, params, lambda payload: payload)

    def join(self, url: str, start: bool):
        """
        Join by Majordomo Link
        """
        params = [url, start, ]
        method = "TincWeb.Join"
        self.__add_request(method, params, lambda payload: Network.from_json(payload))

    def __add_request(self, method: str, params, factory):
        request_id = self.__next_id()
        request = {
            "jsonrpc": "2.0",
            "method": method,
            "id": request_id,
            "params": params
        }
        self.__requests.append(request)
        self.__batch[request_id] = (request, factory)

    async def __aenter__(self):
        self.__batch = {}
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self()

    async def __call__(self) -> list:
        offset = 0
        num = len(self.__requests)
        results = []
        while offset < num:
            next_offset = offset + self.__batch_size
            batch = self.__requests[offset:min(num, next_offset)]
            offset = next_offset

            responses = await self.__post_batch(batch)
            results = results + responses

        self.__batch = {}
        self.__requests = []
        return results

    async def __post_batch(self, batch: list) -> list:
        response = await self.__client._invoke(batch)
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        results = await response.json()
        ans = []
        for payload in results:
            request, factory = self.__batch[payload['id']]
            if 'error' in payload:
                raise TincWebError.from_json(request['method'], payload['error'])
            else:
                ans.append(factory(payload['result']))
        return ans