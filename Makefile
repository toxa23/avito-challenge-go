default: build
	@echo Done, build completed.

build: *.go
	go build

run: build
	godotenv go run .

./build/docker-image: build
	docker build -t .
	@mkdir -p ./build && touch $@

docker-image: ./build/docker-image
	@echo "Docker image created"

build-test:

test: build-test build
	CGO_ENABLED=0 go test -v ./... -cover
