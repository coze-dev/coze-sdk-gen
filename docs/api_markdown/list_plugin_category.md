# 查询插件分类
查询插件分类信息。
你可以通过本 API 获取插件的分类 ID，后续调用[查询插件列表](https://docs.coze.cn/developer_guides/list_plugin)  API 查询指定分类下的插件。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/stores/categories <br> ``` <br>  |
| **权限** | 无 |
| **接口说明** | 查询插件分类信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| entity_type | String | 必选 | plugin | 实体类型，当前仅支持 `plugin`，表示插件。 |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即返回第一页数据。 |
| page_size | Integer | 可选 | 20 | 每页返回的数据条数，用于分页查询。默认值为 20，最大值为 100。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [ListCategoryData](#listcategorydata) | {"items":[{"id":"732713727571473***","name":"社交"}],"has_more":false} | 返回插件分类列表及相关分页信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的详细日志信息，用于问题排查和技术支持。 |
### ListCategoryData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [ProductCategory](#productcategory) | { "id": "7475930039976984614", "name": "智能硬件" } | 插件分类列表。 |
| has_more | Boolean | false | 是否已返回全部消息。 <br>  <br> * true：未返回全部数据，可再次调用此接口查看其他分页。 <br> * false：已返回全部数据。 |
### ProductCategory
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 732713727571473*** | 插件分类 ID。 |
| name | String | 社交 | 插件分类名称。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/stores/categories?entity_type=plugin' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": "",
    "detail": {
        "logid": "2025092517382882216851350ED***"
    },
    "data": {
        "has_more": false,
        "items": [
            {
                "id": "7475930039976984614",
                "name": "智能硬件"
            },
            {
                "id": "7327137275714732069",
                "name": "新闻阅读"
            },
            {
                "id": "7327137275714781221",
                "name": "便利生活"
            },
            {
                "id": "7327137275714748453",
                "name": "图像"
            },
            {
                "id": "7327137275714764837",
                "name": "实用工具"
            },
            {
                "id": "7327137275714797605",
                "name": "网页搜索"
            },
            {
                "id": "7327137275714813989",
                "name": "科学与教育"
            },
            {
                "id": "7327137275714830373",
                "name": "社交"
            }
        ]
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。