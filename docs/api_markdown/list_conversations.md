# 查看会话列表
查看指定智能体的会话列表。
* 调用此 API 之前，应确认当前使用的访问密钥拥有智能体所在工作空间的权限。
* 仅支持通过此 API 查看智能体在 API 或 Chat SDK 渠道产生的会话。
* 仅支持查询本人创建的会话。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/conversations <br> ``` <br>  |
| **权限** | `listConversation` <br> 确保调用该接口使用的访问令牌开通了 `listConversation` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 获取指定智能体的会话列表。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 必选 | 73428668***** | 智能体 ID，获取方法如下： <br> 进入智能体的 开发页面，开发页面 URL 中 `bot` 参数后的数字就是智能体 ID。例如`https://www.coze.cn/space/341****/bot/73428668*****`，智能体 ID 为`73428668*****`。 |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即返回第一页数据。 |
| page_size | Integer | 可选 | 40 | 每一页的数据条目个数，默认为 50，最大为 50。 |
| sort_order | String | 可选 | ASC | 会话列表的排序方式： <br>  <br> * **ASC**：按创建时间升序排序，最早创建的会话排序最前。 <br> * **DESC**：（默认）按创建时间降序排序，最近创建的会话排序最前。 |
| connector_id | String | 可选 | 999 | 发布渠道 ID，用于筛选指定渠道的会话。仅支持查看以下渠道的会话： <br>  <br> * （默认）API 渠道，渠道 ID 为 1024。 <br> * Chat SDK 渠道，渠道 ID 为 999。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [ListConversationData](#listconversationdata) | [{"id":"737999610479815****","name":"推荐杭州美食","meta_data":{"uuid":"newid1234"},"creator_id":"247877439325****","created_at":1718289297,"updated_at":1718289297,"last_section_id":"7495664347616952360","connector_id":"1024"}] | 会话列表的详细。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "202412345678901234567890" } | 响应的详细信息。 |
### ListConversationData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| has_more | Boolean | false | 是否还有更多会话未在本次请求中返回。 <br>  <br> * true：还有更多未返回的会话。 <br> * false：已返回符合筛选条件的全部会话。 |
| conversations | Array of [ConversationData](#conversationdata) | {"id":"737999610479815****","name":"推荐杭州美食","meta_data":{"uuid":"newid1234"},"creator_id":"247877439325****","created_at":1718289297,"updated_at":1718289297,"last_section_id":"7495664347616952360","connector_id":"1024"} | 会话的详细信息。 |
### ConversationData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 737999610479815**** | Conversation ID，即会话的唯一标识。 |
| name | String | 推荐杭州美食 | 会话的名称，用于标识和区分不同的会话。 |
| meta_data | JSON Map | { "uuid": "newid1234" } | 附加信息，通常用于封装一些业务相关的字段。[查看会话信息](https://docs.coze.cn/developer_guides/retrieve_conversation)时，扣子编程会透传此附加信息，[查看消息列表](https://docs.coze.cn/developer_guides/list_message)时不会返回该附加信息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
| created_at | Long | 1718289297 | 会话创建的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| updated_at | Long | 1718289297 | 会话的最后更新时间，格式为 10 位的 Unix 时间戳，单位为秒。 |
| last_section_id | String | 749566434761695*** | 会话中最新的一个上下文片段 ID。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/conversations?bot_id=73428668*****&page_num=1&page_size=40&sort_order=ASC&connector_id=999' \
--header 'Authorization: Bearer pat_xitq9LWlowpX3qGCih1lwpAdzvXNqgmpfhpV28HLWFN****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "has_more": false,
    "conversations": [
      {
        "id": "737999610479815****",
        "name": "推荐杭州美食",
        "meta_data": {
          "uuid": "newid1234"
        },
        "creator_id": "247877439325****",
        "connector_id": "1024",
        "created_at": 1718289297,
        "updated_at": 1718289297,
        "last_section_id": "7495664347616952360"
      }
    ]
  },
  "detail": {
    "logid": "20241210152726467C48D89D6DB2****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。