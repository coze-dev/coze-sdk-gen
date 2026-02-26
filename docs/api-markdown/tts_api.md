# 双向流式语音合成
扣子编程提供流式语音合成 WebSocket OpenAPI，可以将文字信息转为指定音色的语音片段。
双向流式语音合成场景下的各类事件详细信息可参考[双向流式语音合成事件](https://docs.coze.cn/api/open/docs/developer_guides/tts_event)。
## 接口信息
| **URL** | `wss://ws.coze.cn/v1/audio/speech` |
| --- | --- |
| **Headers** | `Authorization Bearer `*`$Access_Token`* <br> 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息请参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| **权限** | `createSpeech` |
| **接口说明** | 将文字信息转为指定音色的语音片段。 |
## 限制说明
单次请求的文字信息长度最大为 1024 个字节。超过上限时会提示 `3010` 错误。
建议一次性不要传输太多文字。
## 建连示例代码

<div type="doc-tabs">
<div type="tab-item" title="ws module(Node.js)" key="jToGp1GlY3">


```JavaScript
import WebSocket from 'ws';

const url = `wss://ws.coze.cn/v1/audio/speech?authorization=Bearer ${ACCESS_TOKEN}`;
const ws = new WebSocket(url);

ws.on('open', function open() {
  console.log('Connected to server.');
});

ws.on('message', function incoming(message) {
  console.log(JSON.parse(message.toString()));
});
```



</div>
<div type="tab-item" title="websocket-client(Python)" key="r8hawZngEG">

```Python
# example requires websocket-client library:
# pip install websocket-client

import os
import json
import websocket

ACCESS_TOKEN = os.environ.get("ACCESS_TOKEN")

url = "wss://ws.coze.cn/v1/audio/speech"

headers = [
    "Authorization: Bearer " + ACCESS_TOKEN
]

def on_open(ws):
    print("Connected to server.")

def on_message(ws, message):
    data = json.loads(message)
    print("Received event:", json.dumps(data, indent=2))

ws = websocket.WebSocketApp(
    url,
    header=headers,
    on_open=on_open,
    on_message=on_message,
)

ws.run_forever()
```


</div>
<div type="tab-item" title="Websocket(browsers)" key="F3UvaLYqYX">


```JavaScript
const url = `wss://ws.coze.cn/v1/audio/speech?authorization=Bearer ${ACCESS_TOKEN}`;
const ws = new WebSocket(url);

ws.addEventListener('open', function () {
  console.log('Connected to server.');
});

ws.addEventListener('message', function (message) {
  console.log(JSON.parse(message.data.toString()));
});
```



</div>
</div>
## API 时序图
交互流程如下：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/dc7c060fe4064cf0b7b60557ff9a99ba~tplv-goo7wpa0wc-image.image)

