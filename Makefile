# This makefile defines the following targets
#
#   - baasBackend - build the baasBackend executable
#   - dockerBaasBackend - build baas backend docker image

VERSION=v0.1

VERSIONPKG=gitlab.chinaunicom.cn/ChinaUnicomBigData/BlockChain/Components/common-backend/version
GO_LDFLAGS = -X $(VERSIONPKG).Main=$(VERSION) -X $(VERSIONPKG).ChangeLog=$(shell git rev-list -1 HEAD) -X $(VERSIONPKG).BuiltAt=$(shell date +%Y-%m-%d\.%H:%M:%S)

PROJECT_FILES = $(shell find . -name "*.go" -or -name "*.h" -or -name "*.c" -or -name "*.s")

PROJECT_PKG=gitlab.chinaunicom.cn/ChinaUnicomBigData/BlockChain/Components/common-backend

pkg-map.common-backend := gitlab.chinaunicom.cn/ChinaUnicomBigData/BlockChain/Components/common-backend

#binary
common-backend: build/bin/common-backend

# docker
docker-common-backend: build/image/backend

.PHONY: clean-all
clean-all:
	rm -rf build/bin/*
	rm -rf build/docker/bin/*
	rm -rf build/image/backend/payload/*

build/bin/%: $(PROJECT_FILES)
	@echo "Building ${@F} in build/bin directory ..."
	mkdir -p build/bin && go build -o build/bin/${@F}  -ldflags "$(GO_LDFLAGS)" $(path-map.$(@F))
	@echo "Built build/bin/${@F}"

build/image/%: Makefile build/image/%/payload build/image/%/docker-entrypoint.sh build/image/%/Dockerfile
	$(eval TARGET = ${patsubst build/image/%,%,${@}})
	@echo "Building docker $(TARGET)-image"
	docker build -t chinaunicom/common-$(TARGET):$(VERSION) $(@)

#new version
build/docker/bin/%: $(PROJECT_FILES)
	@echo "Building ${@F} with ldflags "$(GO_LDFLAGS)" in build/docker/bin directory ..."
	@mkdir -p build/docker/bin
	mkdir -p build/docker/bin && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags backend -o build/docker/bin/${@F}  -ldflags "$(GO_LDFLAGS)" $(path-map.$(@F))


build/image/backend/payload: \
	build/docker/bin/common-backend \
	conf/common-backend.yaml
		@echo "Copying $^ to $@"
		mkdir -p $@
		cp $^ $@