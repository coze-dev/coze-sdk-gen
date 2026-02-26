# 查询付费插件列表
查询插件商店中的付费插件列表。
## 接口描述
查询插件商店中的付费插件列表，包括扣子编程官方插件和三方插件，或通过关键词模糊搜索相关插件。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/stores/plugins <br> ``` <br>  |
| **权限** | 无 |
| **接口说明** | 查询插件商店中的付费插件列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| keyword | String | 可选 | 语音识别 | 插件搜索的关键词，支持模糊匹配。 |
| is_official | Boolean | 可选 | true | 是否为扣子编程官方插件。默认返回官方插件和三方插件。 <br>  <br> *  `true` ：官方插件。 <br> *  `false`：三方插件。 |
| category_ids | Array of String | 可选 | 747593003997698***,73271372757147*** | 插件分类 ID 列表，用于筛选指定多个分类下的插件。默认为空，即返回所有分类下的插件。可以通过[查询插件分类](https://docs.coze.cn/developer_guides/list_plugin_category)API 获取对应的插件分类 ID。 |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即返回第一页数据。 |
| page_size | Integer | 可选 | 20 | 每页返回的数据条数，用于分页查询。默认值为 20，最大值为 100。 |
| sort_type | String | 可选 | heat | 排序类型，用于指定返回插件的排序方式。支持的排序方式如下所示： <br>  <br> * `heat`：最受欢迎。 <br> * `newest`：最近发布。 <br> * `relative`：相关性，仅用于搜索场景。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [ListPluginData](#listplugindata) | \ | 返回的插件列表及其相关信息的对象。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### ListPluginData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [ProductPluginInfo](#productplugininfo) | \ | 插件列表，包含多个插件项的数组。每个插件项包含插件的元信息和运行统计信息。 |
| has_more | Boolean | true | 是否已返回全部消息。 <br>  <br> * true：未返回全部信息，可再次调用此接口查看其他分页。 <br> * false：已返回全部信息。 |
### ProductPluginInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| metainfo | Object of [ProductMetainfo](#productmetainfo) | \ | 插件的详细信息。 |
| plugin_info | Object of [PluginInfo](#plugininfo) | {"heat":1500,"call_count":1000,"description":"该插件提供高质量的语音识别功能。","success_rate":0.95,"bots_use_count":7,"favorite_count":50,"total_tools_count":1,"avg_exec_duration_ms":120,"associated_bots_use_count":2} | 插件的运行数据和统计信息，包括热度、调用次数、成功率、收藏数等指标，以及被智能体使用的情况。 |
### ProductMetainfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| name | String | 语音识别插件 | 插件的名称。 |
| category | Object of [ProductCategory](#productcategory) | {"id":"732713727571473***","name":"社交"} | 插件所属的分类信息，包含分类 ID 和分类名称。 |
| icon_url | String | https://example.com/plugins/voice-recognition/icon.png | 插件的图标 URL。 |
| entity_id | String | 75464997639954*** | 插件 ID。 |
| listed_at | Long | 1757474787 | 插件上架时间，以 Unix 时间戳格式表示，单位为秒。 |
| paid_type | String | paid | 插件的付费类型，固定为 `paid`，即付费插件。 |
| user_info | Object of [ProductUserInfo](#productuserinfo) | {"user_id":32351795934732,"nick_name":"开发者小张","user_name":"developer_zhang","avatar_url":"https://example.com/img/user-avatar/e67e7ddd636a2087e79d624a64a***~300x300.image"} | 插件开发者的相关信息，包含开发者的用户 ID、昵称、用户名和头像 URL。 |
| product_id | String | 75462004326583*** | 商品的 ID，用于在插件商店中唯一标识该插件商品。 |
| description | String | 该插件提供高质量的语音识别功能。 | 插件的描述信息，用于说明插件的功能、用途和特点。 |
| entity_type | String | plugin | 实体类型，当前仅支持 `plugin`，表示插件。 |
| is_official | Boolean | true | 标识插件是否为官方发布。 <br>  <br> * `true` 表示该插件由扣子编程官方发布。 <br> * `false`表示由第三方开发者发布。 |
| entity_version | String | 0 | 插件的版本号，用于标识插件的当前版本。 |
### ProductCategory
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 732713727571473*** | 插件分类 ID。 |
| name | String | 社交 | 插件分类名称。 |
### ProductUserInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| user_id | Long | 32351795934732*** | 插件开发者的扣子用户 ID。 |
| nick_name | String | 开发者小张 | 插件开发者在扣子编程上的昵称。 |
| user_name | String | developer_zhang | 插件开发者在扣子编程上的用户名。 |
| avatar_url | String | https://example.com/img/user-avatar/e67e7ddd636a2087e79d624a64a***~300x300.image | 插件开发者在扣子编程上的头像 URL，用于展示开发者的头像信息。 |
### PluginInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| heat | Integer | 1500 | 插件的热度值，用于表示该插件的受欢迎程度或使用频率。数值越大表示热度越高。 |
| call_count | Long | 1000 | 插件的调用量，表示该插件的累计调用次数。 |
| description | String | 该插件提供高质量的语音识别功能。 | 插件的详细描述信息，用于说明插件的功能、用途和特点。 |
| success_rate | Double | 0.95 | 插件的调用成功率，以小数形式表示，数值范围为 0 到 1。 |
| bots_use_count | Long | 7 | 该插件在智能体或工作流中的累计关联次数。 |
| favorite_count | Integer | 50 | 插件的收藏量，表示该插件被用户收藏的总次数。 |
| is_call_available | Boolean | true | 标识该插件当前是否可被调用。 <br>  <br> * `true`：插件可正常调用。 <br> * `false`：插件暂时不可用，例如已下架。 |
| total_tools_count | Integer | 1 | 插件包含的工具总数。 |
| avg_exec_duration_ms | Double | 120 | 插件执行的平均耗时，单位为毫秒。 |
| associated_bots_use_count | Long | 2 | 当前扣子商店中关联了该插件的智能体数量。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/stores/plugins' \
--header 'Authorization: Bearer pat_3UXDHN82MgQANqK59BIAeH****' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": " ",
    "data": {
        "items": [
            {
                "metainfo": {
                    "name": "语音识别插件",
                    "category": {
                        "id": "7327137275714732069",
                        "name": "社交"
                    },
                    "icon_url": "https://example.com/plugins/voice-recognition/icon.png",
                    "entity_id": "75464997639954***",
                    "listed_at": 1757474787,
                    "paid_type": "paid",
                    "user_info": {
                        "user_id": 32351795934732,
                        "nick_name": "开发者小张",
                        "user_name": "developer_zhang",
                        "avatar_url": "https://example.com/img/user-avatar/e67e7ddd636a2087e79d624a64a***~300x300.image"
                    },
                    "product_id": "75462004326583***",
                    "description": "该插件提供高质量的语音识别功能。",
                    "entity_type": "plugin",
                    "is_official": true,
                    "entity_version": "0"
                },
                "plugin_info": {
                    "heat": 1500,
                    "call_count": 1000,
                    "description": "该插件提供高质量的语音识别功能。",
                    "success_rate": 0.95,
                    "bots_use_count": 7,
                    "favorite_count": 50,
                    "total_tools_count": 1,
                    "is_call_available": true,
                    "avg_exec_duration_ms": 120,
                    "associated_bots_use_count": 2
                }
            }
        ],
        "has_more": false
    },
    "detail": {
        "logid": "20241210152726467C48D89D6DB2****"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
