# 查看应用列表
查看指定工作空间的应用列表。
## 接口描述
查看指定工作空间的应用列表，包含草稿状态的应用和已发布的应用。

* 扣子个人版中，你只能查询你作为工作空间所有者的应用。
* 扣子企业版中，你可以查看指定工作空间下的所有应用。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/apps <br> ``` <br>  |
| **权限** | `listApp` <br> 确保调用该接口使用的访问令牌开通了 `listApp` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查看指定工作空间的应用列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workspace_id | String | 必选 | 75123945629*** | 应用所在的工作空间的 Space ID。Space ID 是工作空间的唯一标识。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，工作空间 ID 为 `75814654762959***`。 |
| publish_status | String | 可选 | all | 应用的发布状态，用于筛选不同发布状态的应用。枚举值如下：  <br>  <br> * all ：所有状态。  <br> * published_online ：（默认值）已发布的正式版。  <br> * published_draft ：已发布但当前为草稿状态。  <br> * unpublished_draft ：从未发布过。 |
| connector_id | String | 可选 | 1024 | 渠道 ID，仅在 `publish_status` 为 `published_online` 或 `published_draft` 时需要设置。用于筛选该渠道下已发布的应用，包括已发布的线上正式版本和已发布的草稿版本。 <br> 默认值为 1024，即默认获取 API 发布渠道下的最新版本。 <br> 扣子编程的渠道 ID 包括： <br>  <br> * 1024：API 渠道。 <br> * 999：Chat SDK。 <br> * 10000122：扣子商店。 <br> * 10000113：微信客服。 <br> * 10000120：微信服务号。 <br> * 10000121：微信订阅号。 <br> * 10000126：抖音小程序。 <br> * 10000127：微信小程序。 <br> * 10000011：飞书。 <br> * 998：WebSDK。 <br> * 自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即返回第一页数据。 |
| page_size | Integer | 可选 | 20 | 分页大小。默认为 20，即每页返回 20 条数据。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenGetProjectData](#opengetprojectdata) | {"items":[{"id":"74876004423701****","name":"test","icon_url":"https://example.com/FileBizType.BIZ_BOT_APP/app1***.png","updated_at":1718289297,"description":"应用测试","is_published":true,"published_at":1718289297,"owner_user_id":"23423423****"}],"total":1} | 返回的应用列表数据，包含应用的基本信息和总数。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"1234567890abcdef"} | 包含请求的详细日志信息，用于问题排查和调试。 |
### OpenGetProjectData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [AppSimpleInfo](#appsimpleinfo) | [ { "id": "751614225987***", "is_published": false, "updated_at": 1749994622, "owner_user_id": "32902037***", "name": "***", "description": "***", "icon_url": "https://example.com/FileBizType.BIZ_BOT_ICON/***" } ] | 返回的应用列表。 |
| total | Long | 2 | 返回的应用的数量。 |
### AppSimpleInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 74876004423701**** | 应用的 ID。 |
| name | String | test | 应用的名称。 |
| icon_url | String | https://example.com/FileBizType.BIZ_BOT_APP/app1***.png | 应用图标的 URL。 |
| folder_id | String | 752316125533*** | 应用所属的文件夹 ID。 |
| updated_at | Long |  1718289297 | 应用的最近一次更新时间。以 Unix 时间戳格式表示。单位为秒。 |
| description | String | 应用测试 | 应用的描述。 |
| is_published | Boolean | true | 应用是否已发布。 <br>  <br> * `true` 表示已发布。 <br> * `false` 表示未发布。 |
| published_at | Long | 1718289297 | 应用的最近一次发布时间。以 Unix 时间戳格式表示。单位为秒。 <br> 仅当应用已发布时返回该值。如果应用是未发布过的草稿版本，则该参数的值为空。 |
| owner_user_id | String | 23423423**** | 应用创建者的扣子用户 ID。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
GET 'https://api.coze.cn/v1/apps?publish_status=all&workspace_id=751543725346***' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \ 
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
  "code": 0,
  "data": {
    "total": 1,
    "items": [
      {
        "id": "751614225987***",
        "is_published": true,
        "updated_at": 1749994622,
        "published_at": 1718289297,
        "owner_user_id": "32902037***",
        "name": "test",
        "folder_id": "752316125533542***",
        "description": "应用测试",
        "icon_url": "https://example.com/FileBizType.BIZ_BOT_ICON/***"
      }
    ]
  },
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。