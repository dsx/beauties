## Description

Beauties is just a random name picked by @goberghen at 1493325888.

It refers to a beauty of the Internet, which allows us to share
information.

## Help

Beauties server
[index page](https://raw.githubusercontent.com/dsx/beauties/master/cmd/beauties/static/index.txt)

```
make help
```

## Build

```
make
```

## Develop

```
make deps
```

## Debian package

For now `/usr/bin/dh_systemd_enable` must be patched with
`misc/dh_systemd_update.patch` to build package properly. There is a bug
[#841746](https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=841746)
with the original patch to address this problem.

```
make deb
```
