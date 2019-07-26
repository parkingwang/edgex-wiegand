# EdgeX-Wiegand 微耕品牌设备驱动

可以支持微耕OEM/ODM的其它品牌门禁设备。已知如下：

1. 东控智能

----

## 配置参考

以下为完整的配置参考文件：

```toml
NodeName = "WiegandNode"
RpcAddress = "0.0.0.0:5577"
Topic = "wiegand/events"

# 主板配置参数
[BoardOptions]
  serialNumber = 123456
  doorCount = 2

# UDP服务端配置
[UdpServerOptions]
  listenAddress = "0.0.0.0:6670"

# UDP客户端配置参数
[UdpClientOptions]
  remoteAddress = "192.168.1.12:60000"

```

| 参数 | 格式 | 必填 | 说明 |
|-----|-----|-----|-----|
| NodeName | string | 是 | 节点名称。NodeName必须与其它节点不同。 |
| RpcAddress | string | 是 | gRPC绑定通讯地址 |
| Topic | string | 是 | 节点接收到数据后，发送的MQTT主题名称。 |
| BoardOptions.serialNumber | int | 是 | 微耕驱动的主板序列号 |
| BoardOptions.doorCount | int | 是 | 主板支持的门锁数量 |
| UdpServerOptions.listenAddress | string | 是 | UDP监听地址及端口，微耕主板连接到此端口来上报数据 |
| UdpClientOptions.remoteAddress | string | 是 | UDP控制端地址及端口，节点向此地址发送控制指令 |

**全局配置说明**

全局参数配置，以及MQTT相关选项的参数配置，
参见：[EdgeX-Go](https://github.com/nextabc-lab/edgex-go) 相关文件说明。


#### 配置文件位置

配置文件默认文件名为`application.toml`，可以放在以下位置，并按以下顺序搜索：

1. 运行目录下；
2. 目录：`/etc/edgex/`下；
3. 环境变量`EDGEX_CONFIG`指定的完整路径；

**修改配置文件名称**

通过运行时指定参数 `-c` 来修改默认文件名。注意，目录搜索依然按以上描述顺序。例如：

```bash
$ ./edgenode-wierand -c wiegand.toml
```

----

### AT指令列表

通过gRPC服务，可以向节点发送控制指令。

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

## 微耕门禁主板通讯协议

1. [短报文格式_操作实例](WG-proto-operator.pdf)
1. [短报文格式](WG-proto.pdf)


