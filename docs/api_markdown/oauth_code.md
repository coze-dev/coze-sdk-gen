# OAuth 授权码授权
扣子编程支持应用程序通过 OAuth 授权码授权（Authorization Code Grant）的方式调用扣子编程 API。用户通过浏览器访问包含扣子编程功能的 Web 应用程序时，Web 应用程序重定向用户到扣子编程服务端以获取授权码（code），然后使用授权码向扣子编程服务端交换访问令牌。例如 Web 应用程序通过扣子编程 API 封装了扣子编程的查看 Bot 的详细信息和指定 Bot 对话等功能，用户使用这些功能之前，需要经由扣子编程服务端鉴权。
通过 OAuth 授权码方式授权时，应用程序需要拥有可通过 Web 访问的前端页面，否则无法实现重定向等一系列授权流程；同时应有稳定的后端架构，可处理前端请求、安全存储客户端密钥（Client Secret），与 OAuth 授权服务器和 OpenAPI 交互。

# 授权流程
Web 应用程序授权流程如下图所示。
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/4b881d91c69c43908995e2d92451b1d1~tplv-goo7wpa0wc-image.image)
具体流程说明如下：

1. 在扣子编程创建 OAuth 应用。
2. 应用程序调用 API [获取授权页面 URL](https://docs.coze.cn/api/open/docs/developer_guides/oauth_code#54010bd0)和[获取 OAuth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_code#b4f74244)，获取 OAuth 访问令牌。
3. 应用程序根据访问令牌调用扣子编程 API。

详细步骤如下：
## 1 创建 OAuth 应用
在扣子企业版（企业标准版、企业旗舰版）中，仅**超级管理员**和**管理员**有权限创建、编辑、删除 OAuth 应用，以及对应用进行授权操作。


1. 登录[扣子编程](https://code.coze.cn/home)。
2. 在左侧导航栏选择 **API & SDK。**
3. 在顶部单击**授权**  > **OAuth 应用**页签。
4. 在 OAuth 应用页面右上角单击**创建新应用**，填写应用的基本信息。 
   | **配置** | **说明** |
   | --- | --- |
   | 应用类型  | 应用的类型，此处设置为**普通**。 |
   | 客户端类型 | 客户端类型，此处设置为**Web 后端应用**。 |
   | 应用名称  | 应用的名称，在扣子编程中全局唯一。  |
   | 描述  | 应用的基本描述信息。  |
5. 填写 App 的配置信息。 
   | **配置** | **说明** |
   | --- | --- |
   | 权限  | 应用程序调用扣子 API 时需要的权限范围。 <br> 此处配置旨在于划定应用的权限范围，并未完成授权操作。创建 Oauth 应用后还需要参考后续操作完成授权。 <br>  |
   | 重定向 URL  | 重定向的 URL 地址。用户完成授权后，扣子编程的授权服务器将通过重定向 URL 返回授权相关的凭据。最多可添加 3 个不同的 URL 地址。 <br>  <br> * 重定向 URL 仅支持 HTTP 和 HTTPs。为了保证数据传输安全，请勿在生产环境使用HTTP 协议地址。 <br> * 对于测试场景，您可以指定引用本地机器的 URL，例如 `http://localhost:8080`。 |
   | 客户端 ID 和客户端密钥  <br>  | 客户端 ID 和客户端密钥均由扣子编程自动生成并配置。单击生成客户端密钥，并复制系统自动生成的客户端密钥。  <br>  <br> * 客户端 ID：即 client id，是应用程序的公共标识符。  <br> * 客户端密钥：即 client secret，仅应用程序和授权服务器有权访问的密钥。用于获取已登录用户的访问令牌。  <br>  <br> * 此客户端密钥仅显示一次。请将其保存到安全且易于访问的地方。不要与他人分享，或在其他环境中暴露。  <br> * 支持生成多个客户端密钥，当您发现密钥泄露时，可以通过此方法进行密钥轮转。 <br> * 扣子编程在任何时候都不会存储您的客户端密钥，扣子服务端仅存储符合行业规范的密钥摘要值。 <br>  |
6. 单击**确定**，完成配置。 

## 2 获取访问令牌

1. 终端用户在 Web 应用程序中触发授权操作，例如点击和 Bot 对话的按钮。该动作对应扣子发起对话 API，应用程序需要获得扣子账号的授权。
2. Web 应用程序重定向用户到授权服务器以获取 code。
   应用程序通过302重定向方式发起 [获取授权页面 URL](https://docs.coze.cn/api/open/docs/developer_guides/oauth_code#54010bd0) API 请求。
   请求中携带 OAuth 应用的客户端 ID 和客户端密钥 、重定向 URL 等信息。请求示例如下：
   ```Shell
   curl --location --request GET 'https://www.coze.cn/api/permission/oauth2/authorize?response_type=code&client_id=8173420653665306615182245269****.app.coze&redirect_uri=https://www.coze.cn/open/oauth/apps&state=1294848'
   ```

   Response Header 中的 location 字段中为跳转链接。例如`https://www.coze.cn/oauth/consent?authorize_key=JacVeqTW93ps5m5N9n349bEBgIsWrnNp`。浏览器跳转到此 URL，引导用户完成扣子账号授权。授权页面示例如下：
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/98212aa5137243c1ab5a6e76f5e6ae93~tplv-goo7wpa0wc-image.image)
   * 如果你是超级管理员和管理员，单击**授权**后，可以主动安装该应用。
   * 如果你是成员，单击**授权**后，如果系统提示授权失败，说明超级管理员或管理员未安装该应用，请单击**发起申请**。超级管理员或管理员收到安装请求后，在对话框中单击**安装**或在**应用安装管理**页面的**操作**列中单击**安装**。
   * 扣子支持多人协作场景下跨账号的 OAuth 授权，发起 [获取授权页面 URL](https://docs.coze.cn/api/open/docs/developer_guides/oauth_code#54010bd0) API 请求时如果指定空间 ID，空间协作者也可以为应用程序授予团队空间中的资源权限。详细说明可参考[OAuth 授权（多人协作场景）](https://docs.coze.cn/api/open/docs/developer_guides/oauth_collaborate)。

3. 扣子服务端会在 API 的响应中返回 code。
   从重定向的 URL 地址中获取 code，例如本示例中 code 为 `code_WZmPRDcjJhfwHD****`。
   ```Shell
   https://www.coze.cn/open/oauth/apps?code=code_WZmPRDcjJhfwHD****&state=1294848
   ```

4. 使用授权码交换访问令牌。
   应用程序向扣子服务端发起 [获取 OAuth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_code#b4f74244) 请求，请求中携带 code，扣子服务端会在 API 的响应中返回 access_token 和 refresh_token。其中：
   * access_token 即访问令牌，用于发起扣子 API 请求时鉴权，有效期为 15 分钟。
   * refresh_token 用于刷新 access_token，有效期为 30 天。refresh_token 到期前可以多次调用 [刷新 OAuth Access Token ](https://docs.coze.cn/api/open/docs/developer_guides/oauth_code#9c63ccf1)接口获取新的 refresh_token 和 access_token。
   接口示例如下：
   
   <div type="doc-tabs">
   <div type="tab-item" title="请求示例" key="brSyOePtEr">
   
   ```Shell
   curl --location --request POST 'https://api.coze.cn/api/permission/oauth2/token' \
   --header 'Authorization: Bearer hDPU8gexPcwChkhMvmjvR14yQ1HWoaB42tCd0rjrc55G****' \
   --header 'Content-Type: application/json' \
   --data '{
       "grant_type": "authorization_code",
       "client_id": "8173420653665306615182245269****.app.coze",
       "redirect_uri": "https://www.coze.cn/open/oauth/apps",
       "code": "bfiqrhedxsdvnuivher****"
   }'
   ```
   
   
   </div>
   <div type="tab-item" title="响应示例" key="tf7wHRc0oO">
   
   ```JSON
   {
       "access_token": "czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****",
       "expires_in": 1720098388,
       "refresh_token": "LBEP9iWU7rn60PWa58GER5rr6vygb5WSACu2vASlCQu7kpFkavCrNa9BBDpHLUlGd46a****"
   }
   ```
   
   
   </div>
   </div>

## 3 发起扣子 API 请求
在 API 请求头中通过 `Authorization=Bearer `*`$Access_Token`* 指定访问令牌，发起扣子 API 请求。每个接口对应的权限点不同。
以[获取已发布智能体的配置（即将下线）](https://docs.coze.cn/api/open/docs/developer_guides/get_metadata) API 为例，完整的 API 请求如下：
```Shell
curl --location --request GET 'https://api.coze.cn/v1/bot/get_online_info?bot_id=73428668*****' \ 
--header 'Authorization: Bearer czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****' \
```

# 获取授权页面 URL
Web 应用程序可调用此 API 获取 OAuth 授权页面 URL 地址。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | * 普通授权场景：https://www.coze.cn/api/permission/oauth2/authorize <br> * 多人协作场景：https://www.coze.cn/api/permission/oauth2/workspace_id/*${workspace_id}*/authorize |
仅扣子企业版支持多人协作场景，扣子个人版不支持多人协作场景。

## Path 参数
| **字段** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| workspace_id | String | 可选 | 空间 ID。多人协作场景下必选。 <br> 请求的 Path 中不指定空间 ID 时，授予 Access Token 当前登录账号拥有的所有空间权限；如果指定了空间 ID，表示授予 Access Token 指定空间的权限。资源范围为此空间下的所有资源，包括 Bot、知识库、工作流等资源。详细说明可参考[OAuth 授权（多人协作场景）](https://docs.coze.cn/api/open/docs/developer_guides/oauth_collaborate)。 |
## Query
| **字段** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| response_type | String | 必选 | 固定为 code。 |
| client_id | String | 必选 | 客户端 ID。创建 OAuth 应用时获取的客户端 ID。  |
| redirect_uri | String | 必选 | 重定向 URL。创建 OAuth 应用时指定的重定向 URL。  |
| state | String | 必选 | 通常情况下，state 是一串随机字符串或哈希值。客户端在发起授权请求时将其附加到请求 URL 中，终端用户完成授权时，客户端会验证返回的 `state` 值是否与原始值匹配。如果不匹配或 `state` 参数丢失，客户端应拒绝这次授权请求。 |
## 返回结果 
返回结果的 Http code 为 302，且 Response Header 中的 location 字段中为跳转链接。例如`https://www.coze.cn/oauth/consent?authorize_key=JacVeqTW93ps5m5N9n349bEBgIsWrn****`。浏览器跳转到此 URL，引导用户完成 Coze 账号授权。
## 示例 
### 请求示例 
```Shell
curl --location --request GET 'https://www.coze.cn/api/permission/oauth2/authorize?response_type=code&client_id=8173420653665306615182245269****.app.coze&redirect_uri=https://www.coze.cn/open/oauth/apps&state=1294848'
```

### 返回示例
```Shell
https://www.coze.cn/oauth/consent?authorize_key=JacVeqTW93ps5m5N9n349bEBgIsW****
```

# 获取 OAuth Access Token
通过授权码（OAuth code）获取 OAuth Access Token。
## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/api/permission/oauth2/token |
## Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Content-Type  | application/json | 请求正文的方式。 |
| Authorization | Bearer <span style="color: #D83931"><strong><em>$Client_Secret</em></strong></span> | 创建 OAuth 应用时获取的客户端密钥。  |
## Body
| **字段** | **类型** | 是否必选 | **说明** |
| --- | --- | --- | --- |
| grant_type | String | 必选 | 固定为 authorization_code。 |
| code | String | 必选 | 授权码，即终端用户授权后，页面重定向的 URL 中 code 参数的值。 <br> 例如页面重定向的 URL 为 `https://www.coze.cn/open/oauth/apps?code=code_IzLl3dnNgvSTq50ijYHZZi6Xqun67ocrxrzeKxDEUo32****&state=1294848`，code 为 `code_IzLl3dnNgvSTq50ijYHZZi6Xqun67ocrxrzeKxDEUo32****` 。 |
| client_id | String | 必选 | 创建 OAuth 应用时获取的客户端 ID。  |
| redirect_uri | String | 必选 | 创建 OAuth 应用时指定的重定向 URL。  |
## 返回结果
| **字段** | **类型** | **说明** |
| --- | --- | --- |
| access_token | String | 访问令牌。 |
| expires_in | Integer | 访问令牌过期时间，格式为 Unixtime 时间戳，精度为秒，有效期为 15 分钟。 |
| refresh_token | String | 刷新令牌，用于重新获取访问令牌。refresh_token 有效期为 30 天。有效期内可以凭 refresh_token 调用 API [刷新 OAuth Access Token ](https://docs.coze.cn/api/open/docs/developer_guides/oauth_code#9c63ccf1)获取新的 OAuth Access Token。 |
## 示例 
### 请求示例 
```Shell
curl --location 'https://api.coze.cn/api/permission/oauth2/token' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer czvSJMcRob40yQ04HmSyCbEw6h22r0LJwHcKyu13H2ic****' \
--data '{
    "grant_type": "authorization_code",
    "code": "code_hS0m5XlDLSAwWjaadftNqiOGAmOW02wSDiGunZETWQuR****",
    "redirect_uri": "https://coze.com/",
    "client_id": "9767336922475223578182683125****.app.coze"
}'
```

### 返回示例
```JSON
{
    "access_token": "czu_UEE2mJn66h0fMHxLCVv9uQ7HAoNNS8DmF6N6grjWmkHX2jPm8SR0tJcKop8v****",
    "expires_in": 1720098388,
    "refresh_token": "LBEP9iWU7rn60PWa58GER5rr6vygb5WSACu2vASlCQu7kpFkavCrNa9BBDpHLUlGd4****"
}
```

# 刷新 OAuth Access Token 
根据 refresh_token 获取新的 OAuth Access Token。
调用[获取 OAuth Access Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_code#b4f74244) API 获取的 OAuth Access Token 有效期为 15 分钟，refresh_token 有效期为 30 天。refresh_token 有效期内可以调用此 API 获取新的 OAuth Access Token。
接口调用成功后，入参中指定的 refresh_token 失效。
## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/api/permission/oauth2/token |
## Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Content-Type  | application/json | 请求正文的方式。 |
| Authorization | Bearer ***$Client_Secret*** | 创建 OAuth 应用时获取的客户端密钥。  |
## Body
| **字段** | **类型** | 是否必选 | **说明** |
| --- | --- | --- | --- |
| client_id | String | 必选 | 创建 OAuth 应用时获取的客户端 ID。  |
| grant_type | String | 必选 | 固定为 refresh_token。 |
| refresh_token | String | 必选 | 调用**获取 OAuth Access Token** API 获取的 refresh_token，且 refresh_token 应在有效期内。 |
## 返回结果
| **字段** | **类型** | **说明** |
| --- | --- | --- |
| access_token | String | 访问令牌。 |
| expires_in | Integer | 访问令牌过期时间，秒级时间戳。 |
| refresh_token | String | 新的 refresh_token，用于重新获取 OAuth Access Token。 |
## 示例 
### 请求示例 
```Shell
curl --location 'https://api.coze.cn/api/permission/oauth2/token' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer czvSJMcRob40yQ04HmSyCbEw6h22r0LJwHcKyu13H2ic****' \
--data '{
    "grant_type": "refresh_token",
    "refresh_token": "1gCCMSkz6F4u0br1zJwoNhK041SToNentkilrWFaKzUVK9c3dds3e5DVn23f1ROkflYf****",
    "client_id": "9767336922475223578182683125****.app.coze"
}'
```

### 返回示例
```JSON
{
    "access_token": "czu_UVbDyB94NwXsMTJEvIwPp15oJycgl1y4rFoy774L9eFA0eKxmWGyUKbUt2uB****",
    "expires_in": 1720853011,
    "refresh_token": "NYbHWyusGWA03Ar1dpUMr6iuImN9bydfwRWWKpcKqcpTN9sDwmCxopt7Jg7HW6DGJnXL****"
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
### 如何处理 OAuth 授权码授权失败，提示“OAuth 错误/请求参数错误”？
请检查回调地址中的特殊字符是否已正确转义。例如，回调地址中的 `#` 需要转义为 `%23`。确保所有特殊字符都已按照URL 编码规则进行转义，然后重新尝试授权。