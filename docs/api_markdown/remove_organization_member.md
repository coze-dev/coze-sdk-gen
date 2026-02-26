# 移除组织成员
移除组织成员，并指定资源接收人。
## 接口描述
如果成员因离职或换部门等场景需要调整归属的组织时，你可以调用本 API 将其移出组织。
移除组织成员时，扣子编程会自动将成员移出工作空间，并将其拥有的资源转移给指定的组织超级管理员。移除组织成员并不会将该成员从企业中删除，如果需要彻底从企业中删除，请调用[移除企业成员](https://docs.coze.cn/developer_guides/remove_enterprise_member) API 将该成员从企业中移除。
## 接口限制

* **套餐限制**：扣子企业版（企业标准版、企业旗舰版）。
* 被移除成员的资源只能转移给组织超级管理员。
* 移除组织超级管理员前，请确保该组织中已存在其他超级管理员，否则，移除时会提示 777074044 错误。

* 每次请求只能移除一位成员。如需移除多位成员，请依次发送请求。
* 该 API 不支持并发请求。

## 基础信息
| **请求方式** | DELETE |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/organizations/:organization_id/members/:user_id <br> ``` <br>  |
| **权限** | `Account.removeOrganizationPeople` <br> 确保调用该接口使用的访问令牌开通了**企业特权应用**中的 `Account.removeOrganizationPeople` 权限，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| **接口说明** | 移除组织成员。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| organization_id | String | 必选 | 74908881444562*** | 组织 ID。你可以在**组织管理** > **组织设置**页面查看对应的组织 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/b57d5e293ff944d2b83efe8217d35517~tplv-goo7wpa0wc-image.image) |
| user_id | String | 必选 | 41135614833**** | 需要移除的扣子用户的 UID。 <br> 你可以调用火山引擎的 [ListCozeUser-成员列表](https://api.volcengine.com/api-docs/view?serviceCode=coze&version=2025-06-01&action=ListCozeUser)API，查看 `CozeUserId` 的值即为扣子用户 UID。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| receiver_user_id | String | 必选 | 411356148551**** | 接收被移除成员资源的扣子用户 UID，需填写**组织超级管理员**的用户 UID。被移除成员的工作空间、智能体、工作流等资源将转移给该用户。 <br> 获取组织超级管理员用户 UID 的步骤如下： <br>  <br> 1. 在**组织管理** > **组织成员管理**页面查看组织的超级管理员信息。 <br> 2. 组织的超级管理员在扣子编程左下角单击头像，选择**账号设置**，查看账号的 UID。 |
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
curl --location --request DELETE 'https://api.coze.cn/v1/organizations/74908881444562***/members/411356148551****' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type : application/json' \
--data-raw '{
    "receiver_user_id": "411379148551****"
}'
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