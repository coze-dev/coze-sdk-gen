# OAuth PKCE
对于移动设备、本地应用、单页应用等公开客户端，授权码授权模式存在一定安全风险，例如授权码拦截攻击等。对于此类应用，推荐使用 PKCE 扩展协议增强授权码流程的安全性。
## 背景信息
在移动设备、本地应用、单页应用等公开客户端授权场景下，使用授权码模式授权时，Client ID 和 Client Secret 可能意外泄露或被恶意窃取，造成访问密钥泄露，影响数据安全。对于公开客户端，推荐使用 PKCE 扩展协议增强授权码流程的安全性。PKCE（Proof Key for Code Exchange）是一个 OAuth 2.0 的扩展协议，PKCE 授权流程中引入了客户端生成的随机字符串（code_verifier）作为临时密钥，防止授权码被拦截和滥用，提高了授权流程的安全性。
PKCE 授权流程如下：

1. 应用程序引导用户发起授权请求，携带临时密钥。
2. 用户同意授权后，扣子 OAuth Server 返回授权码。
3. 应用程序通过临时密钥和授权码获取 OAuth 访问令牌。
4. 扣子 OAuth Server 验证临时密钥和授权码，返回 Access Token。

## 准备工作
在应用程序启动授权流程之前，开发者应在扣子编程中创建 Oauth 应用，获取客户端 ID。
在扣子企业版（企业标准版、企业旗舰版）中，仅**超级管理员**和**管理员**有权限创建、编辑、删除 OAuth 应用，以及对应用进行授权操作。


1. 登录[扣子编程](https://code.coze.cn/home)。
2. 在左侧导航栏选择 **API & SDK。**
3. 在顶部单击**授权**  > **OAuth 应用**页签。
4. 在 OAuth 应用页面右上角单击**创建新应用**，填写应用的基本信息，并单击**创建并继续**。
   | **配置** | **说明** |
   | --- | --- |
   | 应用类型  | OAuth 应用的类型，此处设置为**普通**。 |
   | 客户端类型 | 客户端类型，此处设置为**移动端/PC 桌面端/单页面应用**。 |
   | 应用名称  | 应用的名称，在扣子编程中全局唯一。  |
   | 描述  | 应用的基本描述信息。  |
5. 填写 App 的配置信息，并单击**确定**。 
   | **配置** | **说明** |
   | --- | --- |
   | 权限  | 应用程序调用扣子 API 时需要的权限范围。 <br> 此处配置旨在于划定应用的权限范围，并未完成授权操作。授权操作可参考**授权流程**部分。 <br>  |
   | 重定向 URL  | 重定向的 URL 地址。用户完成授权后，扣子编程的授权服务器将通过重定向 URL 返回授权相关的凭据。最多可添加 3 个不同的 URL 地址。 <br>  <br> * 重定向 URL 仅支持 HTTP 和 HTTPs。为了保证数据传输安全，请勿在生产环境使用HTTP 协议地址。 <br> * 对于测试场景，您可以指定引用本地机器的 URL，例如 `http://localhost:8080`。 |
   | 客户端 ID | 客户端 ID，即 client id，是应用程序的公共标识符。 由扣子编程自动生成。 |

## 授权流程
PKCE 协议的授权流程如下：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/551be271cc5b4b67afe0985764a2e7e5~tplv-goo7wpa0wc-image.image)
### 步骤一：发起授权请求

1. 终端用户在 Web 应用程序中触发授权操作。
   例如点击和 Bot 对话的按钮。该动作对应扣子**发起对话** API，应用程序需要获得扣子账号的授权。
2. 应用程序生成一个临时密钥。
   客户端生成一个随机值 code_verifier，并根据指定算法将其转换为 code_challenge。其中转换算法为 code_challenge_method，转换通常使用 SHA-256 算法，并进行 Base64URL 编码，即 `code_challenge = BASE64URL-ENCODE(SHA256(ASCII(code_verifier)))`。
