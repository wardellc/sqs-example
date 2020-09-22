.PHONY: build deploy

build:
	go get ./...
	sam build

deploy: build
	sam deploy
