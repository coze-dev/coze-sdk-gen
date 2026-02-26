# OAuth JWT 授权（开发者）
纯后端应用通常不包含任何用户界面元素，用户无法直接与之交互。该场景下建议使用 **JWT**（JSON Web Token）模式，扣子账号直接永久授予 OAuth App 权限，OAuth App 随时可以通过后端应用签发的 JWT 获取访问令牌，以后端应用的身份请求扣子 API。
OAuth 授权码授权、PKCE 授权、设备授权模式下，扣子编程支持多人协作场景下跨账号的 OAuth 授权，详细说明可参考[OAuth 授权（多人协作场景）](https://docs.coze.cn/api/open/docs/developer_guides/oauth_collaborate)。

# 示例项目源码
使用 JWT 进行鉴权的步骤较为复杂，扣子编程提供了 Python 等多语言 SDK，能够有效简化鉴权流程，支持扣子编程的各种鉴权模式。推荐使用扣子 SDK 实现鉴权。以下是基于扣子 SDK 的 OAuth JWT 授权的示例代码，供参考：
|  **Python**   |  **Node.js** |  **Java**  |  **Go** |
| --- | --- | --- | --- |
| [examples/auth_oauth_jwt.py](https://github.com/coze-dev/coze-py/blob/main/examples/auth_oauth_jwt.py) | [auth/auth-oauth-jwt.ts](https://github.com/coze-dev/coze-js/blob/main/examples/coze-js-node/src/auth/auth-oauth-jwt.ts) <br> [auth/auth-oauth-jwt-channel.ts](https://github.com/coze-dev/coze-js/blob/main/examples/coze-js-node/src/auth/auth-oauth-jwt-channel.ts) | [JWTOauthExample.java](https://github.com/coze-dev/coze-java/blob/main/example/src/main/java/example/auth/JWTOAuthExample.java) | [jwt_example.go](https://github.com/coze-dev/coze-go/blob/main/examples/auth/jwt_oauth/main.go) |
# 授权流程
授权流程如下：

1. 在扣子编程创建 OAuth 应用并完成授权。
2. 应用程序通过公钥和私钥签署 JWT。
3. 应用程序通过 API，获取访问令牌，具体请参见[通过 JWT 获取 Oauth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt#a458f4b1) 。
4. 应用程序根据访问令牌调用扣子 API。

详细步骤如下：
## 1 创建 Oauth 应用并授权
你需要在扣子编程中创建 OAuth 应用。 
* 在扣子企业版（企业标准版、企业旗舰版）中，仅**组织超级管理员**和**管理员**有权限创建、编辑、删除 OAuth 应用，以及对应用进行授权操作。
* 扣子企业旗舰版多组织场景中，创建的 OAuth 应用默认只有本组织资源的访问权限。如果要访问其他组织的资源，可以将 OAuth 应用的授权链接分享给目标组织的超级管理员或管理员，对方确认授权后才能访问，具体操作请参见[2 分享授权链接并授权](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt_collaborate#0bd747e7)。


1. 登录[扣子编程](https://code.coze.cn/home)。
2. 在左侧导航栏选择 **API & SDK。**
3. 在顶部单击**授权**  > **OAuth 应用**页签。
4. 在 OAuth 应用页面右上角单击**创建新应用**，填写应用的基本信息，并单击**创建并继续**。
   | **配置** | **说明** |
   | --- | --- |
   | 应用类型  | 应用的类型，此处应指定为**普通**。 |
   | 客户端类型 | 客户端类型，此处设置为**服务类应用**。 |
   | 应用名称  | 应用的名称，在扣子编程中全局唯一。  |
   | 描述  | 应用的基本描述信息。  |
5. 填写应用的配置信息。 
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/e39056a1768a4508824c53b9575cff9e~tplv-goo7wpa0wc-image.image)
   | **配置** | **说明** |
   | --- | --- |
   | 公钥和私钥  | 用于应用程序客户端身份认证的非对称密钥。  <br> 单击**创建 Key**，页面将自动创建一对公钥和私钥，公钥自动配置在扣子中，私钥以 `private_key.pem` 文件格式由网页自动下载至本地 **Downloads** 目录下。支持创建最多三对公钥和私钥。 <br> * 建议将 `private_key.pem` 文件安全地存储在只有您的应用可以访问的位置。 <br> * 扣子编程使用符合行业标准的 WebCrypto 加密标准，在浏览器前端创建非对称密钥，密钥强度符合行业标准。扣子编程任何时候都不会上传私钥到后端。请您放心使用。 <br>  |
   | 权限  | 应用程序调用扣子 API 时需要的权限范围。不同层级权限的生效范围请参见[权限层级](https://docs.coze.cn/api/open/docs/developer_guides/authentication#fc1e1372)。 <br> 此处配置旨在于划定应用的权限范围，并未完成授权操作。创建 OAuth 应用后还需要参考后续操作完成授权。 <br>  |
6. 单击**确定**，完成配置。 
7. 在弹出的对话框中，确认权限等信息无误后，单击**授权**。
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/fcf288a64b704109b3a6306a12b732e3~tplv-goo7wpa0wc-image.image)

## 2 签署 JWT
扣子编程的 JWT 生成方式及部分参数定义沿用业界统一流程规范，但 JWT 中 Header 和 Payload 部分由扣子编程自行定义。
在 [JWT](https://jwt.io/)（JSON Web Tokens）的流程中，通常使用私钥来签署（sign）token。JWT 包含三部分，即 Header、Payload 和 signature，其中 header 和 payload 由参数拼接而成，signature 根据指定的签名算法和私钥对 Header 和 Payload 自动计算生成。三部分之间用点（.）分隔。详细的签署方式可参考[JWT 官方文档](https://jwt.io/)。
### Header 和 Payload
扣子编程对 Header 和 Payload 的定义如下：

* Header
   Header 部分的参数定义如下：
   | **参数** | **类型** | **是否必选** | **说明** |
   | --- | --- | --- | --- |
   | alg | String | 必选 | 签名使用的加密算法，固定为 `RS256`。 <br> `RS256` 是一种基于非对称加密的签名算法，它结合了 RSA 加密和 SHA-256 哈希算法。 |
   | typ | String | 必选 | 固定为 JWT。 |
   | kid | String | 必选 | OAuth 应用的**公钥指纹**，可以在[OAuth 应用](https://www.coze.cn/open/oauth/apps)页面找到这个应用，在**操作**列单击编辑图标，进入**配置**页面查看公钥指纹。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/c2842dca7947408685dacc5a8509c6e9~tplv-goo7wpa0wc-image.image) |
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
   | session_name | String | 可选 | 访问令牌的会话标识。目前仅限在会话隔离场景下使用，即将 session_name 指定为用户在业务侧的 UID，以此区分不同业务侧用户的对话历史。 <br> 若未指定 session_name，不同用户的对话历史可能会掺杂在一起。 <br> 会话隔离的详细实现方法请参见[如何实现会话隔离](https://docs.coze.cn/api/open/docs/developer_guides/session_isolation)。 <br>  |
   | session_context | Object | 可选  | 会话上下文信息，包含设备相关信息等。 <br>  |
   | session_context.device_info | Object | 可选  | 用于配置设备相关信息，扣子编程基于该部分信息对设备做用量管控以及账单记录。 <br> 仅扣子企业旗舰版支持该参数。硬件设备用量管控的具体操作可参考[终端用户用量查询和配额管控](https://docs.coze.cn/api/open/docs/dev_how_to_guides/device_usage)。 <br>  |
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
       "session_name": "user_2222",  //用户在业务侧的 UID
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

## 3 获取访问令牌
应用程序调用 [通过 JWT 获取 Oauth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt#a458f4b1) API ，请求 Header 中携带 JWT，扣子服务端会在响应中通过  access_token 字段返回访问令牌。
请求示例如下：

<div type="doc-tabs">
<div type="tab-item" title="扣子个人版和企业标准版" key="M1RxTpWwWv">

```Shell
curl --location --request POST 'https://api.coze.cn/api/permission/oauth2/token' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIyRE****.eyJpc3MiOiIzMTAwMDAwMDAwMDIiLCJhdWQiOiJhcGkuY296ZS5jb20iLCJpYXQiOjE1MTYyMzkwMjIsImV4cCI6MTkxNjI1OTAyMiwianRpIjoiZmhqaGFsc2tqZmFkc2pld3F****.CuoiCCF-nHFyGmu2EKlwFoyd3uDyKQ3Drc1CrXQyMVySTzZlZd2M7zKWsziB3AktwbUZiRJlQ1HbghR05CW2YRHwKL4-dlJ4koR3onU7iQAO5DkPCaIxbAuXfcrb4YPpkw3GhmuPShenNLObfSWS0CqIyakXL8qD5AgXLoB-SejAsRdzloSUInNXENJHfSVMkThxRhJy7yEjX3BmculC54fMKENRfLElBqwJyLLUjeRHsYnaru2ca4W8_yaPJ7F****' \
--data '{
    "duration_seconds": 86399,
    "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer"
}'
```


</div>
<div type="tab-item" title="扣子企业旗舰版多组织" key="gclnQFKfYy">

```Shell
curl --location --request POST 'https://api.coze.cn/api/permission/oauth2/account/7490888144456***/token' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InZZd2ZsdFR1OWZBbWtwWFhSdnR5UmREc3RONVMzZWNFcDFqVzB6dVQyRE****.eyJpc3MiOiIzMTAwMDAwMDAwMDIiLCJhdWQiOiJhcGkuY296ZS5jb20iLCJpYXQiOjE1MTYyMzkwMjIsImV4cCI6MTkxNjI1OTAyMiwianRpIjoiZmhqaGFsc2tqZmFkc2pld3F****.CuoiCCF-nHFyGmu2EKlwFoyd3uDyKQ3Drc1CrXQyMVySTzZlZd2M7zKWsziB3AktwbUZiRJlQ1HbghR05CW2YRHwKL4-dlJ4koR3onU7iQAO5DkPCaIxbAuTsQobtCAdkkZTg8gav9EnN1QN_1xq0w8BzuuhS7wCeY8UbaskkTK9GnO4eU9tEINmVw-2CrfB-kNbEHlEDwXfcrb4YPpkw3GhmuPShenNLObfSWS0CqIyakXL8qD5AgXLoB-SejAsRdzloSUInNXENJHfSVMkThxRhJy7yEjX3BmculC54fMKENRfLElBqwJyLLUjeRHsYnaru2ca4W8_yaPJ7F****' \
--data '{
    "duration_seconds": 86399,
    "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer"
}'
```


</div>
</div>
返回示例如下：
```JSON
{
    "access_token": "czs_RQOhsc7vmUzK4bNgb7hn4wqOgRBYAO6xvpFHNbnl6RiQJX3cSXSguIhFDzgy****",
    "expires_in": 1721135859
}
```

## 4 发起扣子 API 请求
在 API 请求头中通过 `Authorization=Bearer `*`$Access_Token`​*指定访问令牌，发起扣子 API 请求。每个接口对应的权限点不同。
以[查看智能体配置](https://docs.coze.cn/api/open/docs/developer_guides/get_metadata_draft_published) API 为例，完整的 API 请求如下：
```Shell
curl --location --request GET 'https://api.coze.cn/v1/bots/73428668*****?is_published=true' \ 
--header 'Authorization: Bearer czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****' \
```

# 通过 JWT 获取 Oauth Access Token
通过 JWT 获取 Oauth Access Token。
* JWT 仅能使用一次，如需再次申请 OAuth Access Token，必须重新生成一个 JWT。
* OAuth Access Token 的有效期默认为 15 分钟，不支持刷新。如需获取新的 Access Token，你需要再次生成一个 JWT，并调用此接口。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | * 扣子个人版和企业标准版：`https://api.coze.cn/api/permission/oauth2/token` <br> * 扣子企业旗舰版多组织：`https://api.coze.cn/api/permission/oauth2/account/{account_id}/token` |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Content-Type  | application/json | 请求正文的方式。 |
| Authorization | Bearer <span style="color: #D83931"><strong><em>$JWT</em></strong></span> | 使用应用的客户端私钥签署的 JWT。生成方式可参考[2 签署 JWT](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt#abe9ef05)。 |
### Path
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| account_id | String | 可选 | 在扣子企业旗舰版中，如果需要在指定组织下获取 Token，需要在 Path 中填写组织 ID。组织 ID 的获取方式如下： <br> 组织超级管理员或管理员可以在**组织设置**页面查看对应的组织 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/0aa86f3e0957493d82b3f534a2800fcc~tplv-goo7wpa0wc-image.image) |
### Body
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| grant_type | String | 必选 | 固定为 `urn:ietf:params:oauth:grant-type:jwt-bearer`。 |
| duration_seconds | Integer | 可选 | 申请的 AccessToken 有效期，单位为秒，默认 900 秒，即 15 分钟。最大可设置为 86399 秒，即 24 小时。 |
## 返回参数
| **参数** | **类型** | **说明** |
| --- | --- | --- |
| access_token | String | OAuth Access Token。 |
| expires_in | Integer | OAuth Access Token 的过期时间，Unixtime 时间戳格式，单位为秒。 |
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


