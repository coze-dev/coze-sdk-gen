# 添加企业成员
添加员工到企业。
在火山引擎创建用户后，**默认会自动将用户添加至企业**，若未成功添加，你可以调用本 API 将用户添加至企业。火山引擎创建用户的具体方法请参见[成员管理](https://docs.coze.cn/developer_guides/create_coze_user)。
## 接口限制

* **套餐限制**：扣子企业版（企业标准版、企业旗舰版）。
* 本 API 仅支持添加员工（火山子用户），不支持添加外部成员（访客）。
* 添加成员总数不能超过企业标准版权益中的成员数量上限（100 个成员），否则会提示 777074011错误。

* 每次请求只能添加一位成员。如需添加多位，请依次发送请求。
* 该 API 不支持并发请求。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/enterprises/:enterprise_id/members <br> ``` <br>  |
| **权限** | `Enterprise.batchAddPeople` <br> 确保调用该接口使用的访问令牌开通了 `Enterprise.batchAddPeople` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 添加企业成员。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| enterprise_id | String | 必选 | volcano_210195***  | 企业 ID，用于标识用户所属的企业。 <br> 你可以在组织管理 > 组织设置页面查看企业 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/02db2078f0c84bc2aa189f5cca93d49d~tplv-goo7wpa0wc-image.image) |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| users | Array of [OpenEnterpriseMember](#openenterprisemember) | 可选 | [ { "user_id": "247877439325****", "role": "enterprise_member" } ] | 待邀请加入企业的用户列表，单次最多添加 `1`个成员。每个成员包含用户 ID 和角色信息。 |
### OpenEnterpriseMember
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| role | String | 必选 | enterprise_member | 成员在企业中的角色，枚举值： <br>  <br> * `enterprise_admin`：企业管理员。 <br> * `enterprise_member`：企业普通成员。 |
| user_id | String | 必选 | 247877439325**** | 需要添加至企业的扣子用户 UID。 <br> 你可以调用火山引擎的 [ListCozeUser-成员列表](https://api.volcengine.com/api-docs/view?serviceCode=coze&version=2025-06-01&action=ListCozeUser) API，查看 `CozeUserId` 的值即为扣子用户 UID。 |

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
curl --location --request POST 'https://api.coze.cn/v1/enterprises/volcano_210195*** /members' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type : application/json' \
--data-raw '{
    "users": [
        {
            "user_id": "247877439325****",
            "role": "enterprise_member"
        }
    ]
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

