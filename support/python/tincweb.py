from aiohttp import client

from dataclasses import dataclass

from typing import Any, List, Optional



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
    auto_start: 'bool'
    mode: 'str'
    ip: 'str'
    mask: 'int'
    device_type: 'Optional[str]'
    device: 'Optional[str]'
    connect_to: 'Optional[List[str]]'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "port": self.port,
            "interface": self.interface,
            "autostart": self.auto_start,
            "mode": self.mode,
            "ip": self.ip,
            "mask": self.mask,
            "deviceType": self.device_type,
            "device": self.device,
            "connectTo": self.connect_to,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Config':
        return Config(
                name=payload['name'],
                port=payload['port'],
                interface=payload['interface'],
                auto_start=payload['autostart'],
                mode=payload['mode'],
                ip=payload['ip'],
                mask=payload['mask'],
                device_type=payload['deviceType'],
                device=payload['device'],
                connect_to=payload['connectTo'] or [],
        )


@dataclass
class PeerInfo:
    name: 'str'
    online: 'bool'
    status: 'Optional[Peer]'
    configuration: 'Optional[Node]'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "online": self.online,
            "status": self.status.to_json(),
            "config": self.configuration.to_json(),
        }

    @staticmethod
    def from_json(payload: dict) -> 'PeerInfo':
        return PeerInfo(
                name=payload['name'],
                online=payload['online'],
                status=Peer.from_json(payload['status']),
                configuration=Node.from_json(payload['config']),
        )


@dataclass
class Peer:
    address: 'str'
    fetched: 'bool'
    config: 'Optional[Node]'

    def to_json(self) -> dict:
        return {
            "address": self.address,
            "fetched": self.fetched,
            "config": self.config.to_json(),
        }

    @staticmethod
    def from_json(payload: dict) -> 'Peer':
        return Peer(
                address=payload['address'],
                fetched=payload['fetched'],
                config=Node.from_json(payload['config']),
        )


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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
        response = await self.__request('POST', self.__url, json={
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
