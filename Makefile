NAME=github-project-exporter
VERSION=0.1.0
COMMIT=$(shell git rev-parse --verify HEAD)

PACKAGES=$(shell go list ./...)
BUILD_FLAGS=-ldflags "-X main.VERSION=$(VERSION) -X main.COMMIT=$(COMMIT)"

.PHONY: build test container clean

build: vendor
	go build $(BUILD_FLAGS) .

test: vendor
	go test -v $(PACKAGES)
	go vet $(PACKAGES)

container:
	GOARCH=amd64 GOOS=linux go build $(BUILD_FLAGS) .
	docker build -t summerwind/$(NAME):latest -t summerwind/$(NAME):$(VERSION) .
	rm -rf $(NAME)

clean:
	rm -rf $(NAME)
	rm -rf dist

dist:
	mkdir -p dist
	
	GOARCH=amd64 GOOS=darwin go build $(BUILD_FLAGS) .
	tar -czf dist/$(NAME)_darwin_amd64.tar.gz $(NAME)
	rm -rf $(NAME)
	
	GOARCH=amd64 GOOS=linux go build $(BUILD_FLAGS) .
	tar -czf dist/$(NAME)_linux_amd64.tar.gz $(NAME)
	rm -rf $(NAME)
	
	GOARCH=arm64 GOOS=linux go build $(BUILD_FLAGS) .
	tar -czf dist/${NAME}_linux_arm64.tar.gz $(NAME)
	rm -rf $(NAME)
	
	GOARCH=arm GOOS=linux go build $(BUILD_FLAGS) .
	tar -czf dist/${NAME}_linux_arm.tar.gz $(NAME)
	rm -rf $(NAME)

vendor:
	glide install
