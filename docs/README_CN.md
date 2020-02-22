# iox

[English](https://github.com/EddieIvan01/iox) | 中文

端口转发 & 内网代理工具，功能类似于`lcx`/`ew`，但是比它们更好

## 为什么写iox?

`lcx`和`ew`是很优秀的工具，但还可以提高

在我最初使用它们的很长一段时间里，我都记不住那些复杂的命令行参数，诸如`tran, slave, rcsocks, sssocks`。工具的工作模式很清晰，明明可以用简单的参数表示，为什么他们要设计成这样（特别是`ew`的`-l -d -e -f -g -h`）

除此之外，我认为网络编程的逻辑可以优化

举个栗子，当运行`lcx -listen 8888 9999`命令时，客户端必须先连`:8888`，再连`:9999`，实际上这两个端口是平等的，在`iox`里则没有这个限制。当运行`lcx -slave 1.1.1.1 8888 1.1.1.1 9999`命令时，`lcx`会串行的连接两个主机，但是并发连接两个主机会更高效，毕竟是纯I/O操作，`iox`就是这样做的

更进一步，`iox`提供了流量加密功能。实际上，你可以直接将`iox`当做一个简易的ShadowSocks使用

当然，因为`iox`是用Go写的，所以静态连接的程序有一点大，原程序有2.2MB（UPX压缩后800KB）

## 特性

+ 流量加密（可选）
+ 友好的命令行参数
+ 部分逻辑优化
+ UDP流量转发（TODO）

## 用法

#### 两种模式

所有参数都是统一的。`-l/--local`意为监听本地端口；`-r/--remote`意为连接远端主机

**fwd**

Local2Local

```
./iox fwd -l 8888 -l 9999


for lcx:
./lcx -listen 8888 9999
```

Local2Remote

```
./iox fwd -l 8888 -r 1.1.1.1:9999


for lcx:
./lcx -tran 8888 1.1.1.1 9999
```

Remote2Remote

```
./iox fwd -r 1.1.1.1:8888 -r 1.1.1.1:9999


for lcx:
./lcx -slave 1.1.1.1 8888 1.1.1.1 9999
```

**proxy**

LocalProxy

```
./iox proxy -l 1080


for ew:
./ew -s ssocksd -l 1080
```

RemoteProxy (command pair)

```
./iox proxy -r 1.1.1.1:9999
./iox proxy -l 9999 -l 1080       // 注意，这两个端口是有顺序的


for ew:
./ew -s rcsocks -l 1080 -e 9999
./ew -s rssocks -d 1.1.1.1 -e 9999
```

***

#### 启用加密

举个栗子，我们把内网3389端口转发到VPS

```
./iox fwd -r 192.168.0.100:3389 -r *1.1.1.1:8888 -k 656565

./iox fwd -l *8888 -l 1080 -k 656565
```

很好理解：被控主机和VPS:8888之间的流量会被加密，预共享的密钥是'AAA'，`iox`会用这个密钥生成种子密钥和IV，并用AES-CTR流加密

所以，`*`应该成对使用

```
./iox fwd -l 1000 -r *127.0.0.1:1001 -k 000102
./iox fwd -l *1001 -r *127.0.0.1:1002 -k 000102
./iox fwd -l *1002 -r *127.0.0.1:1003 -k 000102
./iox proxy -l *1003
```

你也可以把`iox`当做一个简单的ShadowSocks来用：

```
./iox proxy -l *9999


./iox fwd -l 1080 -r *VPS:9999
```

## 许可

The MIT license

