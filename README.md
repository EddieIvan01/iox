# iox

English | [中文](https://github.com/EddieIvan01/iox/tree/master/docs/README_CN.md)

Tool for port forward & intranet proxy

# Features

+ Traffic encryption and compress (optional)
+ Humanized CLI option
+ Multiple protocol supporting (TCP/UDP/KCP)
+ Stream multiplexing

# Build

```
go build -ldflags='-s -w' -mod=mod
```

With KCP protocol support

```
go build -ldflags='-s -w' -mod=mod -tags kcp
```

# Usage

```
iox <MODE> [OPTIONS] <SOCKET_DESCRIPTOR> [SOCKET_DESCRIPTOR]
```

Forward TCP traffic between `:9999` (listen TCP) and `127.0.0.1:8888` (listen TCP)

```
iox fwd tcp-l:9999 tcp-l:127.0.0.1:8888
```

Forward traffic between `:9999` (listen TCP) and `1.1.1.1:9999` (connect KCP)

```
iox fwd tcp-l::9999 kcp:1.1.1.1:9999
```

Forward traffic between `1.1.1.1:9999` (connect TCP) and `1.1.1.1:8888` (connect KCP)

```
iox fwd tcp:1.1.1.1:9999 kcp:1.1.1.1:8888
```

Socks5 proxy over TCP on `:1080`

```
iox proxy proxy:1080
```

Reverse proxy

```
iox proxy proxy:1080 tcp-l:8888
iox proxy tcp:1.1.1.1:8888
```

Reverse proxy over KCP

```
iox proxy proxy:1080 kcp-l:8888
iox proxy kcp:1.1.1.1:8888
```

Socks5 proxy over KCP, enable three transmission features (need an local agent)

```
iox fwd tcp-l:1080 cxs@kcp:127.0.0.1:9999 -k 000102
iox proxy xcs@kproxy:9999 -k 000102
```

Multi forwarder chain

```
iox fwd tcp-l:1111 sxc@tcp:127.0.0.1:2222 -k 000102
iox fwd sxc@tcp-l:2222 sxc@kcp:127.0.0.1:3333 -k 000102
iox fwd sxc@kcp-l:3333 sxc@kcp:127.0.0.1:4444 -k 000102
iox proxy sxc@kproxy:4444 -k 000102
```

Advanced reverse proxy over KCP

```
iox proxy sxc@kproxy:8888 sxc@kcp-l:9999 -k 000102
iox proxy sxc@kcp:1.1.1.1:9999 -k 000102
iox fwd tcp-l:1080 sxc@kcp:1.1.1.1:8888 -k 000102
```

## MODE

There are two working mode: 

+ fwd: traffic forwarding
+ proxy: Socks5 proxy (forward or reverse)


## Socket Descriptor

Socket descriptor is in the following format:

```
[tags@]protocol[-l]:address
```

### E.g.

+ sxc@tcp-l:9999              // Listen TCP on `:9999`, enable three transmission features
+ tcp-l::9999                 // Listen TCP on `:9999`
+ tcp4-l:9999                 // Listen TCP4 on `:9999`
+ udp6:127.0.0.1:9999         // Connect UDP6 `127.0.0.1:9999`
+ kcp-l:127.0.0.1:9999        // Listen KCP on `127.0.0.1:9999`
+ xs@proxy:1080               // Socks5 proxy is listening on `:1080`, with encryption and multiplexing
+ c@kproxy:1080               // Socks5 proxy over KCP is listening on `:1080`, with compress 

### Tag

The tags specify the transmission features of the socket, which could be combined freely:

+ s: encrypted
+ x: multiplexing I/O
+ c: compress

### Protocol

+ tcp / tcp4 / tcp6
+ udp / udp4 / udp6
+ kcp
+ proxy (Socks5)
+ kproxy (Socks5 over KCP)


# License

The MIT license

