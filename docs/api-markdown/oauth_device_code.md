# OAuth 设备授权
## 背景信息
设备授权模式（Device Authorization Grant）是 OAuth 2.0 的一个扩展，它允许那些没有浏览器或输入能力受限的设备获得访问令牌。这种模式特别适用于智能电视、游戏机、打印机等设备，这些设备可能没有方便的用户界面或输入方法来执行传统的用户代理（如浏览器）基础的授权流程。
在设备授权模式中，设备首先向授权服务器请求一个设备代码、用户代码和验证地址。设备代码用于设备随后向授权服务器申请令牌时的验证，而用户代码和验证 URI 则用于用户在另一台设备（如智能手机或电脑）上的浏览器中完成授权步骤。用户访问验证 URI 并输入用户代码，以批准设备的访问请求。一旦用户完成授权，设备就可以使用设备代码来获取访问令牌和刷新令牌。
这种模式的优势在于它不需要设备和用户代理之间有双向通信，使得那些无法运行完整 OAuth 流程的设备能够安全地获得访问令牌。RFC 8628 定义了设备授权模式的具体规范和流程。
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
   | 客户端类型 | 客户端类型，此处设置为**TV端/设备应用/类命令行程序**<span style="color: #060709">。</span> |
   | 应用名称 | 应用的名称，在扣子编程中全局唯一。  |
   | 描述  | 应用的基本描述信息。  |
5. 填写 App 的配置信息，单击**确定**。 
   | **配置** | **说明** |
   | --- | --- |
   | 权限 | 应用程序调用扣子 API 时需要的权限范围。 <br> 此处配置旨在于划定应用的权限范围，并未完成授权操作。授权操作可参考**授权流程**部分。 <br>  |
   | 客户端 ID | 客户端 ID，即 client id，是应用程序的公共标识符。 由扣子编程自动生成。 |

## 授权流程
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/8c0dad30104644dfbe11ff91502d6a4b~tplv-goo7wpa0wc-image.image)
设备码授权流程如下：

1. 设备客户端向授权服务器的设备授权端点发送请求（[Device Authorization](https://docs.coze.cn/api/open/docs/developer_guides/oauth_device_code#d54aa59f)），获取设备验证码（device_code）和用户验证码（user_code）。
   扣子编程支持多人协作场景下跨账号的 OAuth 授权，发起  API 请求时如果指定空间 ID，空间协作者也可以为应用程序授予团队空间中的资源权限。详细说明可参考[OAuth 授权（多人协作场景）](https://docs.coze.cn/api/open/docs/developer_guides/oauth_collaborate)。

2. 设备客户端将用户验证码和验证 URI 展示给用户，指导用户在另一台设备上输入这些信息以完成授权。
3. 用户在浏览器中输入验证 URI 和用户验证码，完成授权流程。
4. 设备客户端定期轮询授权服务器的令牌端点[Token](https://docs.coze.cn/api/open/docs/developer_guides/oauth_device_code#dbb2cf62)，使用设备代码来获取访问令牌。
5. 一旦用户授权成功，授权服务器响应设备客户端的令牌请求，并返回访问令牌和刷新令牌。

## API 参考
### Device Authorization
获取设备验证码。
#### 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | * 普通授权场景：https://api.coze.cn/api/permission/oauth2/device/code <br> * 多人协作场景：https://api.coze.cn/api/permission/oauth2/workspace_id/*${workspace_id}*/device/code |
仅扣子企业版支持多人协作场景，扣子个人版不支持多人协作场景。

#### Path
| **字段** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| workspace_id | String | 可选 | 空间 ID。多人协作场景下必选。 <br> 请求的 Path 中不指定空间 ID 时，授予 Access Token 当前登录账号拥有的所有空间权限；如果指定了空间 ID，表示授予 Access Token 指定空间的权限。资源范围为此空间下的所有资源，包括智能体、知识库、工作流等资源。详细说明可参考[OAuth 授权（多人协作场景）](https://docs.coze.cn/api/open/docs/developer_guides/oauth_collaborate)。 |
#### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Content-Type  | application/json | 请求正文的方式。 |
#### Body
| **字段** | **类型** | 是否必选 | **说明** |
| --- | --- | --- | --- |
| client_id | String | 必选 | 客户端 ID。创建 OAuth 应用时获取的客户端 ID。  |
#### Response
| **字段** | **类型** | **说明** |
| --- | --- | --- |
| device_code | String | 设备验证码 |
| user_code | Integer | 用户验证码 |
| verification_uri | String | 验证页面地址 |
| expires_in | Integer | device_code 和 user_code 可用时间，单位：秒，默认返回300秒； |
| interval <br>  | Integer | 客户端轮询Token请求的最短间隔时间，单位：秒， 默认为5秒。 |
#### 示例 
##### 请求示例 
```Shell
POST /device_authorization HTTP/1.1
Host: server.example.com
Content-Type: application/json

{
    "client_id": "1406020730"
}
```

##### 返回示例
```Shell
HTTP/1.1 200 OK
Content-Type: application/json

{
    "device_code": "GmRhmhcxhwAzkoEqiMEg_DnyEysNkuNhszIyS****",
    "user_code": "WDJB-MJHT",
    "verification_uri": "https://example.com/device",
    "expires_in": 1800,
    "interval": 5
}
```

### Token
获取 Token。
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
| grant_type | String | 必选 | 固定值，"urn:ietf:params:oauth:grant-type:device_code" |
| device_code <br>  | String | 必选 | 设备验证码，从 Device Authorization API response 中获取 |
#### 返回结果
| **字段** | **类型** | **说明** |
| --- | --- | --- |
| access_token | String | 访问令牌。 |
| expires_in | Integer | 访问令牌过期时间，秒级时间戳。 |
| refresh_token | String | 新的 refresh_token，用于重新获取 OAuth Access Token。 |
| error | String | 错误码： <br>  <br> * authorization_pending：用户还未完成授权，请稍后重试 <br> * slow_down：请求太频繁，请稍后重试 <br> * access_denied：用户已拒绝授权请求 <br> * expired_token：“device_code”已过期 |
| error_description | String | 错误详细描述 |
### 
