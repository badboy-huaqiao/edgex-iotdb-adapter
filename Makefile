.PHONY: build test unittest lint clean prepare update docker

GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=GCO_ENABLED=1 GO111MODULE=on go

# see https://shibumi.dev/posts/hardening-executables
CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2"
CGO_CFLAGS="-O2 -pipe -fno-plt"
CGO_CXXFLAGS="-O2 -pipe -fno-plt"
CGO_LDFLAGS="-Wl,-O1,–sort-common,–as-needed,-z,relro,-z,now"

# Don't need CGO_ENABLED=1 on Windows w/o ZMQ.
# If it is enabled something is invoking gcc and causing errors
ifeq ($(OS),Windows_NT)
  GO=CGO_ENABLED=0 GO111MODULE=on go
endif

MICROSERVICES=cmd/adapter-server

.PHONY: $(MICROSERVICES)

DOCKERS=docker_edgex_iotdb_adapter
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.1.0)
GIT_SHA=$(shell git rev-parse HEAD)
GOFLAGS=-ldflags "-X  github.com/edgexfoundry/edgex-iotdb-adapter.Version=$(VERSION)" -trimpath -mod=readonly
CGOFLAGS=-ldflags "-linkmode=external -X  github.com/edgexfoundry/edgex-iotdb-adapter.Version=$(VERSION)" -trimpath -mod=readonly -buildmode=pie

tidy:
	go mod tidy

build: $(MICROSERVICES)

cmd/adapter-server:
	$(GOCGO) build $(CGOFLAGS) -o $@ ./cmd

test:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) vet ./...
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]
	./bin/test-attribution-txt.sh

clean:
	rm -f $(MICROSERVICES)

docker: $(DOCKERS)

docker_edgex_iotdb_adapter:
	docker build \
		--label "git_sha=$(VERSION)" \
		-t vmware/edgex-iotdb-adapter:$(VERSION) \
		-t vmware/edgex-iotdb-adapter:$(VERSION)-dev \
		.

vendor:
	CGO_ENABLED=0 GO111MODULE=on go mod vendor
