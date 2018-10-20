
build-go:
	go build -o drone-docker-image-promote

build-docker:
	docker build -t mrupgrade/drone-docker-image-promote .

build: build-go build-docker


test-pipeline: build
	drone exec .test-drone.yaml