# 查询终端用户的权益配额
查询已配置的权益配额信息，包括权益配额、状态、激活模式等详细信息。
* **套餐限制**：扣子企业旗舰版。
* **角色限制**：企业超级管理员和管理员可以调用该 API。
* 调用此 API 查询设备权益配额之前，需要确保企业下的设备已成功上报了设备信息，否则会导致权益配额对该设备无法生效，设备信息的配置方法可参考[设置设备信息](https://docs.coze.cn/dev_how_to_guides/deviceInfo)。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/commerce/benefit/limitations <br> ``` <br>  |
| **权限** | `listBenefitLimitation` <br> 确保调用该接口使用的访问令牌开通了 `listBenefitLimitation` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询已配置的权益额度信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 用于指定解析请求正文的格式，表明请求体为 JSON 格式数据。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| entity_type | String | 必选 | enterprise_all_devices | 权益配额的生效范围，枚举值： <br>  <br> * `enterprise_all_devices`：权益配额对企业下的所有设备生效。 <br> * `enterprise_all_custom_consumers`：权益配额对企业下所有自定义维度的实体生效。 <br> * `single_device`：权益配额对单个设备生效。 <br> * `single_custom_consumer`：权益配额对单个自定义维度的实体生效。 |
| entity_id | String | 可选 | SN12345********* | 该权益配额对哪个实体 ID 生效。 <br>  <br> * 仅在 `entity_type` 为 `single_device` 或  <br>  <br> `single_custom_consumer` 时，需填写entity_id，分别为对应的 Device ID 或 custom consumer ID。 <br>  <br> * `entity_type` 为其他类型时，无需填写entity_id。 |
| benefit_type | String | 必选 | resource_point | 权益配额类型，枚举值： <br>  <br> * `resource_point`：资源点配额。 <br> * `voice_unified_duration_system`： <br>  <br> 语音通话时长配额（系统音色）。 <br>  <br> * `voice_unified_duration_custom`： <br>  <br> 语音通话时长配额（复刻音色）。 |
| status | String | 可选 | valid | 设备权益的当前状态，枚举值： <br>  <br> * `valid`：（默认值）正常使用。 <br> * `frozen`：已冻结。 |
| page_token | String | 可选 | - | 翻页标识，表示下一页的起始位置。当查询结果超过 `page_size` 时，返回的 `page_token`可用于获取下一页数据。 <br> 首次请求不填或置空，后续翻页需带上上一次返回的 page_token。 |
| page_size | Integer | 可选 | 30 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1~200，默认为 20。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [ListBenefitLimitationData](#listbenefitlimitationdata) | { "benefit_id": 123***, "benefit_type": "resource_point", "active_mode": "absolute_time", "started_at": 1741708800, "ended_at": 1741708800, "limit": 100, "status": "valid" } | 接口调用成功时，返回的详细数据信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | - | 响应详情信息。 |
### ListBenefitLimitationData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| has_more | Boolean | true  | 是否还有下一页。 <br>  <br> * true：还有下一页。 <br> * false：没有下一页。 |
| page_token | String | - | 翻页标识。 |
| benefit_infos | Array of [BenefitInfo](#benefitinfo) | - | 权益配额列表。 |
### BenefitInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| limit | Long | 100 | 可用的额度。 <br>  <br> * 如果权益配额类型为 `resource_point`，可用额度的单位为资源点。 <br> * 如果权益配额类型为`voice_unified_duration_system`或`voice_unified_duration_custom`，可用额度的单位为秒。 |
| benefit_type | String | resource_point | 权益配额类型，枚举值： <br>  <br> * `resource_point`：资源点配额。 <br> * `voice_unified_duration_system`： <br>  <br> 语音通话时长配额（系统音色）。 <br>  <br> * `voice_unified_duration_custom`： <br>  <br> 语音通话时长配额（复刻音色）。 <br> * 增购 **AI 智能通话许可（系统音色）​**的企业，配置的 `voice_unified_duration_system` 才生效，购买语音通话时长的详细步骤请参见[音视频费用](https://docs.coze.cn/coze_pro/asr_tts_fee)。 <br> * 增购 **AI 智能通话许可（复刻音色）​**的企业，配置的  <br>  <br> `voice_unified_duration_custom` 才生效。 <br>  <br> * 未增购 AI 智能通话许可的企业，仅支持配置  <br>  <br> `resource_point`。 <br>  |
| active_mode | String | absolute_time | 激活模式，当前仅支持设置为 `absolute_time` 模式，即绝对时间。 <br> 该模式下，权益生效时间由 `started_at` 和 `ended_at` 两个时间确定。 |
| started_at | Long | 1753996800 | 该条配额规则的生效起始时间，Unixtime 时间戳格式，单位为秒。 |
| ended_at | Long | 253402300799 | 该条配额规则的生效截止时间，Unixtime 时间戳格式，单位为秒。过期后，该条配额规则会失效。 |
| entity_id | String | SN12345********* | 该权益配额对哪个实体 ID 生效。仅在 `entity_type` 为 `single_device`或  <br> `single_custom_consumer` 时，会返回对应的实体 ID。 |
| status | String | valid | 设备权益的当前状态，枚举值： <br>  <br> * `valid`：正常使用。 <br> * `frozen`：已冻结。 |
| entity_type | String | enterprise_all_devices | 权益配额的生效范围，枚举值： <br>  <br> * `enterprise_all_devices`：权益配额对企业下的所有设备生效。 <br> * `enterprise_all_custom_consumers`：权益配额对企业下所有自定义维度的实体生效。若某设备在设备信息中未上报 custom_consumers，则该设备无法生效权益配额。 <br> * `single_device`：权益配额对单个设备生效。 <br> * `single_custom_consumer`：权益配额对单个自定义维度的实体生效。 |
| trigger_unit | String | day | 权益可用额度的重置周期，即额度按指定时间间隔恢复或重新计算。枚举值： <br>  <br> * `never`：（默认）不重置额度，适用于设置累计可用额度上限。 <br> * `minute`：以分钟为周期重置额度。 <br> * `hour`：以小时为周期重置额度。 <br> * `day`：以天为周期重置额度。 |
| trigger_time | Long | 1 | 权益配额重置周期的频率： <br>  <br> * 当 `trigger_unit` 为 `never` 时，`trigger_time` 的值为 1，且无意义。 <br> * 当 `trigger_unit` 为 `minute`、`hour` 或 `day` 时，`trigger_time` 表示具体的刷新频率，例如：`trigger_unit=day` 且 `trigger_time = 1`，表示每日刷新配额。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/commerce/benefit/limitations?entity_type=enterprise_all_devices&entity_id=SN12345*********&benefit_type=resource_point' \
--header 'Authorization : Bearer pat_OYDacMzM3WyOWV3P****' \
--header 'Content-Type : application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "has_more": true,
    "page_token": "next_page_token_123",
    "benefit_infos": [
      {
        "limit": 100,
        "status": "valid",
        "ended_at": 1804185600,
        "entity_id": "SN12345*********",
        "started_at": 1753996800,
        "active_mode": "absolute_time",
        "entity_type": "enterprise_all_devices",
        "benefit_type": "resource_point",
        "trigger_time": 1,
        "trigger_unit": "never"
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