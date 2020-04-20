# TincWebUI

Operations with tinc-web-boot related to UI


* [TincWebUI.IssueAccessToken](#tincwebuiissueaccesstoken) - Issue and sign token
* [TincWebUI.Notify](#tincwebuinotify) - Make desktop notification if system supports it



## TincWebUI.IssueAccessToken

Issue and sign token

* Method: `TincWebUI.IssueAccessToken`
* Returns: `string`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | validDays | `uint` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWebUI.IssueAccessToken",
    "params" : []
}
EOF
```

## TincWebUI.Notify

Make desktop notification if system supports it

* Method: `TincWebUI.Notify`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | title | `string` |
| 1 | message | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWebUI.Notify",
    "params" : []
}
EOF
```