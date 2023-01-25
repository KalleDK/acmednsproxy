DIRTY := $(shell git diff --stat)

.PHONY: all
all: goreleaser.yaml goreleaser-small.yaml

goreleaser.yaml: goreleaser/goreleaser.yaml.j2 goreleaser/full.yaml
	jinja2 goreleaser/goreleaser.yaml.j2 goreleaser/full.yaml > goreleaser.yaml

goreleaser-small.yaml: goreleaser/goreleaser.yaml.j2 goreleaser/full.yaml
	jinja2 goreleaser/goreleaser.yaml.j2 goreleaser/small.yaml > goreleaser-small.yaml


.PHONY: clean
clean:
	rm -f goreleaser.yaml
	rm -f goreleaser-small.yaml

.PHONY: bump
bump:
	[ "${DIRTY}" = "" ]
	echo h
	