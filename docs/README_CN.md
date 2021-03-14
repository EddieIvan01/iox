# iox

[English](https://github.com/EddieIvan01/iox) | 中文

端口转发 & 内网代理工具

# 特性

+ 流量加密和压缩 (可选)
+ 人性化的命令行参数
+ 多协议支持 (TCP/UDP/KCP)
+ 流多路复用

# 用法

```
iox <MODE> [OPTIONS] <SOCKET_DESCRIPTOR> [SOCKET_DESCRIPTOR]
```

在 `:9999` (listen TCP) 和 `127.0.0.1:8888` (listen TCP) 间流量转发

```
iox fwd tcp-l:9999 tcp-l:127.0.0.1:8888
```

在 `:9999` (listen TCP) 和 `1.1.1.1:9999` (KCP) 间流量转发

```
iox fwd tcp-l::9999 kcp:1.1.1.1:9999
```

在 `1.1.1.1:9999` (connect TCP) 和 `1.1.1.1:8888` (connect KCP)间流量转发

```
iox fwd tcp:1.1.1.1:9999 kcp:1.1.1.1:8888
```

监听在 `:1080` 的基于TCP的正向Socks5代理

```
iox proxy proxy:1080
```

反向代理

```
iox proxy proxy:1080 tcp-l:8888
iox proxy tcp:1.1.1.1:8888
```

基于KCP的反向代理

```
iox proxy proxy:1080 kcp-l:8888
iox proxy kcp:1.1.1.1:8888
```

基于KCP的正向代理，开启三种传输特性 (需要一个本地agent配合)

```
iox fwd tcp-l:1080 xsc@kcp:127.0.0.1:9999 -k 000102
iox proxy sxc@kproxy:9999 -k 000102
```

多级转发链

```
iox fwd tcp-l:1111 sxc@tcp:127.0.0.1:2222 -k 000102
iox fwd sxc@tcp-l:2222 sxc@kcp:127.0.0.1:3333 -k 000102
iox fwd sxc@kcp-l:3333 sxc@kcp:127.0.0.1:4444 -k 000102
iox proxy sxc@kproxy:4444 -k 000102
```

高级基于KCP的反向代理

```
iox proxy sxc@kproxy:8888 sxc@kcp-l:9999 -k 000102
iox proxy sxc@kcp:1.1.1.1:9999 -k 000102
iox fwd tcp-l:1080 sxc@kcp:1.1.1.1:8888 -k 000102
```


## MODE

两种工作模式: 

+ fwd: 流量转发
+ proxy: Socks5代理 (正反向)


## Socket Descriptor

套接字描述符为以下格式:

```
[tags@]protocol[-l]:address
```

### E.g.

+ sxc@tcp-l:9999              // 监听TCP `:9999`, 开启三种传输特性
+ tcp-l::9999                 // 监听TCP `:9999`
+ tcp4-l:9999                 // 监听TCP4 `:9999`
+ udp6:127.0.0.1:9999         // 连接UDP6 `127.0.0.1:9999`
+ kcp-l:127.0.0.1:9999        // 监听KCP `127.0.0.1:9999`
+ xs@proxy:1080               // 监听在 `:1080` 的Socks5代理, 开启加密和多路复用
+ c@kproxy:1080               // 监听在 `:1080` 的基于KCP的Socks5代理, 开启压缩

### Tag

标签用来指定套接字的传输特性，可以自由组合:

+ s: 加密
+ x: I/O多路复用
+ c: 压缩

### Protocol

+ tcp / tcp4 / tcp6
+ udp / udp4 / udp6
+ kcp
+ proxy: socks5 proxy (仅可在proxy模式中只用)
+ kproxy: socks5 proxy over KCP (仅可在proxy模式中只用)


# License

The MIT license

