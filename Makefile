.PHONY: frontend

build: frontend backend

docker-run:
	docker-compose up

docker-run-build:
	docker-compose up --build
run: frontend
	go run cmd/server/main.go

frontend:
	npm run build --prefix frontend/

backend:
	go build -o out/web .

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