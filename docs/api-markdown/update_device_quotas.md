# 更新终端用户的权益额度
更新终端用户的权益额度。当你[创建设备的权益额度](https://docs.coze.cn/developer_guides/create_device_quotas)后，可以调用本 API 更新相关额度。
* **套餐限制**：扣子企业旗舰版。
* **角色限制**：企业超级管理员和管理员可以调用该 API。
* 调用此 API 更新设备权益额度之前，需要确保企业下的设备已成功上报了设备信息，否则会导致权益额度对该设备无法生效，设备信息的配置方法可参考[设置设备信息](https://docs.coze.cn/dev_how_to_guides/deviceInfo)。

## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/commerce/benefit/limitations/:benefit_id <br> ``` <br>  |
| **权限** | `updateBenefitLimitation` <br> 确保调用该接口使用的访问令牌开通了 `updateBenefitLimitation` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 更新终端用户的权益额度。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 用于指定解析请求正文的格式，表明请求体为 JSON 格式数据。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| benefit_id | String | 可选 | 123 | 权益 ID。你可以在[创建设备的权益额度](https://docs.coze.cn/developer_guides/create_device_quotas)的返回参数中获取对应的权益 ID 的值。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| active_mode | String | 可选 | absolute_time | 激活模式，当前仅支持设置为 `absolute_time` 模式，即绝对时间。 <br> 该模式下，权益生效时间由 `started_at` 和 `ended_at` 两个时间确定。 |
| started_at | Long | 可选 | 1743213600 | 该条配额规则的生效起始时间，Unixtime 时间戳格式，单位为秒。 |
| ended_at | Long | 可选 | 1743386400 | 该条配额规则的生效截止时间，Unixtime 时间戳格式，单位为秒。过期后，该条配额规则会失效。 |
| limit | Long | 可选 | 100 | 可用的资源点额度。 |
| status | String | 可选 | valid | 设备权益的当前状态，枚举值： <br>  <br> * valid：（默认值）正常使用。 <br> * frozen：已冻结。 |
| trigger_unit | String | 可选 | day | 权益可用额度的重置周期，即额度按指定时间间隔恢复或重新计算： <br>  <br> * never：不重置额度，适用于设置累计可用额度上限。 <br> * minute：以分钟为周期重置额度。 <br> * hour：以小时为周期重置额度。 <br> * day：以天为周期重置额度。 |
| trigger_time | Long | 可选 | 1 | 权益额度重置周期的频率： <br>  <br> * 当 `trigger_unit` 为 `never` 时，`trigger_time` 的值为 1，且无意义。 <br> * 当 `trigger_unit` 为 `minute`、`hour` 或 `day` 时，`trigger_time` 表示具体的刷新频率，例如：`trigger_unit=day` 且 `trigger_time = 1`，表示每日刷新配额。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | - | 响应详情信息。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request PUT 'https://api.coze.cn/v1/commerce/benefit/limitations/12345***' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
{
  "benefit_id": "12345***",
  "active_mode": "absolute_time",
  "started_at": 1741708800,
  "ended_at": 1741708800,
  "limit": 100,
  "status": "valid"
}
```

### 返回示例
```JSON
{
  "data": {
    "benefit_id": "12345***",
    "benefit_type": "resource_point",
    "active_mode": "absolute_time",
    "started_at": 1741708800,
    "ended_at": 1741708800,
    "limit": 100,
    "status": "valid"
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

