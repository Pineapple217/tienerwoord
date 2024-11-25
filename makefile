DOCKER_TAG ?= latest

docker-build:
	docker buildx build --platform linux/amd64,linux/arm64 -t pineapple217/woord-vote:$(DOCKER_TAG) . 

docker-update:
	docker buildx build --platform linux/amd64,linux/arm64 -t pineapple217/woord-vote:$(DOCKER_TAG) --push . 

build:
	go build -o ./tmp/main.exe ./main.go

start:
	@./tmp/main.exe
	
make run:
	@make --no-print-directory build
	@make --no-print-directory start