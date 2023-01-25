.PHONY: all
all: goreleaser.yaml

goreleaser.yaml: goreleaser/goreleaser.yaml.j2 goreleaser/goreleaser-matrix.yaml
	jinja2 goreleaser/goreleaser.yaml.j2 goreleaser/goreleaser-matrix.yaml > goreleaser.yaml

.PHONY: clean
clean:
	rm goreleaser.yaml