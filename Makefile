setup:
	go mod tidy

clean:
	rm -rf ./bin

build: clean setup
	CGO_ENABLED=0 go build -v -o ./bin/main ./cmd/main.go

image:
	docker build -t suricat/rinha-3-solucao1 .

docker-push:
	docker buildx build --push --platform linux/amd64 --tag suricat/rinha-3-solucao1:latest .

dev:
	docker compose -f ./docker-compose.yml up redis

image-cleanup:
	@for i in $$(docker ps --filter "name=rinha-publisher-consumer-1" --filter "name=rinha-publisher-consumer-2" --format "{{.ID}}"); do docker rm -f $$i; done
	@for i in $$(docker image ls --filter "reference=rinha-publisher-consumer-1" --filter "reference=rinha-publisher-consumer-2" --format "{{.ID}}"); do docker image rm -f $$i; done

run: image-cleanup
	docker compose up --build

stats:
	docker container stats rinha-nginx rinha-redis rinha-publisher-consumer-1 rinha-publisher-consumer-2
