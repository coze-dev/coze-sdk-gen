# 创建终端用户权益额度
创建终端用户的权益额度。
* **套餐限制**：扣子企业旗舰版。
* **角色限制**：企业超级管理员和管理员可以调用该 API。
* 调用此 API 创建设备权益额度之前，需要确保企业下的设备已成功上报了设备信息，否则会导致权益额度对该设备无法生效，设备信息的配置方法可参考[设置设备信息](https://docs.coze.cn/dev_how_to_guides/deviceInfo)。

## 接口描述
你可以调用本 API 创建终端用户的权益额度，用于设置终端用户可使用的资源点或语音时长配额，以便用户在初始阶段免费体验设备功能，同时避免资源的过度使用。

* 生效对象：权益额度的生效对象可以设置为企业中的所有设备、所有自定义维度的实体、某个设备或某个自定义维度的实体。
* 额度类型：支持按资源点维度或语音通话时长维度设置配额，你可以为每个设备设置累计可用额度以及时间周期内的额度（例如每日 1000 资源点）。当设备在当前周期内的资源点使用达到周期上限或累计额度上限时，将无法继续使用，直至下一个周期或额度重置。例如设置每个设备累计额度为 5000 资源点，每天可用 1000 资源点。当设备 A 今天的使用资源点达到 1000 上限后，设备 A 今天将无法继续使用，需等到次日才能恢复使用，当设备 A 的累计使用资源点达到 5000 上限后，设备 A 将无法继续使用。
* 配置的优先级：若同时为单个设备和企业下所有设备配置了额度，则优先使用单个设备的额度。
* 生效时间：通过扣子编程的[管理设备配额](https://docs.coze.cn/dev_how_to_guides/device_usage#75617e0a)页面设置的针对企业下所有设备和自定义维度实体的额度，默认有效期为永不过期，即 `started_at = 1970-01-01` 至 `ended_at = 9999-12-31`。若权益生效时间已过期，则该条配额规则会失效。
* 未配置额度的设备：未设置权益额度的设备，将默认无限制使用资源点。

## 接口限制

* 当权益配额的生效范围为企业下所有设备或企业下所有自定义维度的实体时，仅能创建一条累计可用额度和一条时间周期内可用额度。
* 增购 **AI 智能通话许可（系统音色）​**的企业，权益额度类型支持配置为**语音通话时长配额（系统音色）**，否则不生效，购买语音通话时长的详细步骤请参见[音视频费用](https://docs.coze.cn/coze_pro/asr_tts_fee)。
* 增购 **AI 智能通话许可（复刻音色）​**的企业，权益额度类型支持配置为**语音通话时长配额（复刻音色）**，否则不生效。
* 未增购 AI 智能通话许可的企业，权益额度类型仅支持资源点配额。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/commerce/benefit/limitations <br> ``` <br>  |
| **权限** | `createBenefitLimitation` <br> 确保调用该接口使用的访问令牌开通了 `createBenefitLimitation` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 创建终端用户的权益配额。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 用于指定解析请求正文的格式，表明请求体为 JSON 格式数据。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| entity_type | String | 必选 | enterprise_all_devices | 权益配额的生效范围，枚举值： <br>  <br> * `enterprise_all_devices`：权益配额对企业下的所有设备生效。 <br> * `enterprise_all_custom_consumers`：权益配额对企业下所有自定义维度的实体生效。若某设备在设备信息中未上报 custom_consumers，则该设备无法生效权益额度。 <br> * `single_device`：权益配额对单个设备生效。 <br> * `single_custom_consumer`：权益配额对单个自定义维度的实体生效。 |
| entity_id | String | 可选 | 12345 | 该权益配额对哪个实体 ID 生效。 <br>  <br> * `entity_type` 为 `single_device` 时，`entity_id` 需要设置为对应的 Device ID。 <br> * `entity_type` 为 `single_custom_consumer` 时，`entity_id` 需要设置为对应的 custom consumer ID。 <br> * `entity_type` 为其他类型时，无需填写 `entity_id`。 |
| benefit_info | Object of [BenefitInfo](#benefitinfo) | 必选 | { "benefit_type": "resource_point", "active_mode": "absolute_time", "started_at": 1741708800, "ended_at": 253402300799, "limit": 100, "status": "valid" } | 权益额度。 |
### BenefitInfo
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| limit | Long | 必选 | 100 | 可用的额度。 <br>  <br> * 如果权益配额类型为 `resource_point`，可用额度的单位为资源点。 <br> * 如果权益配额类型为`voice_unified_duration_system`或`voice_unified_duration_custom`，可用额度的单位为秒。 |
| benefit_type | String | 必选 | resource_point | 权益配额类型，枚举值： <br>  <br> * `resource_point`：资源点配额。 <br> * `voice_unified_duration_system`： <br>  <br> 语音通话时长配额（系统音色）。 <br>  <br> * `voice_unified_duration_custom`： <br>  <br> 语音通话时长配额（复刻音色）。 <br> * 增购 **AI 智能通话许可（系统音色）​**的企业，配置的 `voice_unified_duration_system` 才生效，购买语音通话时长的详细步骤请参见[音视频费用](https://docs.coze.cn/coze_pro/asr_tts_fee)。 <br> * 增购 **AI 智能通话许可（复刻音色）​**的企业，配置的  <br>  <br> `voice_unified_duration_custom` 才生效。 <br>  <br> * 未增购 AI 智能通话许可的企业，仅支持配置  <br>  <br> `resource_point`。 <br>  |
| active_mode | String | 必选 | absolute_time | 激活模式，当前仅支持设置为 `absolute_time` 模式，即绝对时间。 <br> 该模式下，权益生效时间由 `started_at` 和 `ended_at` 两个时间确定。 |
| started_at | Long | 必选 | 1753996800 | 该条配额规则的生效起始时间，Unixtime 时间戳格式，单位为秒。 |
| ended_at | Long | 必选 | 253402300799 | 该条配额规则的生效截止时间，Unixtime 时间戳格式，单位为秒。过期后，该条配额规则会失效。 |
| status | String | 可选 | valid | 设备权益的当前状态，枚举值： <br>  <br> * `valid`：正常使用。 <br> * `frozen`：已冻结。 |
| trigger_unit | String | 可选 | day | 权益可用额度的重置周期，即额度按指定时间间隔恢复或重新计算。枚举值： <br>  <br> * `never`：（默认）不重置额度，适用于设置累计可用额度上限。 <br> * `minute`：以分钟为周期重置额度。 <br> * `hour`：以小时为周期重置额度。 <br> * `day`：以天为周期重置额度。 |
| trigger_time | Long | 可选 | 1 | 权益配额重置周期的频率： <br>  <br> * 当 `trigger_unit` 为 `never` 时，`trigger_time` 的值为 1，且无意义。 <br> * 当 `trigger_unit` 为 `minute`、`hour` 或 `day` 时，`trigger_time` 表示具体的刷新频率，例如：`trigger_unit=day` 且 `trigger_time = 1`，表示每日刷新配额。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [CreateBenefitLimitationData](#createbenefitlimitationdata) | { "benefit_id": 123***, "benefit_type": "resource_point", "active_mode": "absolute_time", "started_at": 1741708800, "ended_at": 1741708800, "limit": 100, "status": "valid" } | 接口调用成功时，返回的详细数据信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | - | 响应详情信息。 |
### CreateBenefitLimitationData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| benefit_info | Object of [BenefitInfo](#benefitinfo) | - | 权益额度。 |
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
| benefit_id | String | 123 | 权益配额的 ID。 |
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
curl --location --request POST 'https://api.coze.cn/v1/commerce/benefit/limitations' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
{
  "entity_type": "single_device",
  "entity_id": "SN12345*********",
  "benefit_info": {
      "benefit_type": "resource_point",
      "active_mode": "absolute_time",
      "started_at": 1741708800,
      "ended_at": 253402300799,
      "limit": 100,
      "status": "valid"
  }
}
```

### 返回示例
```JSON
{ 
    "data": {
        "benefit_id": 123***,
        "benefit_type": "resource_point",
        "active_mode": "absolute_time",
        "started_at": 1741708800,
        "ended_at": 253402300799,
        "limit": 100,
        "status": "valid"
    }, 
    "code": 0, 
    "msg": "" 
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

