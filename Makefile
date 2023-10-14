.PHONY: frontend

build: frontend backend

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