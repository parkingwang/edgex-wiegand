# EdgeX-Wiegand 微耕品牌设备驱动

包括两个微耕品牌驱动程序：

1. `endpoint`：接受AT指令控制，执行后返回响应结果的输出终端驱动；
1. `trigger`：接受微耕硬件UDP通讯数据的触发器，可以将微耕的刷卡事件返回到系统内部；

## 其它品牌

可以支持微耕OEM/ODM的其它品牌门禁设备。已知如下：

1. 东控智能

## Endpoint - 控制终端

接受AT控制指令的输出操作终端驱动。

> endpoint-wiegand

**程序参考配置** `/etc/edgex/application.toml`

```toml
NodeName = "WiegandEndpoint"
RpcAddress = "0.0.0.0:5570"
Broadcast = false

# 微耕主板配置参数
[BoardOptions]
  serialNumber = 223177933
  doorCount = 2

# Socket客户端配置参数
[SocketClientOptions]
  remoteAddress = "192.168.1.50:60000"
  readTimeout = "1s"
  writeTimeout = "1s"
```

配置说明：

- `NodeName` 节点名称，须保证项目内部唯一性；
- `RpcAddress` gRPC绑定地址；通过gRPC控制设备时的通讯地址；
- `Broadcast` Endpoint设备设置为广播模式时，在接收控制指令并处理后，将不读取设备的响应结果，直接返回成功。
- `BoardOptions.serialNumber` 微耕控制器的序列号。
- `BoardOptions.doorCount` 微耕控制器的控制门数量。
- `SocketClientOptions.remoteAddress` 微耕控制器的UDP通讯地址及端口。


----

### AT指令列表

#### AT+OPEN - 远程开关

远程开启指定编号的开关。

格式：

> AT+OPEN={SWITCH_ID} 

- `SWITCH_ID` 开关编号，有效为\[1-4\] 
    
#### AT+DELAY - 设置开关延时

格式：

> AT+DELAY={SWITCH_ID},{DELAY_SEC}
 
设置开关的延时时间。

- `SWITCH_ID` 开关编号，有效为\[1-4\]
- `DELAY_SEC` 延时时间，单位：秒

#### AT+ADD - 添加门禁卡

将卡号添加到门禁控制器内部系统，开启对应门号的授权。

格式：

> AT+ADD={CARD},{START_DATE},{END_DATE},{DOOR1},{DOOR2},{DOOR3},{DOOR4}

- `CARD` 卡号，卡号原始10位格式；
- `START_DATE` 有效期开始日期，格式为 YYYYMMdd，如: 20190521
- `END_DATE` 有效期结束日期，格式为 YYYYMMdd，如: 20190521
- `DOOR1 - DOOR4` 门号1-4，设置为1表示有权限，设置为0表示无权限；

#### AT+DELETE - 删除门禁卡

删除门禁控制器内部系统的卡号授权。

格式：

> AT+DELETE={CARD}

- `CARD` 卡号，卡号原始10位格式；

#### AT+CLEAR - 清空授权

添空门禁控制器内部系统的授权卡。

### AT指令响应

成功：

> EX=OK

错误：

> EX=ERR:{MESSAGE}

- `MESSAGE` 出错消息

----

## Trigger - 事件触发器

**程序配置**

```toml
# 顶级必要的配置参数
NodeName = "WiegandTrigger"
Topic = "wiegand/events"

# SocketServer配置参数
[SocketServerOptions]
  address = [
      "udp://0.0.0.0:5570"
  ]
```


配置说明：

- `NodeName` 设备名称，在项目内部每个设备名称必须保持唯一性；
- `Topic` 每个Trigger都必须指定一个Topic；不得以`/`开头；
- `SocketServerOptions.address` 服务端监听地址列表；可监听多个地址；

#### 程序说明

Trigger启动后，等待微耕控制器连接到程序的UDP服务端，并接收其刷卡广播数据。
接收到刷卡数据后，将数据生成以下JSON格式数据包，以指定的Topic发送到MQTT服务器。

消息Name格式：

> TRIGGER-{serialNumber}-{doorId]-{direct}

消息数据格式：

```json
{
  "sn": 123,
  "card": "0005653307",
  "index": 123,
  "type": 1,
  "typeName": "CARD",
  "doorId": 1,
  "direct": "IN",
  "state": 1
}
```

- `sn` 设备序列号；
- `card` 卡号，卡号原始10位数字。
- `doorId` 刷卡门号；
- `direct` 进出方向；
- `state` 刷卡状态；

## 参考资料

1. [短报文格式_操作实例](WG-proto-operator.pdf)
1. [短报文格式](WG-proto.pdf)


