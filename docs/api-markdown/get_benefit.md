# 查询套餐权益
查询当前账号的套餐权益信息。
## 接口描述
你可以通过本 API 查询当前账号的以下套餐权益信息：

* 所属的套餐类型。
* 扩容管理页面中 API 扩容的信息，包括套餐默认的 API QPS、扩容新增的 API QPS 额度，以及当前实际生效的 API QPS 额度。
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/720efb1da1e54938ae626248aebd08c3~tplv-goo7wpa0wc-image.image)
* 套餐权益内通过 MCP 方式调用付费插件的次数限制。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/commerce/benefit/benefits/get <br> ``` <br>  |
| **权限** | `getBenefit` <br> 确保调用该接口使用的访问令牌开通了 `getBenefit` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询套餐权益。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| benefit_type_list | Array of String | 可选 | api_run_qps,call_tool_limit | 权益类型列表，多个类型用逗号分隔。支持的权益类型如下： <br>  <br> * `api_run_qps`： API 扩容的信息，你需要在 `resource_id` 中传入待查询的 API 类型。 <br> * `call_tool_limit`：通过 MCP 方式调用付费插件的次数限制。 <br>  <br> 默认为空，即返回订阅的套餐类型，不含额外扩容的权益。 |
| resource_id | String | 可选 | plugin | API 类型，当 `benefit_type_list `为 `api_run_qps`时，需要配置该参数。当前仅支持查询可扩容的 API 类型。枚举值： <br>  <br> * `plugin`：插件相关 API 的 QPS 限制。 <br> * `chat`：对话相关的 API 的 QPS 限制。 <br> * `workflow`：工作流相关的 API 的 QPS 限制。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| data | Object of [BenefitData](#benefitdata) | \ | 用户订阅的套餐类型和权益额度的详细信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
### BenefitData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| basic_info | Object of [BasicInfo](#basicinfo) | {"user_level":"enterprise"} | 用户订阅的扣子套餐类型。 |
| benefit_info | Array of [BenefitInfo](#benefitinfo) | [{"basic":{"status":"valid","item_info":{"used":100,"total":1000,"end_at":1735689600,"start_at":1735689600,"strategy":"by_quota"}},"extra":[{"status":"valid","item_info":{"used":50,"total":500,"end_at":1735689600,"start_at":1735689600,"strategy":"by_quota"}}],"effective":{"status":"valid","item_info":{"used":200,"total":2000,"end_at":1735689600,"start_at":1735689600,"strategy":"by_quota"}},"resource_id":"plugin","benefit_type":"resource_point"}] | 权益额度，包含用户当前享有的各项权益详情，如默认的权益、扩容权益及实际生效的权益等。 |
### BasicInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| user_level | String | enterprise | 用户订阅的扣子套餐类型，枚举值： <br>  <br> * `free`：个人免费版。 <br> * `pro_personal`：个人付费版。 <br> * `team`：企业标准版。 <br> * `enterprise`：企业旗舰版。 |
### BenefitInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| basic | Object of [BenefitTypeInfoItem](#benefittypeinfoitem) | {"status":"valid","item_info":{"used":100,"total":1000,"end_at":1735689600,"start_at":1735689600,"strategy":"by_quota"}} | 默认的权益信息，包含权益的状态、用量及权益配额 ID。 |
| extra | Array of [BenefitTypeInfoItem](#benefittypeinfoitem) | [{"status":"valid","item_info":{"used":50,"total":500,"end_at":1735689600,"start_at":1735689600,"strategy":"by_quota"}] | 资源扩容新增的权益额度。如果未购买扩容服务，则该值为空。 |
| effective | Object of [BenefitTypeInfoItem](#benefittypeinfoitem) | {"status":"valid","item_info":{"used":200,"total":2000,"end_at":1735689600,"start_at":1735689600,"strategy":"by_quota"}} | 当前实际生效的权益额度。 |
| resource_id | String | 753754678968821*** | 当前权益对应的资源 ID。当前仅如下 API 类型支持扩容。枚举值： <br>  <br> * `plugin`：插件相关 API 的 QPS 限制。对应的 API 包括[调用插件工具](https://docs.coze.cn/developer_guides/call_plugin_tool)、[询插件详情](https://docs.coze.cn/developer_guides/get_plugin)。 <br> * `chat`：对话相关 API 的 QPS 限制。对应的 API 包括：[发起对话](https://docs.coze.cn/developer_guides/chat_v3)、[查看对话详情](https://docs.coze.cn/developer_guides/retrieve_chat)、[查看对话消息详情](https://docs.coze.cn/developer_guides/list_chat_messages)。 <br> * `workflow`：工作流相关 API 的 QPS 限制。对应的 API 包括：[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) 、[执行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_stream_run) 、[执行对话流](https://docs.coze.cn/developer_guides/workflow_chat) 、[查询工作流异步执行结果](https://docs.coze.cn/developer_guides/workflow_history) 、[查询输出节点的执行结果](https://docs.coze.cn/developer_guides/get_node_execute_history_response) 。 |
| benefit_type | String | resource_point | 权益配额类型，枚举值： <br>  <br> * `api_run_qps`：扩容管理页面中扩容的 API QPS。 <br> * `call_tool_limit`：通过 MCP 方式调用付费插件的次数限制。 |
### BenefitTypeInfoItem
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| status | String | valid | 权益状态，表示当前权益的使用状态。枚举值： <br>  <br> * `valid`：正常使用。 <br> * `frozen`：冻结使用。 <br> * `cancel`：取消。 <br> * `pending`：待生效。 <br> * `invalid`：不可用。 <br> * `auditing`：审核中。 <br> * `expired`：已过期。 |
| item_info | Object of [CommonCounter](#commoncounter) | {"used":100,"total":1000,"end_at":1735689600,"start_at":1735689600,"strategy":"by_quota"} | 当前权益的使用量、总量、生效时间及资源使用策略。 |
### CommonCounter
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| used | Double | 100 | 当 `strategy` 为 `ByQuota` 时，表示当前权益的已使用量。若权益无相关用量数据则返回 `0`。 |
| total | Double | 1000 | * 当 `strategy` 为 `ByQuota` 时，表示当前权益的用量上限。 <br> * 当 `strategy` 为 `unlimit` 或 `forbidden`  时，该值为 1。 |
| end_at | Long | 1735689600 | 权益的结束生效时间，以 Unix 时间戳格式表示，单位为秒。 |
| start_at | Long | 1735689600 | 权益的开始生效时间，以 Unix 时间戳格式表示，单位为秒。 |
| strategy | String | by_quota | 资源使用策略，表示当前权益的资源使用方式。枚举值： <br>  <br> * `unlimit`：不限制用量。 <br> * `forbidden`：不支持调用，该套餐不支持该权益。 <br> * `by_quota`：按额度限制使用。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例

<div type="doc-tabs">
<div type="tab-item" title="查询套餐类型" key="mgoI9vMCeCD2uoRN1nd-4">

```JSON
curl --location --request GET 'https://api.coze.cn/v1/commerce/benefit/benefits/get' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
```


</div>
<div type="tab-item" title="查询插件 API QPS限制" key="CX4A_TXrVP2_9VFlkXjbh">

```JSON
curl --location --request GET 'https://api.coze.cn/v1/commerce/benefit/benefits/get?benefit_type_list=api_run_qps&resource_id=plugin' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
```


</div>
<div type="tab-item" title="查询多个权益类型" key="4C8OlTz24mhgeMg4lZspQ">

```JSON
curl --location --request GET 'https://api.coze.cn/v1/commerce/benefit/benefits/get?benefit_type_list=api_run_qps,call_tool_limit&resource_id=plugin' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
```


</div>
</div>
### 返回示例

<div type="doc-tabs">
<div type="tab-item" title="查询套餐类型" key="x6p1NxHPKbC3xbuugZhwh">

```JSON
{
    "msg": "",
    "data": {
        "basic_info": {
            "user_level": "enterprise"
        }
    },
    "detail": {
        "logid": "20251023180624A0C1***"
    },
    "code": 0
}
```


</div>
<div type="tab-item" title="查询插件 API QPS限制" key="7RdwGuq1wd6mA1dgmkHr-">

```JSON
{
    "code": 0,
    "msg": "",
    "data": {
        "basic_info": {
            "user_level": "enterprise"
        },
        "benefit_info": [
            {
                "resource_id": "plugin",
                "benefit_type": "api_run_qps",
                "basic": {
                    "item_info": {
                        "used": 0,
                        "total": 100,
                        "start_at": 0,
                        "end_at": 0,
                        "strategy": "by_quota"
                    },
                    "status": ""
                },
                "extra": [
                       {
                        "item_info": {
                            "start_at": 1758361772,
                            "end_at": 1758448172,
                            "strategy": "by_quota",
                            "used": 0,
                            "total": 10
                        },
                        "status": "expired"
                    }
                ],
                "effective": {
                    "item_info": {
                        "used": 0,
                        "total": 130,
                        "start_at": 0,
                        "end_at": 0,
                        "strategy": "by_quota"
                    },
                    "status": "valid"
                }
            }
        ]
    },
    "detail": {
        "logid": "202510231845400B8770A4***"
    }
}
```


</div>
</div>
## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
