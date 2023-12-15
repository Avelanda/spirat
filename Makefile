VERSION ?= $(or $(git rev-list -1 HEAD),unknown)
BUILD ?= $(or $(git rev-list -1 HEAD),unknown)

COMPILE_FLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.build=$(BUILD)" -trimpath

.PHONY: build
build: build-linux build-win

.PHONY: build-linux
build-linux:
	go build -o build/spirat $(COMPILE_FLAGS) .

.PHONY: build-win
build-win:
	GOOS=windows go build -o build/spirat.exe $(COMPILE_FLAGS) .

.PHONY: docker
docker:
	docker build -t spirat .

.PHONY: e2e
e2e:
	test/all.sh
