# OAuth JWT 授权（跨账号授权场景）
在多人协作的纯后端应用场景下，当你需要通过 OAuth JWT Flow 访问其他账号下的资源时，可参考本文完成 OAuth JWT 授权。授权完成后，OAuth 应用将被永久授予指定工作空间下所有资源的访问权限，并随时可以通过后端应用签发的 JWT 获取访问令牌，以后端应用的身份请求扣子 API。
在 OAuth 授权码授权、PKCE 授权、设备授权模式下，扣子编程也支持多人协作场景下跨账号的 OAuth 授权。详细说明可参考[OAuth 授权（多人协作场景）](https://docs.coze.cn/api/open/docs/developer_guides/oauth_collaborate)。

## 授权场景
### **扣子个人版**
JWT Access Token 默认仅具备本账号资源的访问权限，因此在多人跨账号协作场景中，还需完成跨账号授权。首先需要在本账号中创建 OAuth 应用，并将 OAuth 应用的授权链接分享给目标工作空间的所有者，由其确认待授予的权限范围。然后在调用 [通过 JWT 获取 Oauth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt_collaborate#a46421b1) 时，务必在 URL 中指定目标空间所有者的 UID。
例如扣子用户 A 需要访问某个工作空间中其他用户创建的资源，则可以将授权链接分享给工作空间所有者 B，所有者 B 选择工作空间并完成授权后，用户 A 可以访问指定工作空间中的所有资源。
### **扣子企业版**
JWT Access Token 默认仅具备本组织资源的访问权限。如果要访问其他组织的资源，首先需要在本组织中创建 OAuth 应用,并将 OAuth 应用的授权链接分享给目标组织的超级管理员或管理员，由其确认待授予的权限范围。
## 示例项目源码
扣子编程提供 Python 等多语言 SDK，支持扣子编程的各种鉴权模式，你可以参考下表中的示例代码实现 OAuth JWT 授权流程。
|  **Python**   |  **Node.js** |  **Java**  |  **Go** |
| --- | --- | --- | --- |
| [examples/auth_oauth_jwt.py](https://github.com/coze-dev/coze-py/blob/main/examples/auth_oauth_jwt.py) | [auth/auth-oauth-jwt.ts](https://github.com/coze-dev/coze-js/blob/main/examples/coze-js-node/src/auth/auth-oauth-jwt.ts) <br> [auth/auth-oauth-jwt-channel.ts](https://github.com/coze-dev/coze-js/blob/main/examples/coze-js-node/src/auth/auth-oauth-jwt-channel.ts) | [JWTsOauthExample.java](https://github.com/coze-dev/coze-java/blob/main/example/src/main/java/example/auth/JWTOAuthExample.java) | [jwt_example.go](https://github.com/coze-dev/coze-go/blob/main/examples/auth/jwt_oauth/main.go) |
## 授权流程
授权流程如下：

1. 在扣子编程创建 OAuth 应用并授权。
2. 分享授权链接给目标对象，由其完成授权。
3. 应用程序通过公钥和私钥签署 JWT。
4. 通过 Token API 获取访问令牌。
5. 应用程序根据访问令牌调用扣子 API。

## 1 创建 OAuth 应用并授权
开发者需要在扣子编程中创建 OAuth 应用，并为其配置相应的权限。 

1. 创建 OAuth 应用。
   1. 登录[扣子编程](https://code.coze.cn/home)。
   2. 在左侧导航栏选择 **API & SDK。**
   3. 在顶部单击**授权**  > **OAuth 应用**页签。
   4. 在 OAuth 应用页面右上角单击**创建新应用**，填写应用的基本信息，并单击**创建并继续**。
      | **配置** | **说明** |
      | --- | --- |
      | 应用类型  | 应用的类型，此处应指定为**普通**。 |
      | 客户端类型 | 客户端类型，此处设置为**服务类应用**。 |
      | 应用名称  | 应用的名称，在扣子编程中全局唯一。  |
      | 应用描述  | 应用的基本描述信息。  |
   5. 填写应用的配置信息。 
      | **配置** | **说明** |
      | --- | --- |
      | 公钥和私钥  | 用于应用程序客户端身份认证的非对称密钥。  <br> 单击**创建 Key**，页面将自动创建一对公钥和私钥，公钥自动配置在扣子编程中，私钥以 `private_key.pem` 文件格式由网页自动下载至本地。支持创建最多三对公钥和私钥。 <br>  <br> * 建议将 `private_key.pem` 文件安全地存储在只有你的应用可以访问的位置。 <br> * 扣子编程使用符合行业标准的 WebCrypto 加密标准，在浏览器前端创建非对称密钥，密钥强度符合行业标准。扣子编程任何时候都不会上传私钥到后端。请你放心使用。 |
      | 权限  | 应用程序调用扣子 API 时需要的权限范围。不同层级权限的生效范围请参见[权限层级](https://docs.coze.cn/api/open/docs/developer_guides/authentication#fc1e1372)。 <br> 此处配置旨在于划定应用的权限范围，并未完成授权操作。创建 OAuth 应用后还需要参考后续操作完成授权。 <br>  |
   6. 单击**确定**，完成配置。 
2. 完成 OAuth 应用授权。
   在弹出的**安装并授权**对话框中，请务必认真核对应用名称、权限列表等重要信息，确认无误后单击**授权**。
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/4be3cdb276e2482e861028e8da853e86~tplv-goo7wpa0wc-image.image)

## 2 分享授权链接并授权
将 OAuth 应用的授权链接分享给目标工作空间所有者，由其确认此 OAuth 应用访问自己资源的权限范围。

1. 在 [OAuth 应用](https://www.coze.cn/open/oauth/apps)页面，找到你已创建的 OAuth 应用，单击**分享应用的授权链接**图标，复制授权链接。
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/d03f5718dce249d38e375446b95a083d~tplv-goo7wpa0wc-image.image)
2. 将授权链接分享给目标对象，此处假设为用户 B。
   * 个人版：需分享给目标工作空间所有者。
   * 企业版：需分享给目标组织的超级管理员或管理员。
3. 用户 B 打开分享链接，选择授权的组织和空间范围，并确认权限范围后，单击**授权**。
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/e1a42f15446b429c9f6f0ffc439f98db~tplv-goo7wpa0wc-image.image)

## 3 签署 JWT
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
   Payload 示例如下：
   ```JSON
   {
       "iss": "310000000002",   // OAuth 应用的 ID
       "aud": "api.coze.cn",   //扣子 API 的Endpoint
       "iat": 1516239022,       // JWT开始生效的时间，秒级时间戳
       "exp": 1516259022,       // JWT过期时间，秒级时间戳
       "jti": "fhjashjgkhalskj" // 随机字符串，防止重放攻击
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

## 4 获取访问令牌
应用程序调用 Token API 时，请求 Header 中将携带 JWT，扣子服务端会在响应中通过  access_token 字段返回访问令牌。Token API 的详细说明，请参考[通过 JWT 获取 Oauth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt_collaborate#a46421b1)。
跨账号授权场景中，申请 Access Token 的 URL 路径为 `/api/permission/oauth2/account/{account_id}/token`。

* 扣子个人版：`{account_id}` 为目标资源所属工作空间的所有者 UID。获取 UID 的具体步骤，请参考[获取资源所属工作空间的所有者 UID](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt_collaborate#9bb2b289)。
* 扣子企业版多组织场景：`{account_id}` 为目标资源所属组织的组织 ID。组织超级管理员或管理员可以在**组织设置**页面查看对应的组织 ID。


* 请求示例
   ```Shell
   curl --location --request POST 'https://api.coze.cn/api/permission/oauth2/account/292298289606****/token' \
   --header 'Content-Type: application/json' \
   --header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InZZd2ZsdFR1OWZBbWtwWFhSdnR5UmREc3RONVMzZWNFcDFqVzB6dVQyRE****.eyJpc3MiOiIzMTAwMDAwMDAwMDIiLCJhdWQiOiJhcGkuY296ZS5jb20iLCJpYXQiOjE1MTYyMzkwMjIsImV4cCI6MTkxNjI1OTAyMiwianRpIjoiZmhqaGFsc2tqZmFkc2pld3F****.CuoiCCF-nHFyGmu2EKlwFoyd3uDyKQ3Drc1CrXQyMVySTzZlZd2M7zKWsziB3AktwbUZiRJlQ1HbghR05CW2YRHwKL4-dlJ4koR3onU7iQAO5DkPCaIxbAuTsQobtCAdkkZTg8gav9EnN1QN_1xq0w8BzuuhS7wCeY8UbaskkTK9GnO4eU9tEINmVw-2CrfB-kNbEHlEDwXfcrb4YPpkw3GhmuPShenNLObfSWS0CqIyakXL8qD5AgXLoB-SejAsRdzloSUInNXENJHfSVMkThxRhJy7yEjX3BmculC54fMKENRfLElBqwJyLLUjeRHsYnaru2ca4W8_yaPJ7F****' \
   --data '{
       "duration_seconds": 86399,
       "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer"
   }'
   ```

* 返回示例
   ```Bash
   { 
       "access_token": "czs_RQOhsc7vmUzK4bNgb7hn4wqOgRBYAO6xvpFHNbnl6RiQJX3cSXSguIhFDzgy****", 
       "expires_in": 1721135859 
   }
   ```


## 5 发起扣子 API 请求
在 API 请求头中通过 `Authorization=Bearer `*`$Access_Token`​*指定访问令牌，发起扣子 API 请求。每个接口对应的权限点不同。
以[查看智能体配置](https://docs.coze.cn/api/open/docs/developer_guides/get_metadata_draft_published) API 为例，完整的 API 请求如下：
```Shell
curl --location --request GET 'https://api.coze.cn/v1/bots/73428668*****?is_published=true' \ 
--header 'Authorization: Bearer czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****' \
```

## 参考信息
### 获取资源所属工作空间的所有者 UID
通过 JWT 授权流程跨账号访问资源时，你需要先确认目标资源所属工作空间，再确认工作空间的所有者，并由其提供账号 UID。具体步骤如下：

1. 确认待访问资源所属工作空间。
2. 确认工作空间的所有者。
   1. 登录扣子编程。
   2. 在页面左上角选择目标工作空间，展开空间下拉列表，选择**当前空间设置**，在顶部选择**成员管理**页签，查看工作空间所有者信息。
      ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/e0cf2d2177b64634aae22ee2b8e1352b~tplv-goo7wpa0wc-image.image)
3. 工作空间所有者在扣子编程左下角单击头像，选择**账号设置**，查看账号的 UID。
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/11e6b7eeb50d42fea1da06f1177aeefc~tplv-goo7wpa0wc-image.image)

### 通过 JWT 获取 Oauth Access Token
通过 JWT 获取 OAuth Access Token。
* JWT 仅能使用一次，如需再次申请 OAuth Access Token，必须重新生成一个 JWT。
* OAuth Access Token 的有效期默认为 15 分钟，不支持刷新。如需获取新的 Access Token，你需要再次生成一个 JWT，并调用此接口。

#### 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/api/permission/oauth2/**account/{account_id}**/token |
#### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Content-Type  | application/json | 请求正文的方式。 |
| Authorization | Bearer ***$JWT*** | 使用应用的客户端私钥签署的 JWT。生成方式可参考[2 签署 JWT](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt#abe9ef05)。 |
#### Path
| 参数 | 类型 | 是否必选 | 说明 |
| --- | --- | --- | --- |
| account_id | String | 必选 | * 扣子个人版：`account_id` 为目标资源所属工作空间的所有者 UID。获取 UID 的具体步骤，请参考[获取资源所属工作空间的所有者 UID](https://docs.coze.cn/api/open/docs/developer_guides/oauth_jwt_collaborate#9bb2b289)。 <br> * 扣子企业版多组织场景：`account_id` 为目标资源所属组织的组织 ID。你可以在**组织管理** > **组织设置**页面查看对应的组织 ID。 <br>    ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/b57d5e293ff944d2b83efe8217d35517~tplv-goo7wpa0wc-image.image) |
#### Body
| **字段** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| grant_type | String | 必选 | 固定为 `urn:ietf:params:oauth:grant-type:jwt-bearer`。 |
| duration_seconds | Integer | 可选 | 申请的 AccessToken 有效期，单位为秒，默认 900 秒，即 15 分钟。最大可设置为 86399 秒，即 24 小时。 |
#### 返回结果
| **字段** | **类型** | **说明** |
| --- | --- | --- |
| access_token | String | OAuth Access Token。 |
| expires_in | Integer | OAuth Access Token 的过期时间，Unixtime 时间戳格式，精度为秒。 |
#### 示例 
##### 请求示例 
申请 Token，并限制 Token 只能和 bot1 对话。
```Shell
curl --location 'https://api.coze.cn/api/permission/oauth2/account/292298289606****/token' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InZZd2ZsdFR1OWZBbWtwWFhSdnR5UmREc3RONVMzZWNFcDFqVzB6dVQyRE****.eyJpc3MiOiIzMTAwMDAwMDAwMDIiLCJhdWQiOiJhcGkuY296ZS5jb20iLCJpYXQiOjE1MTYyMzkwMjIsImV4cCI6MTkxNjI1OTAyMiwianRpIjoiZmhqaGFsc2tqZmFkc2pld3F****.CuoiCCF-nHFyGmu2EKlwFoyd3uDyKQ3Drc1CrXQyMVySTzZlZd2M7zKWsziB3AktwbUZiRJlQ1HbghR05CW2YRHwKL4-dlJ4koR3onU7iQAO5DkPCaIxbAuTsQobtCAdkkZTg8gav9EnN1QN_1xq0w8BzuuhS7wCeY8UbaskkTK9GnO4eU9tEINmVw-2CrfB-kNbEHlEDwXfcrb4YPpkw3GhmuPShenNLObfSWS0CqIyakXL8qD5AgXLoB-SejAsRdzloSUInNXENJHfSVMkThxRhJy7yEjX3BmculC54fMKENRfLElBqwJyLLUjeRHsYnaru2ca4W8_yaPJ7F****' \
--data '{
    "duration_seconds": 86399,
    "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer"
}'
```

##### 返回示例
```JSON
{
    "access_token": "czs_RQOhsc7vmUzK4bNgb7hn4wqOgRBYAO6xvpFHNbnl6RiQJX3cSXSguIhFDzgy****",
    "expires_in": 1721135859
}
```

#### 错误码
| **error_code** | **error_message** | **说明** |
| --- | --- | --- |
| invalid_request | invalid request: {parameter} <br>  | * 原因：请求参数 {parameter} 错误。 <br> * 解决方案：请参考 API 文档查看参数说明。 |
| invalid_client | / | * 原因：客户端凭证（JWT Token 或者 Client Secret）无效。 <br> * 解决方案：请校验您的客户端凭证。 |
| unsupported_grant_type | not supported grant type: {grant type} | * 原因：不支持的授权类型 {grant type}。 <br> * 解决方案：请参考 API 文档指定正确的授权类型。 |
| access_deny | app: {app name} is currently deactivated by the owner | * 原因：OAuth 应用已被禁用。 <br> * 解决方案：在扣子编程中启用 OAuth 应用。 |
|  | invalid app type | * 原因：应用类型错误。 <br> * 解决方案：渠道应用暂不支持授权码模式。 |
|  | login session invalid | * 原因：登录态无效。 <br> * 解决方案：用户需要重新登录扣子编程。 |
| internal_error | Service internal error. | * 原因：服务内部错误。 <br> * 解决方案：建议稍后重试。 |


