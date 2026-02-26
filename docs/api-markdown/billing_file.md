# 查询账单文件
查询账单文件。
## 接口描述
当你[导出设备账单](https://docs.coze.cn/developer_guides/create_bill_task)后，可以通过本 API 查询账单文件，获取对应账单文件的 URL 链接，以便下载或查看已导出的账单数据。你可以通过账单文件查看智能硬件设备的用量明细，包括：语音识别 ASR 的音频时长、语音合成 TTS 的字符数、语音合成 TTS 的对话次数、RTC 通话时长、金额等信息。
* 仅扣子企业旗舰版的超级管理员和管理员可以调用该 API。
* 调用此 API 之前，需要确保企业下的设备已成功上报了设备信息，设备信息的配置方法可参考[设置设备信息](https://docs.coze.cn/dev_how_to_guides/deviceInfo)。

## 接口限制

* 账单下载链接的有效期为 7 天，过期后需要调用[导出设备账单](https://docs.coze.cn/developer_guides/create_bill_task) API 重新导出账单。
* 只有当任务状态为 `succeed` 时，才会返回账单的下载链接。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/commerce/benefit/bill_tasks <br> ``` <br>  |
| **权限** | `listBillDownloadTask` <br> 确保调用该接口使用的访问令牌开通了 `listBillDownloadTask` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询账单文件。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。可在扣子编程生成，具体生成步骤及相关信息，详见准备工作文档。 |
| Content-Type | application/json | 用于指定解析请求正文的格式，表明请求体为 JSON 格式数据。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| task_ids | Array of String | 可选 | [123, 456] | 账单导出任务 ID 列表，最多填写 100 个任务 ID。 <br> 若不填写该参数，默认查询最近 7 天内创建的账单导出任务。 |
| page_num | Integer | 可选 | 1 | 用于设置查询结果分页展示时的页码，最小值为 1，默认值为 1。 |
| page_size | Integer | 可选 | 20 | 用于设置查询结果分页展示时每页返回的数据量，取值范围为 1 ~ 200，默认值为 20。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [BillBusinessData](#billbusinessdata) | - | 接口调用成功时返回的任务详细信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### BillBusinessData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| total | Integer | 12 | 当前查询条件下的总账单任务数。 |
| task_infos | Array of [BillTaskInfo](#billtaskinfo) | - | 包含多个账单导出任务信息的列表，每个任务信息包含任务 ID、时间范围、文件 URL 等详细内容。 |
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
curl --location --request GET 'https://api.coze.cn/v1/commerce/benefit/bill_tasks?page_num=1&page_size=50' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{ 
    "data": {
        "total": 1,
        "task_infos": [
            {
                "task_id": 123***,
                "started_at": 1743004800,
                "ended_at": 1743091200,
                "created_at": 1743213605,
                "expires_at": 1743177600,
                "file_urls": ["https://example.com/bill_1.csv", "https://example.com/bill_2.csv"],
                "status": "succeed"
            }
        ]
    },
    "code": 0, 
    "msg": "" 
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
