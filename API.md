# TincWeb

Public Tinc-Web API (json-rpc 2.0)


* [TincWeb.Networks](#tincwebnetworks) - List of available networks (briefly, without config)
* [TincWeb.Network](#tincwebnetwork) - Detailed network info
* [TincWeb.Create](#tincwebcreate) - Create new network if not exists
* [TincWeb.Remove](#tincwebremove) - Remove network (returns true if network existed)
* [TincWeb.Start](#tincwebstart) - Start or re-start network
* [TincWeb.Stop](#tincwebstop) - Stop network
* [TincWeb.Peers](#tincwebpeers) - Peers brief list in network  (briefly, without config)
* [TincWeb.Peer](#tincwebpeer) - Peer detailed info by in the network



## TincWeb.Networks

List of available networks (briefly, without config)

* Method: `TincWeb.Networks`
* Returns: `[]*Network`

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWeb.Networks",
    "params" : []
}
EOF
```
### Network

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| running | `bool` |  |
| config | `*network.Config` |  |

## TincWeb.Network

Detailed network info

* Method: `TincWeb.Network`
* Returns: `*Network`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | name | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWeb.Network",
    "params" : []
}
EOF
```
### Network

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| running | `bool` |  |
| config | `*network.Config` |  |

## TincWeb.Create

Create new network if not exists

* Method: `TincWeb.Create`
* Returns: `*Network`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | name | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWeb.Create",
    "params" : []
}
EOF
```
### Network

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| running | `bool` |  |
| config | `*network.Config` |  |

## TincWeb.Remove

Remove network (returns true if network existed)

* Method: `TincWeb.Remove`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | network | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWeb.Remove",
    "params" : []
}
EOF
```

## TincWeb.Start

Start or re-start network

* Method: `TincWeb.Start`
* Returns: `*Network`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | network | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWeb.Start",
    "params" : []
}
EOF
```
### Network

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| running | `bool` |  |
| config | `*network.Config` |  |

## TincWeb.Stop

Stop network

* Method: `TincWeb.Stop`
* Returns: `*Network`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | network | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWeb.Stop",
    "params" : []
}
EOF
```
### Network

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| running | `bool` |  |
| config | `*network.Config` |  |

## TincWeb.Peers

Peers brief list in network  (briefly, without config)

* Method: `TincWeb.Peers`
* Returns: `[]*PeerInfo`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | network | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWeb.Peers",
    "params" : []
}
EOF
```
### PeerInfo

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| online | `bool` |  |
| status | `*tincd.Peer` |  |
| config | `*network.Node` |  |

## TincWeb.Peer

Peer detailed info by in the network

* Method: `TincWeb.Peer`
* Returns: `*PeerInfo`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | network | `string` |
| 1 | name | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWeb.Peer",
    "params" : []
}
EOF
```
### PeerInfo

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| online | `bool` |  |
| status | `*tincd.Peer` |  |
| config | `*network.Node` |  |