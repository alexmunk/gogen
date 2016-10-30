GITHUB_OAUTH_CLIENT_ID = 39c483e563cd5cedf7c1
GITHUB_OAUTH_CLIENT_SECRET = 024b16270452504c35f541aca4bf78781cd06db9

all: install

build:
	go build -ldflags "-X github.com/coccyx/gogen/github.gitHubClientID=$GITHUB_OAUTH_CLIENT_ID -X github.com/coccyx/gogen/github.gitHubClientSecret=$GITHUB_OAUTH_CLIENT_SECRET"

install:
	go install -ldflags "-X github.com/coccyx/gogen/share.gitHubClientID=$(GITHUB_OAUTH_CLIENT_ID) -X github.com/coccyx/gogen/share.gitHubClientSecret=$(GITHUB_OAUTH_CLIENT_SECRET)"

test:
	go test -v ./...