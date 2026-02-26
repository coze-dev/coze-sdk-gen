# 修改组织基本信息
修改指定组织的基本信息，包括组织名称和描述。
## 接口限制

* **套餐限制**：仅扣子企业旗舰版支持。
* 组织的名称和描述需要符合内容安全规范，不得包含涉政、涉黄等违规内容，否则扣子编程会提示 4102 错误。

## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/organizations/:organization_id <br> ``` <br>  |
| **权限** | `Account.updateOrganization` <br> 确保调用该接口使用的访问令牌开通了 `Account.updateOrganization` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 修改组织基本信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| organization_id | String | 必选 | 7559861372637*** | 待修改组织信息的组织 ID。  <br> 你可以在**组织管理** > **组织设置**页面查看对应的组织 ID，或通过[查询组织列表](https://docs.coze.cn/developer_guides/list_organization) API 查询组织 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/b57d5e293ff944d2b83efe8217d35517~tplv-goo7wpa0wc-image.image) |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| name | String | 可选 | 研发部 | 修改后的组织名称。默认为空，表示不修改。 <br> 最大长度为 30 个字符。 |
| description | String | 可选 | 研发部内部使用的组织 | 修改后的组织描述。默认为空，表示不修改。 <br> 最大长度为 100 个字符。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的日志信息，用于问题排查和技术支持。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request PUT 'https://api.coze.cn/v1/organizations/7559861372637***' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
--data-raw '{
    "name": "研发部",
    "description": "研发部内部使用的组织"
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