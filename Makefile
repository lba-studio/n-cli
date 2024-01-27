tag_candidate = ${shell go run . version}

config:
	code $(shell go run . where config)

tag:
	@echo "Using tag: ${tag_candidate}"
	git tag ${tag_candidate}
	git push origin refs/tags/${tag_candidate}

test_version:
	VERSION_TEST=true go test ./pkg/version/...

build:
	go build -v .