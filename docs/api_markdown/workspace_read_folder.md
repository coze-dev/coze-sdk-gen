# 查询文件夹详情
工作空间中的用户可以查询工作空间中指定文件夹的详情，包括文件夹的名称、描述、所属工作空间、文件夹创建者的 UID等。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/folders/:folder_id <br> ``` <br>  |
| **权限** | `readFolder` <br> 确保调用该接口使用的访问令牌开通了 `readFolder` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询文件夹详情。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| folder_id | String | 必选 | 7523161255335***1 | 文件夹 ID。你可以通过扣子编程页面获取文件夹 ID： <br> 进入指定工作空间，单击**项目管理**，单击目标文件夹，在 URL 中 `folder_id` 参数后的数字就是文件夹 ID。例如：`https://code.coze.cn/w/74982048832804***/projects?folder_id=75635667770837***`，`folder_id` 为 `75635667770837***`。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [FolderSimpleInfo](#foldersimpleinfo) | {"id":"1234567****","name":"项目文档","description":"存放2023年项目相关文档","workspace_id":"5123945629***","creator_user_id":"24787743932***","parent_folder_id":"1234567890***"} | 包含文件夹的详细信息，如文件夹 ID、名称、描述、所属工作空间 ID、创建者用户 UID 和父文件夹 ID。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的详细日志信息，用于问题排查和技术支持。 |
### FolderSimpleInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 1752316125533542*** | 文件夹 ID。 |
| name | String | 项目文档 | 文件夹的名称，用于标识和区分不同的文件夹。 |
| description | String | 存放2023年项目相关文档 | 文件夹的描述信息，用于提供关于文件夹的额外说明或备注。 |
| folder_type | String | development | 文件夹类型，当前仅支持 `development`。 |
| workspace_id | String | 5123945629*** | 文件夹所属的工作空间的 Space ID。Space ID 是空间的唯一标识。 |
| creator_user_id | String | 24787743932*** | 创建该文件夹的扣子用户 UID，用于标识文件夹的创建者。 |
| parent_folder_id | String | 754268556394029*** | 当前文件夹的父文件夹 ID，用于标识文件夹的层级关系。如果未返回该字段，则表示当前文件夹为根文件夹。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/folders/752316125533***' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
  "msg": "",
  "detail": {
    "logid": "20250707155659005CF0FD***"
  },
  "code": 0,
  "data": {
    "creator_user_id": "14232418519***",
    "parent_folder_id": "754268556394029***",
    "folder_type": "development",
    "name": "项目文件夹1",
    "description": "",
    "id": "752316131905***",
    "workspace_id": "7506047187433***"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。