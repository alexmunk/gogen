GITHUB_OAUTH_CLIENT_ID = 39c483e563cd5cedf7c1
GITHUB_OAUTH_CLIENT_SECRET = 024b16270452504c35f541aca4bf78781cd06db9

ifeq ($(OS),Windows_NT)
	dockercmd := docker run -e TERM --rm -it -v $(CURDIR):/go/src/github.com/coccyx/gogen clintsharp/gogen bash
else
	cd := $(shell pwd)
	dockercmd := docker run --rm -it -v $(cd):/go/src/github.com/coccyx/gogen clintsharp/gogen bash
endif

all: install

build:
	go build -ldflags "-X github.com/coccyx/gogen/github.gitHubClientID=$GITHUB_OAUTH_CLIENT_ID -X github.com/coccyx/gogen/github.gitHubClientSecret=$GITHUB_OAUTH_CLIENT_SECRET"

install:
	go install -ldflags "-X github.com/coccyx/gogen/share.gitHubClientID=$(GITHUB_OAUTH_CLIENT_ID) -X github.com/coccyx/gogen/share.gitHubClientSecret=$(GITHUB_OAUTH_CLIENT_SECRET)"

test:
	go test -v ./...

docker:
	$(dockercmd)
