.PHONY: init/build
init/build:
	CGO_ENABLED=0 go build -o shared-init cmd/shared-init/main.go

# build 
.PHONY: init/build-image
init/build-image: init/build
	docker build -t $(image) . 

release:
	CGO_ENABLED=0 go build -o shared-init cmd/shared-init/main.go