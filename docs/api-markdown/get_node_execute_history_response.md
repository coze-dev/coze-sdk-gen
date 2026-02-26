# 查询输出节点的执行结果
查询输出节点的执行结果。
## 接口描述
通过 [查询工作流异步执行结果](https://docs.coze.cn/developer_guides/workflow_history) API 查询工作流执行结果时，如果工作流输出节点的输出内容超过 1MB，查询工作流异步执行结果 API 无法返回完整的输出节点内容。需要调用本 API，根据工作流的执行 ID 以及[查询工作流异步执行结果](https://docs.coze.cn/developer_guides/workflow_history) API 返回的节点执行 UUID，逐一查询每个节点的输出内容。
## 接口限制

* 本 API 的流控限制请参见 [API 介绍](https://docs.coze.cn/developer_guides/coze_api_overview)。
* 输出节点的输出数据最多保存 24 小时。
* 仅支持查询输出节点的执行结果，不支持查询结束节点的执行结果。
* 输出节点的输出内容超过1MB 时，无法保证返回内容的完整性。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflows/:workflow_id/run_histories/:execute_id/execute_nodes/:node_execute_uuid <br> ``` <br>  |
| **权限** | `listRunHistory` <br> 确保调用该接口使用的访问令牌开通了 `listRunHistory` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询输出节点的执行结果。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String | 必选 | 73505836754923*** | 要执行的 Workflow ID，需要先完成发布 Workflow 的操作。 <br> 进入 Workflow 编排页，在页面 URL 中，workflow 参数后的数字就是 Workflow ID。例如： <br> https://www.coze.com/work_flow?space_id=73119690542463***&workflow_id=73505836754923*** , Workflow ID 为 73505836754923*** 。 |
| execute_id | String | 必选 | 743104097880585**** | 工作流的执行 ID。调用接口[执行工作流](https://docs.coze.cn/developer_guides/workflow_run)，如果选择异步执行工作流，响应信息中会返回 execute_id。 |
| node_execute_uuid | String | 必选 | 78923456777***** | [工作流异步执行结果](https://docs.coze.cn/developer_guides/workflow_history) API 中返回的节点执行 uuid。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [WorkflowNodeExecuteHistory](#workflownodeexecutehistory) | \ | 节点的执行结果。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "202410242028595CCF353CEC86A8*****" } | 该请求的响应信息。 |
### WorkflowNodeExecuteHistory
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| is_finish | Boolean | true | 节点是否执行完成。 <br>  <br> * `true` 表示执行已完成。 <br> * `false`表示执行未完成。 |
| node_output | String | 0《零的奇妙冒险》\n\n在一个神秘而遥远的数字王国里，有一个特殊的数字——零。 | 节点的执行结果输出信息。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/workflows/74730423****702316/run_histories/75142305***8428/execute_nodes/63c7f****6d2' \
--header 'Authorization: Bearer pat_uk74546778ffgd****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "is_finish": false,
    "node_output": "0《零的奇妙冒险》\n\n在一个神秘而遥远的数字王国里，有一个特殊的数字——零。零一直觉得自己很渺小，没有什么特别之处，常常感到很孤独。\n\n有一天，零百无聊赖地在数字王国的小路上游荡。突然，它听到了一阵哭声。零好奇地顺着哭声找去，发现是数字一在角落里哭泣。\n\n“你怎么了，一？”零关心地问。\n\n一抽噎着说：“大家都觉得我很孤单，很弱小，没有什么用处。我也不知道自己能做什么。”\n\n零听了，心中涌起一股同情。它笑着对一说：“别难过，一。其实我们可以一起变得很强大。你看，当我和你站在一起的时候，我们就变成了十。”\n\n一惊讶地看着零，眼中渐渐燃起了希望。从那以后，零和一成为了好朋友，它们一起在数字王国里探索。\n\n不久，它们遇到了数字二。二正为自己不能变成更大的数字而苦恼。零和一相视一笑，走过去对二说：“别担心，我们可以一起组成十二。”\n\n数字二高兴极了，加入了它们的队伍。随着它们的冒险继续，越来越多的数字加入了零的小团队。它们发现，只要零和其他数字在一起，就能创造出无数的可能。\n\n有一次，数字王国遭遇了一场巨大的危机。一个邪恶的数字巫师出现了，他想要把所有的数字都变成他的奴隶，让数字王国陷入黑暗。\n\n零和它的朋友们决定挺身而出，保卫数字王国。它们齐心协力，运用各自的力量，与数字巫师展开了一场激烈的战斗。\n\n零虽然本身没有很大的力量，但它却能让其他数字发挥出更大的作用。在关键时刻，零带领着大家组成了一个巨大的数字组合，成功地击败了数字巫师。\n\n数字王国恢复了往日的和平与欢乐。零也不再觉得自己渺小和孤独了。它明白了，每个数字都有自己的价值，而当大家团结在一起的时候，就能创造出无穷的奇迹。\n\n从那以后，零成为了数字王国里最受欢迎的数字之一。它和它的朋友们继续在数字王国里冒险，用它们的团结和智慧，为数字王国带来更多的美好。\n\n在这个充满奇幻的数字王国里，零的故事告诉我们，即使是最渺小的存在，也能在合适的时候发挥出巨大的作用。只要我们相信自己，勇敢地去探索，去团结他人，就能创造出属于我们自己的精彩世界。而零，也将继续带着它的勇气和善良，在数字王国里书写更多的传奇故事。"
  },
  "detail": {
    "logid": "202506112018****AB2ADDA"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

