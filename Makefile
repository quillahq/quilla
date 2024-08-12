JOBDATE		?= $(shell date -u +%Y-%m-%dT%H%M%SZ)
GIT_REVISION	= $(shell git rev-parse --short HEAD)
VERSION		?= $(shell git describe --tags --abbrev=0)

LDFLAGS		+= -linkmode external -extldflags -static
LDFLAGS		+= -X github.com/quilla-hq/quilla/version.Version=$(VERSION)
LDFLAGS		+= -X github.com/quilla-hq/quilla/version.Revision=$(GIT_REVISION)
LDFLAGS		+= -X github.com/quilla-hq/quilla/version.BuildDate=$(JOBDATE)

ARMFLAGS		+= -a -v
ARMFLAGS		+= -X github.com/quilla-hq/quilla/version.Version=$(VERSION)
ARMFLAGS		+= -X github.com/quilla-hq/quilla/version.Revision=$(GIT_REVISION)
ARMFLAGS		+= -X github.com/quilla-hq/quilla/version.BuildDate=$(JOBDATE)

.PHONY: release

fetch-certs:
	curl --remote-name --time-cond cacert.pem https://curl.haxx.se/ca/cacert.pem
	cp cacert.pem ca-certificates.crt

compress:
	upx --brute cmd/quilla/release/quilla-linux-arm
	upx --brute cmd/quilla/release/quilla-linux-aarch64

build-binaries:
	go get github.com/mitchellh/gox
	@echo "++ Building quilla binaries"
	cd cmd/quilla && CC=arm-linux-gnueabi-gcc gox -verbose -output="release/{{.Dir}}-{{.OS}}-{{.Arch}}" \
		-ldflags "$(LDFLAGS)" -osarch="linux/arm"

build-arm:
	cd cmd/quilla && env CC=arm-linux-gnueabihf-gcc CGO_ENABLED=1 GOARCH=arm GOOS=linux go build -ldflags="$(ARMFLAGS)" -o release/quilla-linux-arm
	# disabling for now 64bit builds
	# cd cmd/quilla && env GOARCH=arm64 GOOS=linux go build -ldflags="$(ARMFLAGS)" -o release/quilla-linux-aarc64

armhf-latest:
	docker build -t quillahq/quilla-arm:latest -f Dockerfile.armhf .
	docker push quillahq/quilla-arm:latest

aarch64-latest:
	docker build -t quillahq/quilla-aarch64:latest -f Dockerfile.aarch64 .
	docker push quillahq/quilla-aarch64:latest

armhf:
	docker build -t quillahq/quilla-arm:$(VERSION) -f Dockerfile.armhf .
	# docker push quillahq/quilla-arm:$(VERSION)

aarch64:
	docker build -t quillahq/quilla-aarch64:$(VERSION) -f Dockerfile.aarch64 .
	docker push quillahq/quilla-aarch64:$(VERSION)

arm: build-arm fetch-certs armhf aarch64

test:
	go install github.com/mfridman/tparse@latest
	go test -json -v `go list ./... | egrep -v /tests` -cover | tparse -all -smallscreen

build:
	@echo "++ Building quilla"
	GOOS=linux cd cmd/quilla && go build -a -tags netgo -ldflags "$(LDFLAGS) -w -s" -o quilla .

install:
	@echo "++ Installing quilla"
	# CGO_ENABLED=0 GOOS=linux go install -ldflags "$(LDFLAGS)" github.com/quilla-hq/quilla/cmd/quilla	
	GOOS=linux go install -ldflags "$(LDFLAGS)" github.com/quilla-hq/quilla/cmd/quilla	

image:
	docker build -t quillahq/quilla:alpha -f Dockerfile .

image-debian:
	docker build -t quillahq/quilla:alpha -f Dockerfile.debian .

alpha: image
	@echo "++ Pushing quilla alpha"
	docker push quillahq/quilla:alpha

e2e: install
	cd tests && go test

run:
	go install github.com/quilla-hq/quilla/cmd/quilla
	quilla --no-incluster --ui-dir ui/dist

lint-ui:
	cd ui && yarn 
	yarn run lint --no-fix && yarn run build

run-ui:
	cd ui && yarn run serve

build-ui:
	docker build -t quillahq/quilla:ui -f Dockerfile .
	docker push quillahq/quilla:ui

run-debug: install
	DEBUG=true quilla --no-incluster