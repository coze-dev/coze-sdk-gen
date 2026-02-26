# OAuth JWT 授权（渠道场景）
扣子编程提供标准的生态对接方式和多种对接方案，支持网站、App、应用市场、手机等硬件厂商作为第三方渠道入驻。渠道入驻场景下，渠道侧需要通过 JWT 授权流程获取渠道的 OAuth Access Token。智能体发布到第三方渠道后，可以使用此渠道专属的 OAuth Access Token 调用扣子 OpenAPI，例如查看智能体信息、和智能体对话等。
关于发布到自定义渠道的详细说明及步骤，可参考[渠道入驻概述](https://docs.coze.cn/api/open/docs/dev_how_to_guides/channel_integration_overview)。

# 示例项目源码
使用 JWT 进行鉴权的步骤较为复杂，扣子编程提供了 Python 等多语言 SDK，能够有效简化鉴权流程，支持扣子编程的各种鉴权模式。推荐使用扣子 SDK 实现鉴权。以下是基于扣子 SDK 的 OAuth JWT 授权的示例代码，供参考：
|  **Python**   |  **Node.js** |  **Java**  |  **Go** |
| --- | --- | --- | --- |
| [examples/auth_oauth_jwt.py](https://github.com/coze-dev/coze-py/blob/main/examples/auth_oauth_jwt.py) | [auth/auth-oauth-jwt.ts](https://github.com/coze-dev/coze-js/blob/main/examples/coze-js-node/src/auth/auth-oauth-jwt.ts) <br> [auth/auth-oauth-jwt-channel.ts](https://github.com/coze-dev/coze-js/blob/main/examples/coze-js-node/src/auth/auth-oauth-jwt-channel.ts) | [JWTsOauthExample.java](https://github.com/coze-dev/coze-java/blob/main/example/src/main/java/example/auth/JWTOAuthExample.java) | [jwt_example.go](https://github.com/coze-dev/coze-go/blob/main/examples/auth/jwt_oauth/main.go) |
# 准备工作
创建一个渠道类型的 OAuth 应用，并在渠道侧实现 JWT 的授权流程、具备获取 OAuth Access Token 的能力。
在扣子企业版（企业标准版、企业旗舰版）中，仅**超级管理员**和**管理员**有权限创建、编辑、删除 OAuth 应用，以及对应用进行授权操作。


1. 创建渠道类型的 OAuth 应用。
   1. 登录[扣子编程](https://code.coze.cn/home)。
   2. 在左侧导航栏选择 **API & SDK。**
   3. 在顶部单击**授权**  > **OAuth 应用**页签。
   4. 在 OAuth 应用页面右上角单击**创建新应用**，填写应用的基本信息，并单击**创建并继续**。
      | **配置** | **说明** |
      | --- | --- |
      | 应用类型 | 应用的类型，此处应指定为**渠道**。  |
      | 客户端类型 | 客户端类型，此处设置为**服务类应用**。 |
      | 应用名称  | 应用的名称，在扣子编程中全局唯一。  |
      | 描述  | 应用的基本描述信息。  |
   5. 填写应用的配置信息。 
      | **配置** | **说明** |
      | --- | --- |
      | 公钥和私钥  | 用于应用程序客户端身份认证的非对称密钥。  <br> 单击**创建 Key**，页面将自动创建一对公钥和私钥，公钥自动配置在扣子编程中，私钥以 `private_key.pem` 文件格式由网页自动下载至本地 **Downloads** 目录下。支持创建最多三对公钥和私钥。 <br>  <br> * 建议将 `private_key.pem` 文件安全地存储在只有您的应用可以访问的位置。 <br> * 扣子编程使用符合行业标准的 WebCrypto 加密标准，在浏览器前端创建非对称密钥，密钥强度符合行业标准。扣子编程任何时候都不会上传私钥到后端。请您放心使用。 |
      | 权限  | 应用程序调用扣子 API 时需要的权限范围。不同层级权限的生效范围请参见[权限层级](https://docs.coze.cn/api/open/docs/developer_guides/authentication#fc1e1372)。 <br> 此处配置旨在于划定应用的权限范围，并未完成授权操作。创建 OAuth 应用后还需要参考后续操作完成授权。 <br>  |
   6. 单击**确定**，完成配置。 
2. 完成 OAuth 应用授权。
   授权后，此 OAuth 应用生成的 OAuth Access Token 将具备 OAuth 应用指定的操作权限，例如调用查看智能体信息等 OpenAPI。后续还需要为应用绑定自定义发布渠道，此 OAuth Access Token 将具备所有发布到这个渠道中的智能体的操作权限，例如调用查看智能体信息 API 接口，查看指定智能体的信息。
   2. 在**授权** > [OAuth 应用](https://www.coze.cn/open/oauth/apps)页面找到需要授权的应用。 
   4. 在对应的操作列单击编辑图标。 
   5. 在编辑应用页面左下角单击**授权**。 
   授权页将显示应用名称、权限列表等重要信息，请务必认真核对信息，确认无误后单击**授权**。

# 授权流程
你需要在渠道的服务端中实现 JWT 授权流程，通过 OAuth 应用的私钥签署 JWT，并获取 Oauth Access Token。扣子编程的 [JWT](https://jwt.io/) 生成方式及各部分参数定义沿用业界统一流程规范。
授权流程如下：

1. 应用程序通过公钥和私钥签署 JWT。
2. 应用程序通过 [API：通过 JWT 获取 Oauth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt_channel#a59d629b) API，获取访问令牌。
3. 应用程序根据访问令牌调用扣子 API。

详细步骤如下：
## 1 签署 JWT
扣子编程的 JWT 生成方式及部分参数定义沿用业界统一流程规范，但 JWT 中 Header 和 Payload 部分由扣子编程自行定义。
在 [JWT](https://jwt.io/)（JSON Web Tokens）的流程中，通常使用私钥来签署（sign）token。JWT 包含三部分，即 Header、Payload 和 signature，其中 header 和 payload 由参数拼接而成，signature 根据指定的签名算法和私钥对 Header 和 Payload 自动计算生成。三部分之间用点（.）分隔。详细的签署方式可参考[JWT 官方文档](https://jwt.io/)。
### Header 和 Payload
扣子编程对 Header 和 Payload 的定义如下：

* Header
   Header 部分的参数定义如下：
   | **参数** | **类型** | **是否必选** | **说明** |
   | --- | --- | --- | --- |
   | alg | String | 必选 | 签名使用的加密算法。固定为 RS256，即非对称加密算法，一种基于 RSA（非对称加密算法）+ SHA256（安全哈希函数）的签名算法，该算法使用私钥进行签名，公钥进行验证。 |
   | typ | String | 必选 | 固定为 JWT。 |
   | kid | String | 必选 | OAuth 应用的**公钥指纹**，可以在[OAuth 应用](https://www.coze.cn/open/oauth/apps)页面找到这个应用，在**操作**列单击编辑图标，进入**配置**页面查看公钥指纹。 |
   Header 示例如下：
   ```JSON
   {
     "alg": "RS256",          // 固定为RS256
     "typ": "JWT",            // 固定为 JWT
     "kid": "gdehvaDegW....." // OAuth 应用的公钥指纹
   }
   ```

* Payload：
   Payload 部分的参数定义如下：
   | **参数** | **类型** | **是否必选** | **说明** |
   | --- | --- | --- | --- |
   | iss | String | 必选 | OAuth 应用的 ID，可以在[OAuth 应用](https://www.coze.cn/open/oauth/apps)页面查看。 |
   | aud | String | 必选 | 扣子 API 的 Endpoint，即 `api.coze.cn`。 |
   | iat | Integer | 必选 | JWT 开始生效的时间，Unixtime 时间戳格式，精确到秒。一般为当前时刻。 |
   | exp | Integer | 必选 | JWT 过期的时间，Unixtime 时间戳格式，精确到秒。必须晚于 iat。 |
   | jti | String | 必选 | 随机字符串，用于防止重放攻击。建议长度大于 32 字节。每次签署 JWT 时应指定为不同的字符串。 |
   | session_name | String | 可选 | 访问令牌的会话标识。目前仅限在会话隔离场景下使用，即将 session_name 指定为用户在业务侧的 UID，以此区分不同业务侧用户的对话历史。 <br> 若未指定 session_name，不同用户的对话历史可能会掺杂在一起。 |
   | session_context | Object | 可选  | 会话上下文信息，包含设备相关信息等。 <br>  |
   | session_context.device_info | Object | 可选  | 用于配置设备相关信息，扣子编程基于该部分信息对设备做用量管控以及账单记录。 <br> 仅扣子企业旗舰版支持该参数。硬件设备用量管控的具体操作可参考[终端用户用量查询和配额管控](https://docs.coze.cn/api/open/docs/dev_how_to_guides/device_usage) <br>  |
   | session_context.device_info.device_id | String | 可选  | IoT 等硬件设备 ID，一个设备对应一个唯一的设备号。 <br> 当需要记录设备用量或对设备用量进行管控时，需要填写该参数，否则，无法对设备进行用量管控，用量统计页面对应的设备 ID 将显示为 N/A。 |
   | session_context.device_info.custom_consumer <br>  | String | 可选  | 自定义维度的实体 ID，你可以根据业务需要进行设置，例如 APP 上的用户名等。 <br> 当需要记录设备用量或对设备用量进行管控，需要填写该参数，否则，无法对设备进行用量管控，用量统计页面对应的自定义 ID 将显示为 N/A。 <br> * `device_id` 和 `custom_consumer` 建议选择其中一个即可。 <br> * `custom_consumer `参数用于设备用量管控，与对话等 API 传入的 `user_id` 无关，`user_id` 主要用于上下文、数据库隔离等场景。 <br> * 出于数据隐私及信息安全等方面的考虑，不建议使用业务系统中定义的用户敏感标识（如手机号等）作为 `custom_consumer` 的值。 <br>  |
   Payload 示例如下：
   ```JSON
   {
       "iss": "310000000002",   // OAuth 应用的 ID
       "aud": "api.coze.cn",    // 扣子 API 的 Endpoint
       "iat": 1516239022,       // JWT 开始生效的时间，秒级时间戳
       "exp": 1516259022,       // JWT 过期时间，秒级时间戳
       "jti": "fhjashjgkhalskj", // 随机字符串，防止重放攻击
       "session_context": {
           "device_info": {
               "device_id": "1234567890" // IoT 等硬件设备的唯一标识 ID
           }
       }
   }
   ```


### 示例代码
你可以直接参考以下示例代码签署 JWT。
```Python
# You must run `pip install PyJWT cryptography` to install the PyJWT and the cryptography packages in order to use this script.

#!/usr/bin/env python3
import sys
import time
import uuid

import jwt

# 替换为你的实际 Coze App 私钥
signing_key = '''
-----BEGIN PRIVATE KEY-----
xxxxxxxxxxxxxxxxxx
-----END PRIVATE KEY-----
'''

payload = {
    'iat': int(time.time()),
    'exp': int(time.time()) + 600,
    "jti": str(uuid.uuid4()),
    'aud': 'api.coze.cn',   # 替换为实际的coze api domain
    'iss': '1127900106117'  # 替换为你的实际 Coze App ID
}

headers = {
    'kid': '_v0VjcMlLdQc3tRTD3jC5Xz29TUnKQOhtuD5k-gpyf4'  # 替换为你的实际 Coze App 公钥指纹
}

# Create JWT with headers
encoded_jwt = jwt.encode(payload, signing_key, algorithm='RS256', headers=headers)

print(f"JWT: {encoded_jwt}")
```

最终生成的 JWT 示例如下：
```Plain Text
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InZZd2ZsdFR1OWZBbWtwWFhSdnR5UmREc3RONVMzZWNFcDFqVzB6dVQyRE****.eyJpc3MiOiIzMTAwMDAwMDAwMDIiLCJhdWQiOiJhcGkuY296ZS5jb20iLCJpYXQiOjE1MTYyMzkwMjIsImV4cCI6MTkxNjI1OTAyMiwianRpIjoiZmhqaGFsc2tqZmFkc2pld3F****.CuoiCCF-nHFyGmu2EKlwFoyd3uDyKQ3Drc1CrXQyMVySTzZlZd2M7zKWsziB3AktwbUZiRJlQ1HbghR05CW2YRHwKL4-dlJ4koR3onU7iQAO5DkPCaIxbAuTsQobtCAdkkZTg8gav9EnN1QN_1xq0w8BzuuhS7wCeY8UbaskkTK9GnO4eU9tEINmVw-2CrfB-kNbEHlEDwXfcrb4YPpkw3GhmuPShenNLObfSWS0CqIyakXL8qD5AgXLoB-SejAsRdzloSUInNXENJHfSVMkThxRhJy7yEjX3BmculC54fMKENRfLElBqwJyLLUjeRHsYnaru2ca4W8_yaPJ7F****
```

## 2 获取访问令牌
应用程序调用 [API：通过 JWT 获取 Oauth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt_channel#a59d629b) API ，请求 Header 中携带 JWT，扣子服务端会在响应中通过  access_token 字段返回访问令牌。
请求示例如下：
```Shell
curl --location --request POST 'https://api.coze.cn/api/permission/oauth2/token' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InZZd2ZsdFR1OWZBbWtwWFhSdnR5UmREc3RONVMzZWNFcDFqVzB6dVQyRE****.eyJpc3MiOiIzMTAwMDAwMDAwMDIiLCJhdWQiOiJhcGkuY296ZS5jb20iLCJpYXQiOjE1MTYyMzkwMjIsImV4cCI6MTkxNjI1OTAyMiwianRpIjoiZmhqaGFsc2tqZmFkc2pld3F****.CuoiCCF-nHFyGmu2EKlwFoyd3uDyKQ3Drc1CrXQyMVySTzZlZd2M7zKWsziB3AktwbUZiRJlQ1HbghR05CW2YRHwKL4-dlJ4koR3onU7iQAO5DkPCaIxbAuTsQobtCAdkkZTg8gav9EnN1QN_1xq0w8BzuuhS7wCeY8UbaskkTK9GnO4eU9tEINmVw-2CrfB-kNbEHlEDwXfcrb4YPpkw3GhmuPShenNLObfSWS0CqIyakXL8qD5AgXLoB-SejAsRdzloSUInNXENJHfSVMkThxRhJy7yEjX3BmculC54fMKENRfLElBqwJyLLUjeRHsYnaru2ca4W8_yaPJ7F****' \
--data '{
    "duration_seconds": 86399,
    "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer"
}'
```

返回示例如下：
```JSON
{
    "access_token": "czs_RQOhsc7vmUzK4bNgb7hn4wqOgRBYAO6xvpFHNbnl6RiQJX3cSXSguIhFDzgy****",
    "expires_in": 1721135859
}
```

# 后续操作
在 API 请求头中通过 `Authorization=Bearer `*`$Access_Token`​*指定访问令牌，发起扣子 API 请求。每个接口对应的权限点不同。
以[获取已发布智能体的配置（即将下线）](https://docs.coze.cn/api/open/docs/developer_guides/get_metadata) API 为例，完整的 API 请求如下：
```Shell
curl --location --request GET 'https://api.coze.cn/v1/bot/get_online_info?bot_id=123123123' \ 
--header 'Authorization: Bearer czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****' \
```

# API：通过 JWT 获取 Oauth Access Token
通过 JWT 获取 Oauth Access Token。
* JWT 仅能使用一次，如需再次申请 OAuth Access Token，必须重新生成一个 JWT。
* OAuth Access Token 的有效期默认为 15 分钟，不支持刷新。如需获取新的 Access Token，你需要再次生成一个 JWT，并调用此接口。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/api/permission/oauth2/token |
## Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Content-Type  | application/json | 请求正文的方式。 |
| Authorization | Bearer ***$JWT*** | 使用应用的客户端私钥签署的 JWT。生成方式可参考[1 签署 JWT](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt_channel#f0dce614)。 |
## Body
| **字段** | **类型** | 是否必选 | **说明** |
| --- | --- | --- | --- |
| grant_type | String | 必选 | 固定为 `urn:ietf:params:oauth:grant-type:jwt-bearer`。 |
| duration_seconds | Integer | 可选 | 申请的 AccessToken 有效期，单位为秒，默认 900 秒，即 15 分钟。最大可设置为 86399 秒，即 24 小时。 |
| scope | Scope object | 可选 | 指定渠道侧申请的 OAuth Access Token 的操作权限，例如渠道服务端使用此令牌调用 OpenAPI 时只能和智能体 A 对话，不能执行其他操作。 <br> * 仅在渠道侧调用此接口时生效。配置渠道入驻的方式可参考[渠道入驻概述](https://docs.coze.cn/api/open/docs/dev_how_to_guides/channel_integration_overview)。 <br> * 如果不指定 scope，表示渠道能调用授权的所有权限。 <br>  |
### Scope object
| **字段** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| account_permission | Object | 必选 | 指定 OAuth Access Token 的操作权限。 |
| account_permission.permission_list | Array of String | 必选 |  OAuth Access Token 的权限点列表。字符串列表格式，例如 `["Connector.botChat"]` 表示 `chat` 权限。 |
| attribute_constraint | Object | 必选 | 指定 OAuth Access Token 的权限限制。 |
| attribute_constraint.connector_bot_chat_attribute | Object | 必选 | 指定 OAuth Access Token 允许访问的智能体范围，此智能体必须已发布为 API 服务。 |
| attribute_constraint.connector_bot_chat_attribute.bot_id_list | Array of String | 必选 | 允许访问的智能体列表，例如 `["bot_id_1", "bot_id_2"] `表示 OAuth Access Token 仅对 ID 为 1 或 2 的智能体有访问权限。 |
比如，申请 AccessToken，限制只允许和 bot1聊天，scope 应指定为：
```JSON
{
    "account_permission": {
        "permission_list": ["Connector.botChat"]
    },
    "attribute_constraint": {
        "connector_bot_chat_attribute": {
            "bot_id_list": ["bot1"]
        }
    }
}
```

## 返回结果
| **字段** | **类型** | **说明** |
| --- | --- | --- |
| access_token | String | OAuth Access Token。 |
| expires_in | Integer | OAuth Access Token 的过期时间，Unixtime 时间戳格式，精度为秒。 |
## 示例 
### 请求示例 
```Shell
curl --location 'https://api.coze.cn/api/permission/oauth2/token' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InZZd2ZsdFR1OWZBbWtwWFhSdnR5UmREc3RONVMzZWNFcDFqVzB6dVQyRE****.eyJpc3MiOiIzMTAwMDAwMDAwMDIiLCJhdWQiOiJhcGkuY296ZS5jb20iLCJpYXQiOjE1MTYyMzkwMjIsImV4cCI6MTkxNjI1OTAyMiwianRpIjoiZmhqaGFsc2tqZmFkc2pld3F****.CuoiCCF-nHFyGmu2EKlwFoyd3uDyKQ3Drc1CrXQyMVySTzZlZd2M7zKWsziB3AktwbUZiRJlQ1HbghR05CW2YRHwKL4-dlJ4koR3onU7iQAO5DkPCaIxbAuTsQobtCAdkkZTg8gav9EnN1QN_1xq0w8BzuuhS7wCeY8UbaskkTK9GnO4eU9tEINmVw-2CrfB-kNbEHlEDwXfcrb4YPpkw3GhmuPShenNLObfSWS0CqIyakXL8qD5AgXLoB-SejAsRdzloSUInNXENJHfSVMkThxRhJy7yEjX3BmculC54fMKENRfLElBqwJyLLUjeRHsYnaru2ca4W8_yaPJ7F****' \
--data '{
    "duration_seconds": 86399,
    "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer"
}'
```

### 返回示例
```JSON
{
    "access_token": "czs_RQOhsc7vmUzK4bNgb7hn4wqOgRBYAO6xvpFHNbnl6RiQJX3cSXSguIhFDzgy****",
    "expires_in": 1721135859
}
```

## 错误码
| **error_code** | **error_message** | **说明** |
| --- | --- | --- |
| invalid_request | invalid request: {parameter} <br>  | * 原因：请求参数 {parameter} 错误。 <br> * 解决方案：请参考 API 文档查看参数说明。 |
| invalid_client | / | * 原因：客户端凭证（JWT Token 或者 Client Secret）无效。 <br> * 解决方案：请校验您的客户端凭证。 |
| unsupported_grant_type | not supported grant type: {grant type} | * 原因：不支持的授权类型 {grant type}。 <br> * 解决方案：请参考 API 文档指定正确的授权类型。 |
| access_deny | app: {app name} is currently deactivated by the owner | * 原因：OAuth 应用已被禁用。 <br> * 解决方案：在扣子编程中启用 OAuth 应用。 |
|  | invalid app type | * 原因：应用类型错误。 <br> * 解决方案：渠道应用暂不支持授权码模式。 |
|  | login session invalid | * 原因：登录态无效。 <br> * 解决方案：用户需要重新登录扣子编程。 |
| internal_error | Service internal error. | * 原因：服务内部错误。 <br> * 解决方案：建议稍后重试。 |


## 常见问题
### 如何处理 JWT 授权时，提示"connector app should bind connector id"？
进行 OAuth JWT 授权时，提示"connector app should bind connector id"，该报错是由于未正确绑定 OAuth 应用到相应的发布渠道，导致授权失败。你需要在创建自定义发布渠道时，绑定渠道类型的 OAuth 应用，具体请操作参见[配置渠道入驻（账号隔离）](https://docs.coze.cn/api/open/docs/dev_how_to_guides/configure_custom_channel1)。

1. 在扣子编程左下角单击头像，选择**账号设置**。
2. 在左侧导航栏选择**发布渠道**，在**企业自定义渠道管理**页面，单击目标渠道操作列中的**更多**图标，选择**编辑**。
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/6eb477f9680f49379a00dfa674e64a6d~tplv-goo7wpa0wc-image.image)





