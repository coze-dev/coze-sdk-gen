# 查询文件夹列表
工作空间中的用户可以查询工作空间中的文件夹列表。
## 接口描述
你可以查询某个文件夹或工作空间根目录下的子文件夹列表。每次查询仅返回下一层级的文件夹信息，不包含更深层级的子文件夹。若需要获取更深层级的文件夹信息，你可以多次调用本 API，利用上一次查询返回的文件夹 ID 继续查询。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/folders <br> ``` <br>  |
| **权限** | `listFolder` <br> 确保调用该接口使用的访问令牌开通了 `listFolder` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询工作空间中的文件夹列表。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workspace_id | String | 必选 | 736163827687053**** | 待查询文件夹列表的工作空间 ID。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，工作空间 ID 为 `75814654762959***`。 |
| folder_type | String | 必选 | development | 文件夹类型，参数值固定为 `development`。 |
| parent_folder_id | String | 可选 | 752430442263*** | 指定要查询的文件夹的父级文件夹 ID。当需要查看某个文件夹下的子文件夹列表时，需传入该文件夹的 ID。若不传或传入 `0`，则表示查询工作空间根目录下的文件夹。 <br> 文件夹 ID 的获取方法如下：进入指定工作空间，单击**项目管理**，单击目标文件夹，在 URL 中 `folder_id` 参数后的数字就是文件夹 ID。例如：`https://code.coze.cn/w/74982048832804***/projects?folder_id=75635667770837***`，`folder_id` 为 `75635667770837***`。 |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即返回第一页数据。 |
| page_size | Integer | 可选 | 20 | 每页返回的数据条数，用于分页查询。默认值为 `20`，最大值为 `50`。 |

## 返回参数

| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenGetSpaceFolderData](#opengetspacefolderdata) | {"items":[{"id":"1234567****","name":"项目文档","description":"存放2023年项目相关文档","workspace_id":"5123945629***","children_count":0,"creator_user_id":"24787743932***","parent_folder_id":"1234567890***"}],"has_more":false,"total_count":10} | 文件夹列表的查询结果，包含文件夹的详细信息、是否还有更多数据以及文件夹的总数量。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的详细日志信息，用于问题排查。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 `logid`联系扣子团队获取帮助。 |
### OpenGetSpaceFolderData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [FolderSimpleInfo](#foldersimpleinfo) | \ | 查询的父文件夹下的文件夹列表，仅展示下一层级文件夹信息，不包含更深层级的子文件夹的信息。 |
| has_more | Boolean | true | 标识当前返回的文件夹列表是否还有更多数据未返回。 <br>  <br> * `true` ：还有更多未返回的文件夹。 <br> * `false`：已返回所有数据。 |
| total_count | Long | 10 | 当前查询条件下第一层级的文件夹总数，不包含更深层级的子文件夹。 |
### FolderSimpleInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 1752316125533542*** | 文件夹 ID。 |
| name | String | 项目文档 | 文件夹的名称，用于标识和区分不同的文件夹。 |
| description | String | 存放2023年项目相关文档 | 文件夹的描述信息，用于提供关于文件夹的额外说明或备注。 |
| folder_type | String | development | 文件夹类型，当前仅支持 `development`。 |
| workspace_id | String | 5123945629*** | 文件夹所属的工作空间的 Space ID。Space ID 是空间的唯一标识。 |
| children_count | Integer | 5 | 当前文件夹下的子文件夹数量，仅统计下一层级文件夹数量，不包含更深层级的子文件夹。 |
| creator_user_id | String | 24787743932*** | 创建该文件夹的扣子用户 UID，用于标识文件夹的创建者。 |
| parent_folder_id | String | 754268556394029*** | 当前文件夹的父文件夹 ID，用于标识文件夹的层级关系。如果未返回该字段，则表示当前文件夹为根文件夹。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |


## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/folders?folder_type=development&workspace_id=75060471874****' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "data": {
    "total_count": 2,
    "has_more": false,
    "items": [
      {
        "name": "项目文件夹2",
        "description": "",
        "id": "7523161255335***",
        "workspace_id": "7506047187433***",
        "creator_user_id": "14232418519***",
        "folder_type": "development",
        "parent_folder_id": "752316125533***",
        "children_count": 0
      },
      {
        "name": "项目文件夹1",
        "description": "存放扣子编程项目的文件",
        "id": "7523161237887238182",
        "workspace_id": "7506047187***",
        "creator_user_id": "14232418***",
        "folder_type": "development",
        "parent_folder_id": "752316125533****",
        "children_count": 0
      }
    ]
  },
  "msg": "",
  "detail": {
    "logid": "0217517972369490000000000000****"
  },
  "code": 0
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

