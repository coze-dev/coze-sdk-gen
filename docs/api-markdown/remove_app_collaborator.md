# 删除应用的协作者
删除低代码应用的协作者。
删除协作者时，扣子编程会将该协作者创建的工作流、插件等资源转移给应用的所有者。
## 接口限制

* 主账号内的所有子账号共享同一 API 的流控额度，单个 API 的流控限制为 5 QPS。
* 每次请求只能删除一位协作者。如需删除多位，请依次发送请求。
* 不支持渠道类型 OAuth 应用。使用 OAuth JWT 应用和服务访问令牌时，只需要有对应权限点即可。其余认证方式，只有**应用的所有者和协作者**有权删除。

## 基础信息
| **请求方式** | DELETE |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/apps/:app_id/collaborators/:user_id <br> ``` <br>  |
| **权限** | `Project.removeCollaborator` <br> 确保调用该接口使用的访问令牌开通了 `Project.removeCollaborator` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 删除低代码应用的协作者。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| app_id | String | 必选 | 75353861140**** | 待删除协作者的低代码应用的 ID。 <br> 进入低代码应用编排页面，页面 URL 中 `project-ide` 参数后的数字就是应用 ID。例如`https://www.coze.cn/space/7532329396037500982/project-ide/7535386114057***`，应用 ID 为 `7535386114057***`。 |
| user_id | String | 必选 | 411479148551**** | 待删除协作者的扣子用户 UID。 <br> 在扣子编程左下角单击头像，选择**账号设置**，查看账号的 UID。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request DELETE 'https://api.coze.cn/v1/apps/75353861140****/collaborators/411479148551****' \
--header 'Authorization : Bearer pat_Osa******' \
--header 'Content-Type : application/json' \
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210152726467C48D89D6DB2****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
