# TincWebMajordomo

Operations for joining public network


* [TincWebMajordomo.Join](#tincwebmajordomojoin) - Join public network if code matched. Will generate error if node subnet not matched



## TincWebMajordomo.Join

Join public network if code matched. Will generate error if node subnet not matched

* Method: `TincWebMajordomo.Join`
* Returns: `*Sharing`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | network | `string` |
| 1 | code | `string` |
| 2 | self | `*Node` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "http://127.0.0.1:8686/api/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "TincWebMajordomo.Join",
    "params" : []
}
EOF
```
### Node

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| subnet | `string` |  |
| port | `uint16` |  |
| address | `[]Address` |  |
| publicKey | `string` |  |
| version | `int` |  |
### Sharing

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| subnet | `string` |  |
| node | `[]*network.Node` |  |