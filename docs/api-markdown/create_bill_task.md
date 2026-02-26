# 导出终端用户账单
导出终端用户的账单，用于查看终端用户的用量明细，包括：语音识别 ASR 的音频时长、语音合成 TTS 的字符数、语音合成 TTS 的对话次数、RTC 通话时长、金额等信息。
* 仅扣子企业旗舰版的超级管理员和管理员可以调用该 API。
* 调用此 API 之前，需要确保企业下的设备已成功上报了设备信息，设备信息的配置方法可参考[设置设备信息](https://docs.coze.cn/dev_how_to_guides/deviceInfo)。

## 接口描述

* 调用此 API 后，后台将创建一个导出任务并返回任务 ID 和状态，你可以通过[查询账单文件](https://docs.coze.cn/developer_guides/billing_file)获取账单下载链接。
* 仅支持按天查询，例如查询 2025 年 3 月 27 日 00:00:00 至 23:59:59 的账单。查询周期默认为当天的 0 时至 24 时。
* 该 API 为异步接口，若账单数据量较大，可能需等待约 1 分钟才能获取账单 URL。
* 下载链接的有效期为 7 天，过期后需要重新导出账单。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/commerce/benefit/bill_tasks <br> ``` <br>  |
| **权限** | `createBillDownloadTask` <br> 确保调用该接口使用的个人令牌开通了 `createBillDownloadTask` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 导出终端用户的账单。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。可在扣子编程生成，具体生成步骤及相关信息，详见准备工作文档。 |
| Content-Type | application/json | 用于指定解析请求正文的格式，表明请求体为 JSON 格式数据。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| started_at | Long | 可选 | 1743004800 | 待导出账单的开始时间，格式为 Unix 时间戳，单位为秒。 <br> 仅支持设置为某天 0 点整的时间戳，例如，2025-03-27 00:00:00。如果填写的时间戳非 00:00:00，系统将自动调整为当天 00:00:00 的时间戳。 <br> * 暂不支持导出当日账单，因此起始时间不能设置为当天的时间。 <br> * 只能导出 2025-03-13 及之后的数据。 <br>  |
| ended_at | Long | 可选 | 1743091199 | 待导出账单的截止时间，格式为 Unix 时间戳，单位为秒。 <br> 仅支持设置为某天 23:59:59 的时间戳，例如，2025-03-27 23:59:59。如果填写的时间戳非 23:59:59，系统将自动调整为当天 23:59:59 的时间戳。 <br> 暂不支持导出当日账单，因此起始时间不能设置为当天的时间。 <br>  |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [BillTaskInfo](#billtaskinfo) | \ | 接口调用成功时返回的任务详细信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### BillTaskInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| status | String | running | 任务当前状态： <br>  <br> * init：任务初始化中。 <br> * running：账单数据生成中。 <br> * succeed：账单已生成，可下载导出文件。 <br> * failed：任务执行失败，可结合 msg 字段查看失败原因。 |
| task_id | String | 123 | 账单导出任务的唯一标识 ID，用于后续查询任务状态或获取导出文件。 |
| ended_at | Long | 1743091200 | 账单上的截止时间。 |
| file_urls | Array of String | ["https://example.com/bill_1.csv", "https://example.com/bill_2.csv"] | 账单导出文件的链接列表，任务完成后，可通过这些链接下载账单文件。 <br> 如果账单中的数据量过大，扣子编程会将数据拆分为多个文件，每个文件最多包含 50 万行数据。 <br> 账单文件中的字段说明请参见[账单推送回调事件](https://docs.coze.cn/developer_guides/billing_callback_message)。 <br>  |
| created_at | Long | 1743213605 | 任务的创建时间，Unix 时间戳，单位为秒。 |
| expires_at | Long | 1743177600 | 任务的过期时间，默认账单将保留 7 天，超过此时间后，无法再下载账单文件。Unix 时间戳，单位为秒。 |
| started_at | Long | 1743004800 | 账单上的起始时间。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/commerce/benefit/bill_tasks' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
{
  "started_at": 1743004800,
  "ended_at": 1743091200
}
```

### 返回示例
```JSON
{ 
    "data": {
        "task_id": 123***,
        "started_at": 1743004800,
        "ended_at": 1743091200,
        "created_at": 1743213605,
        "expires_at": 1743177600,
        "file_urls": "",
        "status": "running"
    },
    "code": 0, 
    "msg": "" 
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

