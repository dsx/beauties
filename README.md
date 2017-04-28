## Description

«Koinonia is a transliterated form of the Greek word, κοινωνία, which
means communion, joint participation; the share which one has in
anything, participation, a gift jointly contributed, a collection, a
contribution, etc.» [Wikipedia](https://en.wikipedia.org/wiki/Koinonia)

Beauties is just a random name picked by @goberghen at 1493325888.

Together it means beauty of the Internet, which allows us to share
information.

## Help

Koinonia server
[index page](https://github.com/dsx/beauties/cmd/koinonia/static/index.txt).

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
