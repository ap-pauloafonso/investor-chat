.PHONY: proto

docker-run:
	docker-compose up --build

docker-start:
	docker-compose up -d

docker-stop:
	docker-compose down

install-front:
	npm install --prefix frontend/

lint:
	golangci-lint run ./...

test:
	go test ./... -count=1 --cover


proto:
	rm -f pb/*.go
	protoc --proto_path=proto \
      --go_out=pb --go_opt=paths=source_relative \
      --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
      proto/*.proto
