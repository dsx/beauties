.PHONY: all

all: build

help:
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo " all (default)	- build everything"
	@echo "	install		- install beauties binary to ${GOPATH}/bin"
	@echo "	clean		- make package directory (${GOPATH}/src/github.com/dsx/beauties) as clean as possible"
	@echo " veryclean	- clean and remove beauties binary"
	@echo "	deb		- build default debian package"
	@echo "	deps		- install dependencies for development"

clean:
	@rm -rf \
	"${GOPATH}/src/github.com/dsx/beauties/debian/debhelper-build-stamp" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/files" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/beauties" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/beauties.debhelper.log" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/beauties.postinst.debhelper" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/beauties.postrm.debhelper" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/beauties.prerm.debhelper" \
	"${GOPATH}/src/github.com/dsx/beauties/debian/beauties.substvars" \
	"${GOPATH}/src/github.com/dsx/beauties/obj-x86_64-linux-gnu" \
	"${GOPATH}/src/github.com/dsx/beauties/pkg"

veryclean: clean
	@rm -f \
	"${GOPATH}/bin/beauties" \
	"${GOPATH}/src/github.com/dsx/beauties/cmd/beauties/bindata.go" \
	"${GOPATH}/src/github.com/dsx/beauties/cmd/beauties/beauties" \
	"${GOPATH}/src/github.com/dsx/beauties/beauties"

build:
	go generate github.com/dsx/beauties/cmd/...
	go build github.com/dsx/beauties/cmd/...

install:
	go install github.com/dsx/beauties/cmd/...

deb: veryclean build
	@strip "${GOPATH}/src/github.com/dsx/beauties/beauties"
	@misc/gen-deb-changelog-from-git.sh
	@debuild -b

deps:
	go get github.com/rogpeppe/godef github.com/FiloSottile/gvt github.com/jteeuwen/go-bindata/...
	go install github.com/rogpeppe/godef github.com/FiloSottile/gvt github.com/jteeuwen/go-bindata
