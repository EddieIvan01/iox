# iox

English | [中文](https://github.com/EddieIvan01/iox/tree/master/docs/README_CN.md)

Tool for port forward & intranet proxy, just like `lcx`/`ew`, but better

## Why write?

`lcx` and `ew` are awesome, but can be improved.

when I first used them, I can't remember these complicated parameters for a long time, such as `tran, slave, rcsocks, sssocks...`. The work mode is clear, why do they design parameters like this(especially `ew`'s `-l -d -e -f -g -h`)

Besides, I think the net programming logic could be optimized. 

For example, while running `lcx -listen 8888 9999` command, client must connect to `:8888` first, then `:9999`, in `iox`, there's no limit to the order in two ports. And while running `lcx -slave 1.1.1.1 8888 1.1.1.1 9999` command, `lcx` will connect two hosts serially, but it's more efficient to connect in concurrently, as `iox` does.

And what's more, `iox` provides traffic encryption feature. Actually, you can use `iox` as a simple ShadowSocks.

Of course, because `iox` is written in Go, the static-link-program is a little big, raw program is 2.2MB (800KB for UPX compression)

## Feature

+ traffic encryption (optional)
+ humanized CLI option
+ logic optimization
+ UDP traffic forward (TODO)

## Usage

#### Two mode

You can see, all params are uniform. `-l/--local` means listen on a local port; `-r/--remote` means connect to remote host

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
./iox proxy -l 9999 -l 1080       // notice, the two port are in order


for ew:
./ew -s rcsocks -l 1080 -e 9999
./ew -s rssocks -d 1.1.1.1 -e 9999
```

***

#### enable encryption

For example, we forward 3389 port in intranet to our VPS

```
./iox fwd -r 192.168.0.100:3389 -r *1.1.1.1:8888 -k 656565

./iox fwd -l *8888 -l 1080 -k 656565
```

It's easy to understand: traffic between be-controlled host and our VPS:8888 will be encrypted, the pre-shared secret key is 'AAA', `iox` will use it to generate seed key and IV, then encrypt with AES-CTR

So, the `*` should be used in pairs

```
./iox fwd -l 1000 -r *127.0.0.1:1001 -k 000102
./iox fwd -l *1001 -r *127.0.0.1:1002 -k 000102
./iox fwd -l *1002 -r *127.0.0.1:1003 -k 000102
./iox proxy -l *1003
```

Using `iox` as a simple ShadowSocks

```
// ssserver
./iox proxy -l *9999


// sslocal
./iox fwd -l 1080 -r *VPS:9999
```

## License

The MIT license

