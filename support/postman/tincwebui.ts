{
  "info": {
    "_postman_id": "tinc-web-boot/web/shared@TincWebUI",
    "name": "TincWebUI",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
    "description": "# TincWebUI\n\nOperations with tinc-web-boot related to UI"
  },
  "item": [
    {
      "name": "IssueAccessToken",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json",
            "type": "text"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"jsonrpc\": \"2.0\",\n  \"method\": \"TincWebUI.IssueAccessToken\",\n  \"id\": 1,\n  \"params\": {}\n}",
          "options": {
            "raw": {
              "language": "json"
            }
          }
        },
        "url": {
          "raw": "http://127.0.0.1:8686/api/",
          "protocol": "http",
          "host": [
            "127",
            "0",
            "0",
            "1"
          ],
          "port": "8686",
          "path": [
            "",
            "api",
            ""
          ]
        },
        "description": "# TincWebUI.IssueAccessToken\n\nIssue and sign token\n\n* Method: `TincWebUI.IssueAccessToken`\n* Returns: `string`\n\n* Arguments:\n\n| Position | Name | Type |\n|----------|------|------|\n| 0 | validDays | `uint` |\n\n\n"
      }
    },
    {
      "name": "Notify",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json",
            "type": "text"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"jsonrpc\": \"2.0\",\n  \"method\": \"TincWebUI.Notify\",\n  \"id\": 1,\n  \"params\": {}\n}",
          "options": {
            "raw": {
              "language": "json"
            }
          }
        },
        "url": {
          "raw": "http://127.0.0.1:8686/api/",
          "protocol": "http",
          "host": [
            "127",
            "0",
            "0",
            "1"
          ],
          "port": "8686",
          "path": [
            "",
            "api",
            ""
          ]
        },
        "description": "# TincWebUI.Notify\n\nMake desktop notification if system supports it\n\n* Method: `TincWebUI.Notify`\n* Returns: `bool`\n\n* Arguments:\n\n| Position | Name | Type |\n|----------|------|------|\n| 0 | title | `string` |\n| 1 | message | `string` |\n\n\n"
      }
    },
    {
      "name": "Endpoints",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json",
            "type": "text"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"jsonrpc\": \"2.0\",\n  \"method\": \"TincWebUI.Endpoints\",\n  \"id\": 1,\n  \"params\": {}\n}",
          "options": {
            "raw": {
              "language": "json"
            }
          }
        },
        "url": {
          "raw": "http://127.0.0.1:8686/api/",
          "protocol": "http",
          "host": [
            "127",
            "0",
            "0",
            "1"
          ],
          "port": "8686",
          "path": [
            "",
            "api",
            ""
          ]
        },
        "description": "# TincWebUI.Endpoints\n\nEndpoints list to access web UI\n\n* Method: `TincWebUI.Endpoints`\n* Returns: `[]Endpoint`\n\n### Endpoint\n\n| Json | Type | Comment |\n|------|------|---------|\n| host | `string` |  |\n| port | `uint16` |  |\n| kind | `EndpointKind` |  |\n\n"
      }
    }
  ]
}