3. 应用程序重定向用户到授权服务器。
   应用程序通过 302 重定向方式发起 [获取授权页面 URL](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#d54aa59f) API 请求。请求中携带 OAuth 应用的客户端 ID、重定向 URL、临时密钥 code_challenge、转换算法 code_challenge_method 等信息。请求示例如下：
   ```Shell
   curl --location --request GET 'https://www.coze.cn/api/permission/oauth2/authorize?response_type=code&client_id=8173420653665306615182245269****.app.coze&redirect_uri=https://www.coze.cn/open/oauth/apps&state=1294848&code_challenge=*****&code_challenge_method=S256'
   ```

   扣子编程支持多人协作场景下跨账号的 OAuth 授权，发起 [获取授权页面 URL](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#d54aa59f) API 请求时如果指定空间 ID，空间协作者也可以为应用程序授予团队空间中的资源权限。详细说明可参考[OAuth 授权（多人协作场景）](https://docs.coze.cn/api/open/docs/developer_guides/oauth_collaborate)。


### 步骤二：获取授权码

1. 终端用户在授权页面单击**授权**。
   Response Header 中的 location 字段中为跳转链接。例如`https://www.coze.cn/oauth/consent?authorize_key=JacVeqTW93ps5m5N9n34***rnNp`。浏览器跳转到此 URL，引导用户完成扣子账号授权。授权页面示例如下：
   若用户未登录，前端页面应引导其先完成登录。在授权页点击授权前，终端用户需已成功登录扣子，才能获取扣子个人付费版、企业版账号权限。

   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/98212aa5137243c1ab5a6e76f5e6ae93~tplv-goo7wpa0wc-image.image)
2. 扣子 OAuth Server 在 API 的响应中返回授权码 code。
   从重定向的 URL 地址中获取 code，例如本示例中 code 为 `code_WZmPRDcjJhfwHD****`。
   ```Shell
   https://www.coze.cn/open/oauth/apps?code=code_WZmPRDcjJhfwHD****&state=1294848
   ```


### 步骤三：请求令牌
应用程序向扣子服务端发起 [获取 OAuth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#089e1996) 请求，请求中携带授权码 code 和临时密钥转换前的原始随机值 code_verifier。
[获取 OAuth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#089e1996) 的请求示例如下：
```Shell
curl --location 'https://api.coze.cn/api/permission/oauth2/token' \
--header 'Content-Type: application/json' \
--data '{
    "grant_type": "authorization_code",
    "code": "code_hS0m5XlDLSAwWjaadftNqiOGAmOW02wSDiGunZETWQuR****",
    "redirect_uri": "https://coze.com/",
    "client_id": "9767336922475223578182683125****.app.coze",
    "code_verifier": "9767****"
}'
```

### 步骤四：获取令牌

1. 收到请求后，扣子 OAuth Server 将对以下三个参数进行对比验证。
   * 原始的随机值 code_verifier
   * 转换算法 code_challenge_method
   * 临时密钥 code_challenge
2. 验证通过后，扣子 OAuth Server 将通过 [获取 OAuth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#089e1996) API 响应的方式颁发令牌。
   其中：
   * access_token 即访问令牌，用于发起扣子 API 请求时鉴权，有效期为 15 分钟。
   * refresh_token 用于刷新 access_token，有效期为 30 天。refresh_token 到期前可以多次调用 [刷新 OAuth Access Token ](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#cc0e4e16)接口获取新的 refresh_token 和 access_token。

 [获取 OAuth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#089e1996) 的响应示例如下：
```JSON
{
    "access_token": "czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****",
    "expires_in": 1720098388,
    "refresh_token": "LBEP9iWU7rn60PWa58GER5rr6vygb5WSACu2vASlCQu7kpFkavCrNa9BBDpHLUlGd46a****"
}
```

## 后续操作
应用程序可以在 API 请求头中通过 `Authorization=Bearer `*`$Access_Token`* 指定访问令牌，发起扣子 API 请求。每个接口对应的权限点不同。
以[获取已发布智能体的配置（即将下线）](https://docs.coze.cn/api/open/docs/developer_guides/get_metadata) API 为例，完整的 API 请求如下：
```Shell
curl --location --request GET 'https://api.coze.cn/v1/bot/get_online_info?bot_id=73428668*****' \ 
--header 'Authorization: Bearer czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****' \
```

## API 参考
### 获取授权页面 URL
Web 应用程序可调用此 API 获取 OAuth 授权页面 URL 地址。
#### 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | * 普通授权场景：https://www.coze.cn/api/permission/oauth2/authorize <br> * 多人协作场景：https://www.coze.cn/api/permission/oauth2/workspace_id/*${workspace_id}*/authorize |
仅扣子企业版支持多人协作场景，扣子个人版不支持多人协作场景。

#### Path 参数
| **字段** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| workspace_id | String | 可选 | 空间 ID。多人协作场景下必选。 <br> 请求的 Path 中不指定空间 ID 时，授予 Access Token 当前登录账号拥有的所有空间权限；如果指定了空间 ID，表示授予 Access Token 指定空间的权限。资源范围为此空间下的所有资源，包括 Bot、知识库、工作流等资源。详细说明可参考[OAuth 授权（多人协作场景）](https://docs.coze.cn/api/open/docs/developer_guides/oauth_collaborate)。 |
#### Query
| **字段** | **类型** | 是否必选 | **说明** |
| --- | --- | --- | --- |
| client_id | String | 必选 | 客户端 ID。创建 OAuth 应用时获取的客户端 ID。  |
| response_type | String | 必选 | 固定为 code。 |
| redirect_uri | String | 必选 | 重定向 URL。创建 OAuth 应用时指定的重定向 URL。  |
| state | String | 必选 | 通常情况下，state 是一串随机字符串或哈希值。客户端在发起授权请求时将其附加到请求 URL 中，终端用户完成授权时，客户端会验证返回的 `state` 值是否与原始值匹配。如果不匹配或 `state` 参数丢失，客户端应拒绝这次授权请求。 |
| code_challenge <br>  | String | 必选 | PKCE 协议的临时密钥。应用程序的服务端将一个随机数根据  <br> code_challenge_method 指定的算法进行计算和转换，生成 code_challenge。 <br> 仅 PKCE 授权场景下使用，且该场景下此参数为必选。 |
| code_challenge_method <br>  | String | 必选 | PKCE 协议的临时密钥算法。支持设置为： <br>  <br> * plain：（默认）原始值，表示未计算，此时`code_challenge = code_verifier`。 <br> * S256：SHA256HASH。转换方式为 <br>  <br> `code_challenge = BASE64URL-ENCODE(SHA256(ASCII(code_verifier)))`。 <br> 仅 PKCE 授权场景下使用，且该场景下此参数为必选。 |
#### 返回结果 
返回结果的 Http code 为 302，其中包含授权页地址 https://www.coze.cn/oauth2/authorize。

* 如果当前用户未合法登录态，则 302 重定向至授权页 `https://www.coze.cn/oauth2/authorize?authorize_key=1234gdaskljgflan`。
* 如果当前用户未登录，则 302 重定向至登录页，登录页重定向至授权页 `https://www.coze.cn/sign?redirect=https://www.coze.cn/oauth2/authorize?authorize_key=1234gdaskljgflan`。

#### 示例 
##### 请求示例 
```Shell
curl --location --request GET 'https://www.coze.cn/api/permission/oauth2/authorize?response_type=code&client_id=8173420653665306615182245269****.app.coze&redirect_uri=https://www.coze.cn/open/oauth/apps&state=1294848&code_challenge=*****&code_challenge_method=S256'
```

##### 返回示例
```Shell
https://www.coze.cn/oauth2/authorize?authorize_key=1234gdaskljgflan
```

### 获取 OAuth Access Token
通过授权码（OAuth code）获取 OAuth Access Token。
#### 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/api/permission/oauth2/token |
#### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Content-Type  | application/json | 请求正文的方式。 |
#### Body
| **字段** | **类型** | 是否必选 | **说明** |
| --- | --- | --- | --- |
| grant_type | String | 必选 | 固定为 authorization_code。 |
| code | String | 必选 | 授权码，即终端用户授权后，页面重定向的 URL 中 code 参数的值。 <br> 例如页面重定向的 URL 为 `https://www.coze.cn/open/oauth/apps?code=code_IzLl3dnNgvSTq50ijYHZZi6Xqun67ocrxrzeKxDEUo32****&state=1294848`，code 为 `code_IzLl3dnNgvSTq50ijYHZZi6Xqun67ocrxrzeKxDEUo32****` 。 |
| client_id | String | 必选 | 创建 OAuth 应用时获取的客户端 ID。  |
| redirect_uri | String | 必选 | 创建 OAuth 应用时指定的重定向 URL。  |
| code_verifier <br>  | String | 可选 | PKCE 协议授权流程中，应用程序的服务端生成的原始随机数。 <br> 仅 PKCE 授权场景下使用。 |
#### 返回结果
| **字段** | **类型** | **说明** |
| --- | --- | --- |
| access_token | String | 访问令牌。 |
| expires_in | Integer | 访问令牌过期时间，格式为 Unixtime 时间戳，精度为秒，有效期为 15 分钟。 |
| refresh_token | String | 刷新令牌，用于重新获取访问令牌。refresh_token 有效期为 30 天。有效期内可以凭 refresh_token 调用 API  [刷新 OAuth Access Token ](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#cc0e4e16)获取新的 OAuth Access Token。 |
#### 示例 
##### 请求示例 
```Shell
curl --location 'https://api.coze.cn/api/permission/oauth2/token' \
--header 'Content-Type: application/json' \
--data '{
    "grant_type": "authorization_code",
    "code": "code_hS0m5XlDLSAwWjaadftNqiOGAmOW02wSDiGunZETWQuR****",
    "redirect_uri": "https://coze.com/",
    "client_id": "9767336922475223578182683125****.app.coze",
    "code_verifier": "9767****"
}'
```

##### 返回示例
```JSON
{
    "access_token": "czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****",
    "expires_in": 1720098388,
    "refresh_token": "LBEP9iWU7rn60PWa58GER5rr6vygb5WSACu2vASlCQu7kpFkavCrNa9BBDpHLUlGd4****"
}
```

### 刷新 OAuth Access Token 
根据 refresh_token 获取新的 OAuth Access Token。
调用[获取 OAuth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_pkce#089e1996) API 获取的 OAuth Access Token 有效期为 15 分钟，refresh_token 有效期为 30 天。refresh_token 有效期内可以调用此 API 获取新的 OAuth Access Token。
接口调用成功后，入参中指定的 refresh_token 失效。
#### 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/api/permission/oauth2/token |
#### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Content-Type  | application/json | 请求正文的方式。 |
#### Body
| **字段** | **类型** | 是否必选 | **说明** |
| --- | --- | --- | --- |
| client_id | String | 必选 | 创建 OAuth 应用时获取的客户端 ID。  |
| grant_type | String | 必选 | 固定为 refresh_token。 |
| refresh_token | String | 必选 | 调用**获取 OAuth Access Token** API 获取的 refresh_token，且 refresh_token 应在有效期内。 |
#### 返回结果
| **字段** | **类型** | **说明** |
| --- | --- | --- |
| access_token | String | 访问令牌。 |
| expires_in | Integer | 访问令牌过期时间，秒级时间戳。 |
| refresh_token | String | 新的 refresh_token，用于重新获取 OAuth Access Token。 |
#### 示例 
##### 请求示例 
```Shell
curl --location 'https://api.coze.cn/api/permission/oauth2/token' \
--header 'Content-Type: application/json' \
--data '{
    "grant_type": "refresh_token",
    "refresh_token": "1gCCMSkz6F4u0br1zJwoNhK041SToNentkilrWFaKzUVK9c3dds3e5DVn23f1ROkflYf****",
    "client_id": "9767336922475223578182683125****.app.coze"
}'
```

##### 返回示例
```JSON
{
    "access_token": "czu_UVbDyB94NwXsMTJEvIwPp15oJycgl1y4rFoy774L9eFA0eKxmWGyUKbUt2uB****",
    "expires_in": 1720853011,
    "refresh_token": "NYbHWyusGWA03Ar1dpUMr6iuImN9bydfwRWWKpcKqcpTN9sDwmCxopt7Jg7HW6DGJnXL****"
}
```

## 
