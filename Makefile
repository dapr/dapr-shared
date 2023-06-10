.PHONY: init/build
init/build:
	CGO_ENABLED=0 go build -o ambient-init cmd/ambient-init/main.go

# build 
.PHONY: init/build-image
init/build-image: init/build
	docker build -t $(image) . 