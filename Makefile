DESTDIR ?= obj
DIRTY := $(shell git diff --stat)

.PHONY: all
all: build

.PHONY: build
build: goreleaser-deb.yaml
	goreleaser build --config goreleaser-deb.yaml --rm-dist --snapshot

goreleaser.yaml: goreleaser/goreleaser.yaml.j2 goreleaser/full.yaml
	jinja2 goreleaser/goreleaser.yaml.j2 goreleaser/full.yaml > goreleaser.yaml

goreleaser-small.yaml: goreleaser/goreleaser.yaml.j2 goreleaser/small.yaml
	jinja2 goreleaser/goreleaser.yaml.j2 goreleaser/small.yaml > goreleaser-small.yaml

goreleaser-deb.yaml: goreleaser/goreleaser.yaml.j2 goreleaser/small.yaml
	jinja2 goreleaser/goreleaser.yaml.j2 goreleaser/deb.yaml > goreleaser-deb.yaml

.PHONY: clean
clean:
	rm -f goreleaser.yaml
	rm -f goreleaser-small.yaml

.PHONY: patch
patch:
	[ "${DIRTY}" = "" ]
	./scripts/bump.sh patch



install:
	mkdir -p ${DESTDIR}/usr/bin/
	cp dist/main-linux-amd64-glibc_linux_amd64_v1/acmednsproxy ${DESTDIR}/usr/bin/
	cp dist/tool-linux-amd64-glibc_linux_amd64_v1/adpcrypt ${DESTDIR}/usr/bin/
	#mkdir -p ${DESTDIR}/etc/default
	#cp ./distros/deb/acmednsproxy.default ${DESTDIR}/etc/default/acmednsproxy
	#mkdir -p ${DESTDIR}/usr/share/doc/hello
	#cp -r examples ${DESTDIR}/usr/share/doc/hello/

.PHONY: deb
deb:
	dpkg-buildpackage -b
