dep: 
	go mod tidy

run: 
	go run main.go

build: 
	go build -o main main.go

run-build: build
	./main

up:
	docker-compose up -d

down:
	docker-compose down

test:
	go test -v ./...