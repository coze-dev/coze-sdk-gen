# 查看智能体版本列表
查看智能体版本列表。
## 接口描述
查询某个智能体的版本列表，支持查询已发布版本或未发布版本的版本号、版本创建者信息、创建时间等。

* 扣子个人版中，仅支持查询作为空间所有者的智能体。 
* 扣子团队版和企业版中，可以查看团队企业下的所有智能体。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/bots/:bot_id/versions <br> ``` <br>  |
| **权限** | `Bot.listVersion` <br> 确保调用该接口使用的访问令牌开通了 `Bot.listVersion` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查看智能体版本列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 必选 | 73428668***** | 待查询的智能体 ID。 <br> 进入智能体的开发页面，开发页面 URL 中 bot 参数后的数字就是智能体 ID。例如`https://www.coze.com/space/341****/bot/73428668*****`，bot ID 为`73428668*****`。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 `1`，即返回第一页数据。 |
| page_size | Integer | 可选 | 20 | 分页查询时每页返回的数据条数。默认值为 `10`，取值范围为 1~30。 |
| publish_status | String | 可选 | published_online | 智能体的发布状态，根据智能体的发布状态筛选对应版本。默认值为 `published_online`。枚举值： <br>  <br> * `published_online`：已发布的线上版本。 <br> * `unpublished_draft`：草稿版本。 |
| connector_id | String | 可选 | 1024 | 渠道 ID，仅在智能体发布状态为 `published_online`  时需要设置。用于筛选指定渠道下已发布的智能体版本。默认为空，表示获取所有发布渠道下的版本列表。 <br> 扣子编程的渠道 ID 包括： <br>  <br> * 1024：API 渠道。 <br> * 999：Chat SDK。 <br> * 10000122：扣子商店。 <br> * 10000113：微信客服。 <br> * 10000120：微信服务号。 <br> * 10000121：微信订阅号。 <br> * 10000126：抖音小程序。 <br> * 10000127：微信小程序。 <br> * 10000011：飞书。 <br> * 10000128：飞书多维表格。 <br> * 10000117：掘金。 <br> * 自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [ListBotVersionsData](#listbotversionsdata) | {"items":[{"creator":{"id":"24787743932***","name":"Susan"},"version":"171638852****","changelog":"更新了智能体的对话逻辑","created_at":"1718609571","publish_status":"published_online"}],"has_more":false} | 包含智能体版本列表数据的对象，用于返回智能体的所有版本信息及其分页状态。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### ListBotVersionsData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [OpenBotVersionInfo](#openbotversioninfo) | [{"creator":{"id":"24787743932***","name":"Susan"},"version":"171638852****","changelog":"更新了智能体的对话逻辑","created_at":"1718609571","publish_status":"published_online"}] | 智能体版本列表数据，包含该智能体的所有版本的详细信息。 |
| has_more | Boolean | false | 标识当前返回结果是否还有更多数据未加载。 <br>  <br> *  `true` ：当前返回的版本列表未包含所有符合条件的数据。 <br> *  `false`：表示已返回所有符合条件的数据。 |
### OpenBotVersionInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| creator | Object of [OpenCreatorInfo](#opencreatorinfo) | {"id":"24787743932***","name":"Susan"} | 版本的创建者信息，包含创建者的用户 ID 和用户名。 |
| version | String | 171638852**** | 智能体的版本号。 |
| changelog | String | 更新了智能体的对话逻辑 | 该版本发布时，用户输入的更新内容或修改说明。 |
| created_at | Long | 1718609571 | 该版本的创建时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| publish_status | String | published_online | 该版本的发布状态。枚举值： <br>  <br> * `published_online`：已发布的线上版本。 <br> * `unpublished_draft`：草稿版本。 |
### OpenCreatorInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 24787743932*** | 版本创建者的扣子用户 ID。 |
| name | String | Susan | 版本创建者的扣子用户名。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/bots/73428668*****/versions?page_num=1&page_size=20&publish_status=&connector_id=1024' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
  "data": {
    "items": [
      {
        "creator": {
          "id": "24787743932***",
          "name": "Susan"
        },
        "version": "171638852****",
        "changelog": "更新了智能体的对话逻辑",
        "created_at": 1718609571,
        "publish_status": "published_online"
      }
    ],
    "has_more": false
  },
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210152726467C48D89D6DB2****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。