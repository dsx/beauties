.PHONY: all

all: build

help:
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo " all (default)	- build everything"
	@echo "	install		- install koinonia binary to ${GOPATH}/bin"
	@echo "	clean		- make package directory (${GOPATH}/src/github.com/dsx/beauties) as clean as possible"
	@echo " veryclean	- clean and remove koinonia binary"
	@echo "	deb		- build default debian package"
	@echo "	deps		- install dependencies for development"

clean:
	@rm -rf \
	"${GOPATH}/src/github.com/dsx/beauties/debian/debhelper-build-stamp" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/files" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/koinonia" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/koinonia.debhelper.log" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/koinonia.postinst.debhelper" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/koinonia.postrm.debhelper" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/koinonia.prerm.debhelper" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/koinonia.substvars" \
	"${GOPATH}/src/github.com/dsx/beauties/obj-x86_64-linux-gnu" \
	"${GOPATH}/src/github.com/dsx/beauties/pkg"

veryclean: clean
	@rm -f \
	"${GOPATH}/bin/koinonia" \
	"${GOPATH}/src/github.com/dsx/beauties/cmd/koinonia/bindata.go" \
	"${GOPATH}/src/github.com/dsx/beauties/cmd/koinonia/koinonia" \
	"${GOPATH}/src/github.com/dsx/beauties/koinonia"

build:
	go get github.com/dsx/beauties/cmd/koinonia
	go generate github.com/dsx/beauties/cmd/koinonia
	go build github.com/dsx/beauties/cmd/koinonia

install:
	go install github.com/dsx/beauties/cmd/koinonia

deb: build
	@echo debuild -e GOPATH -b

deps:
	gvt update -all
	go get github.com/rogpeppe/godef github.com/FiloSottile/gvt
	go install github.com/rogpeppe/godef github.com/FiloSottile/gvt
